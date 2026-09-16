# 05 — Styled

Heatmap, timeline, and candlestick with a custom palette.

## What it shows

A heatmap with explicit row/column labels, a timeline with event detail
sidebars, and a candlestick chart — all sharing a custom palette
(`Cyan, Salmon, Lime, Orange`) passed through `WithPalette`.

## How to run

```sh
go run ./examples/05_styled
```

## Caveats

- `WithGap(1)` inserts blank rows between diagrams for visual separation.
- The candlestick `UpColor`/`DownColor` overrides the palette for up/down
  bars; set them individually to control direction styling.
