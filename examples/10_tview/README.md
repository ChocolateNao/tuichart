# 10 — tview

A live tuichart chart embedded in a tview (tcell-based) application.

## What it shows

Each canvas cell is painted directly onto the tcell screen via
`Canvas.At`, framed by the same title/footer chrome used by the other
integration examples. The tview `Application` loop handles resize and
repaint.

## How to run

```sh
cd examples/10_tview
go run .
```

## Caveats

- **Separate module**: run from inside `examples/10_tview/`.
- Requires a real terminal. The tview library takes full control of the
  screen; Ctrl+C exits cleanly.
