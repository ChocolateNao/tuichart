# Getting started

## Install

tuichart has no runtime or build dependencies beyond the Go standard library.

```bash
go get github.com/ChocolateNao/tuichart
```

All examples in this guide assume `go >= 1.27` (the module's own `go`
directive). In a real (color-capable) terminal the chart above draws in
color; every output block in this guide was captured with `WithNoColor()` and
`WithUnicode(true)` so it is reproducible on any machine — see
[Styling & degradation](styling.md).

## Your first chart

Create a `main.go`:

```go
package main

import (
	"fmt"
	"math"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	g := tuichart.New(tuichart.WithWidth(52))
	p := tuichart.NewPlot()
	v := make([]float64, 60)
	for i := range v {
		v[i] = math.Sin(float64(i) / 9)
	}
	p.Add(tuichart.NewLineVals("sin", v))
	g.Add(p)
	fmt.Print(g.Render())
}
```

Run it:

```bash
go run main.go
```

You should see a plot of `sin`, rendered with Unicode braille cells:

```
┌──────────────────────────────────────────────────┐
│     │  ·       ⢀⣀⡀  ·            ·       ─── sin │
│    1┤········⡠⠒⠁·⠈⠢⢄···························· │
│     │  ·   ⢀⠜      ⠈⢆            ·           ·   │
│     │  ·  ⢀⠇        ⠈⠢⡀          ·           ·   │
│  0.5┤····⢀⠎···········⢱························· │
│     │  ·⢀⠎          ·  ⠱⡀        ·           ⡀   │
│     │  ·⡜           ·   ⢱        ·          ⡜·   │
│     │  ⠰⠁           ·    ⢇       ·         ⡸ ·   │
│    0┤····················⠈⢆···············⡰⠁···· │
│     │  ·            ·     ⠈⡆     ·       ⡰⠁  ·   │
│     │  ·            ·      ⠸⡀    ·      ⢠⠃   ·   │
│ -0.5┤·······················⠘⡄·········⡠⠃······· │
│     │  ·            ·        ⠘⢄  ·    ⡰⠁     ·   │
│     │  ·            ·          ⠣⡀·  ⢠⠒⠁      ·   │
│   -1┤···························⠈⠉⠒⠉⠁··········· │
│     │  ·            ·            ·           ·   │
│     └──┬────────────┬────────────┬───────────┬── │
│        0           20           40          60   │
└──────────────────────────────────────────────────┘
```

The three moving parts:

1. **`tuichart.New(options...)`** builds a `Board` container (see
   [Composition & layout](charts.md)).
2. **`p := tuichart.NewPlot()`** creates a diagram, which you configure and
   fill with data.
3. **`g.Add(p)`** places the diagram in the board, and **`g.Render()`** turns
   the whole composition into a string you can print, write, or stream.

## What you can build

The next three pages are the practical reference:

- [Charts](charts.md) — stacking, side-by-side rows, sizing, options.
- [Diagrams](diagrams.md) — every diagram type: bar, pie, gauge, sparkline,
  timeline, gantt, candlestick, time series, histogram, heatmap.
- [Styling](styling.md) — colors, profiles, and automatic degradation to
  terminals without color or Unicode.

When you want the diagram to *animate*, see [Live rendering](live.md). When
you want to plug tuichart into another text UI, see
[Embedding in other UIs](embedding.md). When you've outgrown the built-in set,
[Writing your own diagrams](extending.md) builds you a custom one from
scratch.