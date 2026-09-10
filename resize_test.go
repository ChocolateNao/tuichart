//go:build unix

package tuichart

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

// TestWatchResizeSIGWINCH verifies the Unix resize watcher fires onResize
// when the terminal sends SIGWINCH.
func TestWatchResizeSIGWINCH(t *testing.T) {
	if !platformIsTTY(os.Stdout) {
		t.Skip("requires a TTY")
	}
	var mu sync.Mutex
	got := 0
	stop := watchResize(os.Stdout, func() {
		mu.Lock()
		got++
		mu.Unlock()
	})
	defer stop()

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGWINCH); err != nil {
		t.Fatalf("kill: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		n := got
		mu.Unlock()
		if n > 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("resize watcher never fired after SIGWINCH")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestResizeEventReachesLoop verifies a resize event funnels through the
// central channel and triggers an immediate repaint, even with an effectively
// inactive ticker. The counting writer observes the repaint write call; an
// unchanged frame writes an empty payload, which countingWriter still counts.
func TestResizeEventReachesLoop(t *testing.T) {
	g := New(WithWidth(30), WithNoColor(), WithUnicode(true))
	g.Add(NewPlot().Title("resize-me"))

	l := NewLive(g, WithLiveOutput(&countingWriter{}), WithInterval(time.Hour))
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = l.Run(ctx) }()
	defer func() { cancel(); <-l.Done() }()

	// The initial repaint is one write; the resize event must add another.
	initial := l.out.(*countingWriter).len()
	deadline := time.Now().Add(2 * time.Second)
	for l.out.(*countingWriter).len() <= initial && time.Now().Before(deadline) {
		l.enqueueResize()
		time.Sleep(10 * time.Millisecond)
	}
	if l.out.(*countingWriter).len() <= initial {
		t.Fatal("resize event never reached the render loop")
	}
}

// countingWriter counts Write calls (including empty payloads) so tests can
// observe repaints that produce no visible bytes.
type countingWriter struct {
	mu sync.Mutex
	n  int
}

func (c *countingWriter) Write(p []byte) (int, error) {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
	return len(p), nil
}

func (c *countingWriter) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}
