# Live rendering

The `Live` renderer drives a `Chart` as a real-time, continuously updated
display: it paints diffs at your chosen refresh rate, redraws the entire
frame when the terminal is resized, swallows stray input bytes so that
background typing doesn't break the display, and always restores the
terminal when it exits.

## Quick sketch

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	g := tuichart.New(tuichart.WithWidth(60))
	p := tuichart.NewPlot().Title("requests/s")
	g.Add(p)

	l := tuichart.NewLive(g,
		tuichart.WithInterval(500*time.Millisecond),
		tuichart.WithLiveOutput(os.Stdout),
		tuichart.OnUpdate(func() {
			// pull fresh data into p, called under the render lock
			val := math.Sin(float64(time.Now().UnixMilli())/500) * 100 + 150
			line.SetValues(append(line.Values(), val))
		}),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := l.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
```

Run it (needs a real TTY):

```bash
go run ./examples/06_live
```

## How painting works

A `Live` instance holds:

- **an event channel** (`chan liveEvent`) that receives ticks, resize
  notifications, and flush requests.
- **a render lock** (`sync.Mutex`) under which both data mutation and frame
  painting happen, so you never see a half-updated screen.

`Run` enters a loop that:

1. Waits on the event channel for the next tick, resize, or flush.
2. If the event is a *tick* (time interval elapsed), calls `OnUpdate` under
   the lock so your pump can update the diagram's data.
3. Paints the updated chart frame to the terminal using *cell-level diffing*:
   only changed cells are emitted; identical frames produce no output.
4. On resize (`SIGWINCH` on Unix, console polling on Windows), clears the
   screen and repaints the full frame immediately.
5. When the context is cancelled, the loop stops and the terminal is
   restored.

## Static frames without a TTY

`Live.Frame(width)` renders one frame at the given width *without* touching
the screen or invoking `OnUpdate`. It returns a plain string. This is useful
for testing, for previewing what a live display would look like, and for
embedding the live chart in another renderer.

```
┌─ requests ───────────────────────────────────────┐
│                                                  │
│   │  ·        ⢀⣀⡀         ·         ·    ─── rps │
│ 30┤········⢀⠔⠊⠁·⠈⠑⢄····························· │
│   │  ·    ⡔⠁   ·   ⠣⡀     ·         ·         ·  │
│   │  ·  ⢀⠎     ·    ⠱⡀    ·         ·         ·  │
│ 25┤····⢀⠎············⠑⡄························· │
│   │  ·⢀⠎       ·      ⠘⡄  ·         ·       ⢀ ·  │
│   │  ⢀⠎        ·       ⠘⡄ ·         ·      ⢠⠃ ·  │
│ 20┤··⠈··················⠸⡀················⢠⠃···· │
│   │  ·         ·         ⠱⡀         ·    ⢠⠃   ·  │
│   │  ·         ·          ⠱⡀        ·   ⢠⠃    ·  │
│ 15┤························⠑⡄··········⡠⠃······· │
│   │  ·         ·          · ⠈⢆      · ⡰⠁      ·  │
│   │  ·         ·          ·   ⠑⢄⡀   ⡠⠊        ·  │
│ 10┤·····························⠈⠑⠒⠊············ │
│   │  ·         ·          ·         ·         ·  │
│   └──┬─────────┬──────────┬─────────┬─────────┬─ │
│      0        10         20        30        40  │
└──────────────────────────────────────────────────┘
```

## Live options

`tuichart.NewLive(chart, opts...)`:

| Option                     | Effect                                                              |
| -------------------------- | ------------------------------------------------------------------- |
| `WithLiveOutput(w)`        | Redirect output away from `os.Stdout`.                              |
| `WithInterval(d)`          | Minimum interval between repaints (clamped to 10 ms).               |
| `WithFPS(fps)`             | Set the interval from a target frame rate.                          |
| `OnUpdate(fn)`             | Register a callback invoked under the render lock on each tick.     |

## Updating the chart

```go
// Update under the lock, then repaint immediately.
l.Update(func() {
	// mutate diagram data here
})

// Repaint without running OnUpdate.
l.Repaint()
```

`Update` runs its function while holding the lock, then paints a single
frame. `Repaint` paints the current state without changing any data — useful
when the chart's data was mutated externally (e.g. from a goroutine that
already holds its own synchronization).

## Resizing and clipping

When the terminal is resized, the live renderer:

1. Receives a resize notification through its platform-specific watcher
   (`SIGWINCH` on Linux/BSD/macOS, console polling on Windows).
2. Clears the alt-screen region.
3. Repaints the full frame at the new width immediately.

Frames taller than the terminal's visible rows are *clipped* so that
re-scrolling and partial reflow never appear.

## Terminal handling

The `Live` renderer modifies the terminal in these ways:

- **Alt screen**: enters the alternate screen buffer when `Run` starts.
- **Input quieting**: disables echo, canonical mode, and (on Unix) swallows
  stray bytes using `select`/`read`, so background keystrokes don't appear
  as noise.
- **SIGWINCH/resize**: registers a platform-specific watcher; on Unix this
  is a signal; on Windows it is a polling loop.
- **ISIG**: the `ISIG` flag is *kept on* so `Ctrl+C` still produces a
  signal and the context can cancel cleanly.

On exit (`Run` returns or the context is cancelled), all terminal settings
are restored to their original values.

## Customisation and borders

The diagram's own `Draw` method decides whether to paint a border around
itself. `Live` calls `Draw` as usual, passing the same rendering context.
You can use the `Chart.Frame(width)` method if you want the border without
entering the live loop — this is useful for embedding in other renderers
(see [Embedding in other UIs](embedding.md)).