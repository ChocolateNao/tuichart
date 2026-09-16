# 02 — Plot

A static sine-wave line plot.

## What it shows

A `Plot` with a single `Line` rendered via the `Board` container at a
fixed width. `WithNoColor()` pins the output to plain text so the
rendering is deterministic in every terminal.

## How to run

```sh
go run ./examples/02_plot
```

## Caveats

Omit `WithNoColor()` in real use, the plot renders in full colour with
Unicode braille slopes in a colour-capable terminal.
