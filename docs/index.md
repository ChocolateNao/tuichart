# tuichart documentation

**tuichart** draws charts, gauges, timelines, sparks, and more straight into
your terminal. It is a pure Go, zero-dependency library: everything it needs
is in the standard library, and a single `go get` pulls in no transitive
dependencies.

Every diagram is a *Drawable* — an object that paints itself onto a cell
buffer given a rendering context. You compose diagram types freely with the
`Board` container, render them once to a string, or drive them continuously
with the `Live` renderer.

## Guide

| Topic                                   | What you will learn                                                                 |
| --------------------------------------- | ----------------------------------------------------------------------------------- |
| [Getting started](quickstart.md)        | Install the module, render your first chart, view it as a static string.            |
| [Charts: composition & layout](charts.md) | Stack diagrams vertically, place them side by side, set width/gap/palette, render to a string, a writer, or a canvas. |
| [Diagram reference](diagrams.md)        | Every built-in diagram type with a short code example and its real terminal output. |
| [Styling & degradation](styling.md)     | Colors, styles, profiles (16/256/truecolor), palettes, and how rendering degrades when a terminal can't do color or Unicode. |
| [Live rendering](live.md)               | Real-time updating frames, diff painting, resize handling, and terminal cleanup.    |
| [Embedding in other UIs](embedding.md)  | Reuse tuichart inside a TUI framework or any other render loop (bubbletea, tview, plain HTTP). |
| [Writing your own diagrams](extending.md) | The `Drawable` contract and a full step-by-step walk-through building a custom diagram type from scratch. |

## Capabilities at a glance

- **Line & function plots** — Unicode braille or ASCII-slope rendering: `Plot`, `Function`, `TimeSeries`.
- **Categorical charts** — `BarChart` (vertical, horizontal, stacked), `Histogram`, `PieChart` (pie and donut), `Heatmap`.
- **Shape & hierarchy charts** — `FunnelChart` (conversion stages), `RadarChart` (multivariate comparison), `TreemapChart` (nested proportions).
- **Time-oriented views** — `Timeline` (annotated events), `Gantt` (activity bars), `TimeSeries` (real timestamps).
- **Single-value widgets** — `Gauge` (five styles), `Sparkline`, and a growing family of small status widgets.
- **Composition** — any diagrams mixed, stacked or side-by-side, with a shared title and frame.
- **Degradation** — truecolor → 256 → 16 → mono glyph ramps; Unicode braille → printable ASCII. No garbled terminals.
- **Live mode** — `Live` paints diffs at up to your refresh rate, redraws on `SIGWINCH`, swallows stray input, and always restores your terminal.

## Deterministic outputs in these docs

Terminals vary, so every example block in this guide was captured with the
fixed options `WithWidth(...)`, `WithNoColor()`, and `WithUnicode(true)`
(plus `WithUnicode(false)` where degradation is being demonstrated). This is
the same trick the test suite uses, and it means the output you see here is
byte-for-byte reproducible on your machine:

```go
g := tuichart.New(tuichart.WithWidth(52), tuichart.WithNoColor(), tuichart.WithUnicode(true))
```

In a real terminal the same code renders in color; how color is negotiated is
covered in [Styling & degradation](styling.md).

## Support

- Repository: <https://github.com/ChocolateNao/tuichart>
- Run every example: `cd examples && ls` then `go run ./01_sparkline` … `go run ./08_custom`
- Live demo (needs a real TTY, Ctrl+C to quit): `go run ./examples/06_live`
- API overview doc for TUI embedding: [`INTEGRATION.md`](../INTEGRATION.md)

## Also in this repo

- `examples/` — runnable examples, one per directory (`01_sparkline` …
  `12_file`). `09_bubbletea`, `10_tview`, `11_server`, and `12_file` are
  separate modules: TUI embedding (bubbletea, tview), an HTTP server, and a
  file writer.
- `docs/adr/` — architecture decision records for notable design choices.
- `docs/agents/` — agent-facing documentation for this repository.