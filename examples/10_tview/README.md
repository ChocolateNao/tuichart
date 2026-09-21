# 10 — tview

A live tuichart chart embedded in a tview (tcell-based) application.

## What it shows

Each canvas cell is painted directly onto the tcell screen via `Canvas.CellAt`,
framed by the same title/footer chrome used by the other integration examples.
`board.RenderLayout` returns each diagram's rectangle, which is drawn as a
small geometry chip in the gap row above it.

## How to run

```sh
cd examples/10_tview
go run .
```

## Caveats

- **Separate module**: run from inside `examples/10_tview/`.
- Requires a real terminal. The tview library takes full control of the screen;
  Ctrl+C exits cleanly.
