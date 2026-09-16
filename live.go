package tuichart

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	seqEnterAlt   = "\x1b[?1049h"
	seqExitAlt    = "\x1b[?1049l"
	seqHideCursor = "\x1b[?25l"
	seqShowCursor = "\x1b[?25h"
	seqHome       = "\x1b[H"
	seqClearAll   = "\x1b[2J"
	seqClearBelow = "\x1b[J"
)

const defaultInterval = 200 * time.Millisecond

// liveEvent is every input to the render loop: a repaint tick, a resize
// notification, or a data update. All sources funnel through one channel so
// the loop owns the render lock and no two writers race on diagram state.
type liveEvent struct {
	kind liveEventKind
}

type liveEventKind uint8

const (
	evTick   liveEventKind = iota // scheduled repaint
	evResize                      // terminal resized (SIGWINCH / poll)
)

// Live re-renders a Board on a fixed interval, painting each frame over the
// previous one in the terminal's alternate screen buffer. Frames are diffed
// against the last painted frame cell-by-cell: only changed cells reach the
// terminal, and an unchanged frame writes nothing at all. This keeps bytes
// small and eliminates the flicker of full-screen rewrites.
//
// The render loop consumes a single central event channel fed by the ticker,
// terminal resize signals (SIGWINCH on Unix, polling elsewhere), and the
// Update/Repaint entry points. OnUpdate data pumps run under the render lock
// so concurrent producers never race the painter.
//
// While Run is active the terminal is put into a quiet input mode: echo and
// canonical mode are turned off and stray input (typed keys, arrow/scroll
// escape sequences, mouse bytes) is swallowed so nothing ever appears on the
// alternate screen. The terminal is restored exactly on exit, including any
// bytes that arrived in the final instant.
//
// Painting respects the real terminal: a resize forces a full repaint (the
// screen is blanked first, because the emulator reflows its alternate buffer
// and leftover content cannot be trusted), and frames taller than the
// visible area are clipped to it so a small window never scrolls garbage
// onto the display.
//
// All of Live's methods are safe for concurrent use.
type Live struct {
	out      io.Writer
	board    *Board
	update   func()
	events   chan liveEvent
	stop     chan struct{}
	done     chan struct{}
	prev     *Canvas
	termW    int // last detected terminal width (0 = not a TTY)
	termH    int // last detected terminal height (0 = not a TTY)
	interval time.Duration
	stopOnce sync.Once
	mu       sync.Mutex
}

// LiveOption configures a Live renderer.
type LiveOption func(*Live)

// WithInterval sets the time between frames. Values below 10ms are clamped.
func WithInterval(d time.Duration) LiveOption {
	return func(l *Live) { l.interval = d }
}

// WithFPS sets the frame rate; interval becomes 1/fps.
func WithFPS(fps float64) LiveOption {
	return func(l *Live) {
		if fps > 0 {
			l.interval = time.Duration(float64(time.Second) / fps)
		}
	}
}

// WithLiveOutput redirects frames; defaults to os.Stdout. Any io.Writer
// works, which makes tests deterministic (escape sequences land in the
// buffer).
func WithLiveOutput(w io.Writer) LiveOption {
	return func(l *Live) { l.out = w }
}

// OnUpdate registers a data-pump callback invoked under the render lock
// before every frame. Mutate the board's diagrams there. Do not call Live
// methods from inside it.
func OnUpdate(fn func()) LiveOption {
	return func(l *Live) { l.update = fn }
}

// NewLive wraps a board for continuous rendering.
func NewLive(b *Board, opts ...LiveOption) *Live {
	l := &Live{
		board:    b,
		interval: defaultInterval,
		out:      os.Stdout,
		events:   make(chan liveEvent, 8),
	}
	for _, o := range opts {
		o(l)
	}
	if l.interval < 10*time.Millisecond {
		l.interval = 10 * time.Millisecond
	}
	return l
}

// Frame renders one frame at the given width without touching the screen.
// It does not invoke the OnUpdate callback.
func (l *Live) Frame(width int) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.board.Render(width)
}

// Update runs fn while holding the render lock, then repaints immediately.
func (l *Live) Update(fn func()) {
	l.mu.Lock()
	fn()
	s := l.paintLocked(0)
	l.mu.Unlock()
	l.write(s)
}

// Repaint renders and paints a frame right away.
func (l *Live) Repaint() {
	l.mu.Lock()
	s := l.paintLocked(0)
	l.mu.Unlock()
	l.write(s)
}

// repaintEvent schedules an immediate repaint through the central event
// channel when the loop is running; it is used by external flush requests.
func (l *Live) enqueueResize() {
	if l.events != nil {
		select {
		case l.events <- liveEvent{kind: evResize}:
		default:
		}
	}
}

// Run drives the render loop until ctx is canceled or Stop is called. It
// enters the alternate screen, repaints every interval, and restores the
// terminal on exit. Returns nil when stopped via Stop, or ctx.Err().
//
// Every event source — the frame ticker, terminal resize signals, and stop
// requests — is funneled into a single channel consumed by one loop, so
// diagram state is only ever mutated under the render lock. Resize events
// (SIGWINCH on Unix, console polling on Windows) force an immediate full
// repaint at the newly detected size, with the screen blanked first so the
// reflowed alternate buffer can never leave artifacts behind.
func (l *Live) Run(ctx context.Context) error {
	l.mu.Lock()
	if l.stop != nil {
		l.mu.Unlock()
		return errors.New("tuichart: live already running")
	}
	l.stop = make(chan struct{})
	l.stopOnce = sync.Once{}
	l.done = make(chan struct{})
	l.prev = nil
	l.termW, l.termH = 0, 0
	stop := l.stop
	done := l.done
	events := l.events
	l.mu.Unlock()

	// Quiet the terminal for the whole run: echo and canonical mode off,
	// stray input swallowed, all restored on exit. Registered before the
	// alt-screen restore so the terminal state comes back last.
	restoreInput := func() {}
	if si := os.Stdin; si != nil {
		if rest, engaged := acquireInput(si); engaged {
			inputStop := make(chan struct{})
			swallowed := swallowInput(si, inputStop)
			restoreInput = func() {
				close(inputStop)
				<-swallowed
				drainStdin(si)
				rest()
			}
		}
	}
	defer restoreInput()

	defer close(done)
	defer func() {
		l.mu.Lock()
		l.stop = nil
		l.mu.Unlock()
	}()
	defer fmt.Fprint(l.out, seqShowCursor+seqExitAlt)
	fmt.Fprint(l.out, seqEnterAlt+seqHideCursor)

	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	// Ticker goroutine: its single job is to funnel ticks into the central
	// event channel, coalescing when the loop is backed up.
	go func() {
		for {
			select {
			case <-ticker.C:
				select {
				case events <- liveEvent{kind: evTick}:
				default:
				}
			case <-ctx.Done():
				return
			case <-stop:
				return
			}
		}
	}()

	var stopWatch func()
	if f, ok := l.out.(*os.File); ok {
		stopWatch = watchResize(f, l.enqueueResize)
	}
	if stopWatch != nil {
		defer stopWatch()
	}

	l.Repaint()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stop:
			return nil
		case ev := <-events:
			switch ev.kind {
			case evTick:
				l.repaintUpdate()
			case evResize:
				l.repaintFull()
			}
		}
	}
}

// repaintUpdate runs the OnUpdate pump then paints, under the render lock.
func (l *Live) repaintUpdate() {
	l.mu.Lock()
	if l.update != nil {
		l.update()
	}
	s := l.paintLocked(0)
	l.mu.Unlock()
	l.write(s)
}

// repaintFull forces a full repaint of the next frame. Resizes take this
// path: the terminal emulator has reflowed its alternate screen buffer, so
// the previous frame can no longer be diffed cell-by-cell against it.
func (l *Live) repaintFull() {
	l.mu.Lock()
	l.prev = nil
	s := l.paintLocked(0)
	l.mu.Unlock()
	l.write(s)
}

// Stop asks Run to return; safe to call multiple times and from any
// goroutine.
func (l *Live) Stop() {
	l.mu.Lock()
	stop := l.stop
	l.mu.Unlock()
	if stop != nil {
		l.stopOnce.Do(func() { close(stop) })
	}
}

// Done returns a channel closed once Run has exited.
func (l *Live) Done() <-chan struct{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.done
}

// paintLocked renders the next frame, diffs it against the last painted
// frame, stores the new frame, and returns the bytes to write. The caller
// holds mu. If the frame is unchanged the returned string is empty and
// nothing reaches the terminal.
//
// When the output is attached to a real terminal, its size is respected:
// any detected width/height change forces a full repaint (the emulator has
// reflowed the alternate screen and stale content cannot be diffed away),
// and a frame taller than the visible rows is clipped to them so a small
// window never scrolls content off the display.
func (l *Live) paintLocked(width int) string {
	w := width
	var termW, termH int
	if f, ok := l.out.(*os.File); ok && platformIsTTY(f) {
		det := DetectWriter(f)
		if det.W > 0 && det.H > 0 {
			termW, termH = det.W, det.H
		}
	}
	if termW != l.termW || termH != l.termH {
		l.termW, l.termH = termW, termH
		l.prev = nil
	}
	if w <= 0 {
		w = l.board.opts.width
		if w <= 0 {
			if termW > 0 {
				w = termW
			} else {
				w = DetectWriter(l.out).W
			}
		}
	}
	cv, info := l.board.RenderCanvas(w)
	if termH > 0 && cv.h > termH {
		cv = cv.Sub(Rect{W: cv.w, H: termH})
	}
	out := paintFrame(l.prev, cv, info.Level, termH)
	l.prev = cv
	return out
}

func (l *Live) write(s string) {
	io.WriteString(l.out, s)
}
