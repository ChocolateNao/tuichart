# 03 — Bar

Grouped bar chart with values shown.

## What it shows

A `BarChart` with two named series ("sales" and "returns") plotted over
five categorical labels, with `ShowValues(true)` printing the raw numbers
inside each bar.

## How to run

```sh
go run ./examples/03_bar
```

## Caveats

With colour disabled (`WithNoColor()`), the two series use distinct ramp
characters (`█` vs `░`) to stay visually separable in mono mode.
