# 04 — Multi

Multiple diagrams stacked together.

## What it shows

Four diagram types composed in one board: a line plot, a bar chart,
and a pie chart. The bar chart is placed in a `Row` alongside the pie
to demonstrate side-by-side layout.

## How to run

```sh
go run ./examples/04_multi
```

## Caveats

Width is fixed at 80 columns to fit most terminals comfortably. The
height hint is calculated from all diagrams; narrower terminals will
clip the bottom of the plot.
