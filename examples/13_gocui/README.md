# 13_gocui — a tuichart board inside a real gocui window

Embeds a tuichart `Board` in a [gocui](https://github.com/awesome-gocui/gocui)
app as an actual window: a framed, titled `View` wraps the chart, and the
board fills the view's interior by writing its ANSI output straight into the
view's content buffer. gocui's `OutputTrue` escape interpreter parses the
per-cell SGR colors, so the chart flows through gocui's own renderer — no
manual `SetRune` painting, no bare cells over the screen.

Focus is always visible: gocui's `Highlight` draws the current window's frame
in an accent color, and the focused footer button fills with that same accent
(the chart window can be focused, clicked and tabbed to). Mouse support
drives everything the keys do: clicks focus the clicked window (and activate
button clicks), and the wheel cycles datasets from anywhere.

## Run

It is a separate nested module (own `go.mod` with
`replace github.com/ChocolateNao/tuichart => ../..`), so run it from inside its
directory:

```sh
cd examples/13_gocui && go run .
```

The tests run against gocui's simulated screen, so no terminal is needed:

```sh
cd examples/13_gocui && go test ./...
```

## Controls

| Input | Action |
| ------- | -------- |
| `Ctrl+C` / `q/Q` | quit |
| `Tab` / `Shift+Tab` | move focus (chart window, then each footer button) |
| `Enter` / `Space` | activate the focused footer button |
| click | focus the clicked window; clicks on buttons activate them |
| mouse wheel | cycle the dataset (requests/s, memory, cpu) |
| `d/D` | cycle the dataset |
| `c/C` | cycle the line color (cyan, yellow, green, red, blue, fuchsia) |
| `u/U` | toggle unicode / ascii rendering |

The letter keys work both lowercase and with Shift (their capitals), since
gocui delivers a `Shift+letter` press as the uppercase rune.

## How it works

- `ui` keeps all mutable state (dataset, color, unicode and a per-selection
  `Board` cache) together; every handler runs on gocui's single goroutine, so
  it needs no locks.
- The layout manager places three `View`s each frame: the framed `chart`
  window (title + border owned by gocui, interior painted by the board), a
  `status` strip styled like the other integration examples' `───` footer,
  and one framed `View` per footer button.
- `bool switch` is the app's single source of truth: gocui frames (`gg.ASCII`),
  the board's glyph choice, and the status text all derive from `ui.uni`.
- The board is cached per `(dataset, color, unicode)` selection so toggles
  repaint instantly from `Board.RenderCanvas` on the next layout pass.

See `docs/embedding.md` / `INTEGRATION.md` for the embedding patterns.
