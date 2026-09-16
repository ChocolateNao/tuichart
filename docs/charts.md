# Charts: composition & layout

The `Board` is the container. It owns the shared title and the layout of
whatever diagrams you add to it, and it renders only as wide as you ask —
diagram frames, tick labels, and legends all negotiate the space together.

## The render pipeline

```
Board
 ├─ title row (optional)
 ├─ diagram 1  (full width)
 ├─ diagram 2  (full width)
 ├─ …          (each Add appends a new full-width row)
 └─ diagram n  (full width)
```

Diagrams are placed in **rows**. A row holds one or more diagrams:

- `Add(d)` — one diagram, taking the full width.
- `Row(a, b, c, …)` — several diagrams side by side, sharing the row equally.

Rows render in order, top to bottom, and the tallest diagram in a row makes
the row's height. Diagrams also *suggest* a preferred height through
`HeightHint(width)`, which the board clamps when space is tight.

## Composing mixed diagrams

`Add` and `Row` take any `Drawable`, so you can freely mix types:

```go
g := tuichart.New(tuichart.WithWidth(52))
p := tuichart.NewPlot().Title("latency")
p.Add(tuichart.NewLineVals("p50", []float64{30, 28, 31, 29, 33, 30, 32}))
p.Add(tuichart.NewLineVals("p99", []float64{90, 84, 120, 96, 110, 100, 105}))
g.Add(p)
g.Row(
	tuichart.NewGauge(70, 100).Title("cpu"),
	tuichart.NewGauge(45, 100).Title("mem"),
)
g.Add(tuichart.NewSpark(3, 5, 2, 6, 4, 7, 5, 2, 1).Title("errors/min"))
fmt.Print(g.Render())
```

```
┌─ latency ────────────────────────────────────────┐
│                                                  │
│    │  ·            ⡀            ─── p50  ─── p99 │
│    │  ·          ⢀⠜⠈⠢⡀          ·            ·  │
│    │  ·         ⢠⠊ · ⠈⠢⡀     ⢀⠤⠒⠉⠒⠤⣀        ⣀· │
│ 100┤···········⡰⠁······⠈⠢⡀⣀⠤⠊⠁······⠉⠒⠤⠤⠒⠒⠉⠉··│
│    │  ⢀⡀      ⡔⠁   ·     ⠈      ·            ·  │
│    │  ·⠈⠉⠒⠢⠤⣀⠎     ·            ·            ·  │
│    │  ·            ·            ·            ·   │
│    │  ·            ·            ·            ·   │
│    │  ·            ·            ·            ·   │
│  50┤············································ │
│    │  ·            ·            ·            ·   │
│    │  ·          ⢀⣀⣀⣀⣀⡀    ⢀⣀⣀⡠⠤⠤⠤⣀⣀⣀     ⣀⣀⣀·│
│    │  ⠈⠉⠉⠉⠒⠒⠒⠒⠊⠉⠉⠁ ·  ⠈⠉⠉⠉⠉⠁    ·    ⠉⠉⠉⠉⠉   │
│    │  ·            ·            ·            ·   │
│    └──┬────────────┬────────────┬────────────┬── │
│       0            2            4            6   │
└──────────────────────────────────────────────────┘

┌─ cpu ─────────────────┐  ┌─ mem ─────────────────┐
│█████████████▎░░░░░ 70%│  │████████▌░░░░░░░░░░ 45%│
└───────────────────────┘  └───────────────────────┘

                     errors/min

▃▆▂▇▅█▆▂▁
```

## Board options

`tuichart.New(opts ...Option)`:

| Option                    | Effect                                                                        |
| ------------------------- | ----------------------------------------------------------------------------- |
| `WithWidth(w int)`        | Fixed output width in terminal cells.                                          |
| `WithGap(rows int)`       | Blank rows inserted between diagrams (0 by default).                           |
| `WithDiagramHeight(h int)`| Override all diagrams' preferred height to `h`.                                |
| `WithPalette(cs ...Color)`| Replace the shared color palette (used by multidrawings and legends).          |
| `WithProfile(l Level)`    | Force a color profile: `LevelNone`, `Level16`, `Level256`, `LevelTrue`.        |
| `WithNoColor()`           | Shorthand for `WithProfile(LevelNone)`.                                        |
| `WithColor16()`           | Shorthand for `WithProfile(Level16)`.                                          |
| `WithColor256()`          | Shorthand for `WithProfile(Level256)`.                                         |
| `WithTrueColor()`         | Shorthand for `WithProfile(LevelTrue)`.                                        |
| `WithUnicode(on bool)`    | Force Unicode on/off (default: auto-detected from the locale).                 |

Width is measured in terminal columns. If you omit `WithWidth`, output is
fuller but still deterministic: diagrams expand to fill. When writing to a
file or piping to another program, pick an explicit width.

## Title and layout controls

- `g.Title("service health")` — centered by default.
- `g.TitleAlign(tuichart.AlignLeft | AlignCenter | AlignRight)` — board-level title alignment.
- `g.Add(d)` / `g.Row(ds...)` — lay out diagrams; `g.Row` splits width equally.
- `g.Clear()` / `g.Reset()` — drop everything and start again.

## Rendering

The `*Board` type is both an `io.WriterTo` and can produce strings, buffers,
and canvases:

| Method                      | Returns                                                            |
| --------------------------- | ------------------------------------------------------------------ |
| `Render(width ...int)`      | A string; optional explicit width.                                 |
| `RenderLines(width ...int)` | `[]string`, one cell-row per element (no trailing newlines).        |
| `RenderCanvas(width int)`   | `(*Canvas, Info)` — draw into a raw cell buffer for embedding.      |
| `RenderTo(w io.Writer, width ...int)` | Writes the rendered string to a writer.                   |
| `WriteTo(w io.Writer)`      | `io.WriterTo` implementation.                                      |
| `Reader()`                  | An `io.Reader` wrapping `RenderLines` for streaming.               |

`RenderCanvas` is the hook for embedding tuichart inside another render loop —
see [Embedding in other UIs](embedding.md).

## Nested composition

Rows may contain any `Drawable`, not just built-ins — including your own
[`extending.md`](extending.md) implementations and even
[custom diagrams that wrap other diagrams](extending.md#beyond-the-simple-case).
There is no hard limit on the number of rows; tall boards are clipped only by
the Live renderer's viewport (see [Live rendering](live.md)).