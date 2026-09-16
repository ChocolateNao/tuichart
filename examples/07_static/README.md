# 07 — Static

All diagram types rendered once as static output.

## What it shows

Every built-in diagram type in one static board: plot, bar, pie,
histogram, sparkline, gauge, heatmap, timeline, gantt, candlestick,
time series, funnel, radar, and treemap. Output goes to stdout and
terminates immediately — no live loop.

## How to run

```sh
go run ./examples/07_static

# disable colour (pure text)
NO_COLOR=1 go run ./examples/07_static

# ASCII only (no unicode braille/slope glyphs)
LC_ALL=C NO_COLOR=1 go run ./examples/07_static
```

## Caveats

- Output width defaults to the detected terminal width. Pipe to a file
  or pass an explicit width if your terminal is narrower than the
  diagram expects.
- Unicode braille slopes degrade to ASCII `/\^v` glyphs when locale is
  `C` or `NO_COLOR` is set.
