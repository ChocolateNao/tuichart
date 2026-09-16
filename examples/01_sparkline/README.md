# 01 — Sparkline

A minimal one-liner sparkline.

## What it shows

`tuichart.Spark(values)` renders a tiny sine-wave sparkline from raw
float64 values. Colors are auto-detected from the terminal; set
`NO_COLOR=1` to suppress them.

## How to run

```sh
go run ./examples/01_sparkline
```

## Caveats

No board or `New()` call needed, as `Spark` is a standalone convenience
that returns a rendered string.
