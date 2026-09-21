# 09 — Bubble Tea

A live tuichart chart embedded in a Bubble Tea application.

## What it shows

A live plot, sparkline, and gauges rendered through `board.RenderLayout` into
a string, framed by a fixed title bar and status footer. The footer counts the
diagrams from the returned per-diagram rectangles.

## How to run

```sh
cd examples/09_bubbletea
go run .
```

## Caveats

- **Separate module**: `go.mod` replaces `github.com/ChocolateNao/tuichart` with
  `../..`, so run from inside this directory, not the repo root.
- Requires a real terminal (alt-screen, raw input, SIGWINCH handling).
- The Bubble Tea `View()` method must return a string; the chart is rendered via
  `board.RenderLayout(...)` (canvas → string) each tick.
