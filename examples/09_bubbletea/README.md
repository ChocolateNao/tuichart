# 09 — Bubble Tea

A live tuichart chart embedded in a Bubble Tea application.

## What it shows

A live plot, sparkline, and gauges rendered through `board.RenderTo`
into an in-memory buffer, framed by a fixed title bar and status footer.
The same `io.Writer` path serves files, sockets, and HTTP handlers.

## How to run

```sh
cd examples/09_bubbletea
go run .
```

## Caveats

- **Separate module**: `go.mod` replaces
  `github.com/ChocolateNao/tuichart` with `../..`, so run from inside
  this directory, not the repo root.
- Requires a real terminal (alt-screen, raw input, SIGWINCH handling).
- The Bubble Tea `View()` method must return a string; the chart is
  rendered via `Board.RenderTo(&buf)` each tick.
