# Diagram reference

Every built-in diagram implements the `Drawable` interface:

```go
type Drawable interface {
	Draw(rc *Ctx, cv *Canvas)
	HeightHint(width int) int
}
```

`Draw` receives a rendering context `rc` and a fresh, empty canvas `cv`. The
canvas is as wide as the chart allows and exactly `HeightHint` rows tall. Use
`cv.Set` and `cv.Text` to paint.

Below is one example per diagram type. Code and output were captured
deterministically (see [Getting started](quickstart.md)).

---

## Plot — line charts

Braille-mode scatter/line plots, one or more series, optional legends and
tick marks.

```go
g := fixed(52)
p := tuichart.NewPlot()
v := make([]float64, 60)
for i := range v {
	v[i] = math.Sin(float64(i) / 9)
}
p.Add(tuichart.NewLineVals("sin", v))
g.Add(p)
```

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

Key methods on `*Plot`:

| Method                             | Effect                                     |
| ---------------------------------- | ------------------------------------------ |
| `Title(t)`                         | Frame title.                               |
| `Add(Line...)`                     | Add one or more lines.                     |
| `Height(rows)`                     | Fixed height in rows.                      |
| `Grid(on)` / `Legend(on)`         | Axis grid and legend display.              |
| `LogX(on)` / `LogY(on)`           | Logarithmic axis.                          |
| `SetYRange(lo, hi)`               | Pin the Y axis (non-fluent, call before Add). |
| `SetXRange(lo, hi)`               | Pin the X axis.                             |
| `SetXFormatter(f)` / `SetYFormatter(f)` | Custom tick labels.                  |

Build a line with `NewLineVals(name, []float64)` or `NewLine(name, Point...)`.
Both are fluent: `.Color(...)`, `.Marker(rune)`, `.Dashed(true)`.

---

## Function — plot a formula

`NewFunction` evaluates a `func(float64) float64` over a domain and draws
the result with braille glyphs.

```go
g := fixed(52)
f := tuichart.NewFunction(math.Cos).Domain(-math.Pi, math.Pi).Samples(60)
f.SetYRange(-1.2, 1.2)
g.Add(f.Title("y = cos(x)"))
```

```
┌─ y = cos(x) ─────────────────────────────────────┐
│                                                  │
│     │         ·           ·           · ─── f(x) │
│    1┤··················⢀⠤⠒⠒⠒⢄··················· │
│     │         ·      ⢀⠔⠁  ·  ⠑⢄       ·          │
│     │         ·     ⡰⠁    ·    ⠣⡀     ·          │
│  0.5┤··············⡰⠁···········⠘⡄·············· │
│     │         ·   ⡜       ·      ⠱⡀   ·          │
│     │         ·  ⡰⠁       ·       ⠱⡀  ·          │
│    0┤···········⡎··················⠘⡄··········· │
│     │         ·⡜          ·         ⠘⢄·          │
│     │        ⢀⠎           ·          ⠈⢆          │
│ -0.5┤·······⢀⠎························⠘⢄········ │
│     │      ⡔⠊ ·           ·           · ⠱⡀       │
│     │  ⢀⣀⠤⠊   ·           ·           ·  ⠈⠑⢄⣀⡀   │
│   -1┤··········································· │
│     │         ·           ·           ·          │
│     └─────────┬───────────┬───────────┬───────── │
│              -2           0           2          │
└──────────────────────────────────────────────────┘
```

Key methods:

| Method              | Effect                                   |
| ------------------- | ---------------------------------------- |
| `Domain(lo, hi)`    | X axis range to sample over.             |
| `Samples(n)`        | Number of samples (>8).                  |
| `SetYRange(lo, hi)` | Pin the Y axis.                          |
| `Name(s)` / `Color(c)` | Legend label and color.              |
| `LogY(on)`          | Logarithmic Y axis.                      |

---

## BarChart — vertical, horizontal, stacked

Categorical bars, either single-series or stacked, with optional value
labels.

```go
g := fixed(52)
g.Add(tuichart.NewBarValues([]string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"},
	[]float64{120, 200, 150, 80, 70, 110, 130}).
	Title("orders").ShowValues(true))
```

```
┌─ orders ─────────────────────────────────────────┐
│                                                  │
│    │       200                               ██  │
│ 200┤······█████································· │
│    │      █████  150                             │
│ 150┤······█████·▇▇▇▇▇····················130···· │
│    │ 120  █████ █████              110  ▇▇▇▇▇    │
│    │█████ █████ █████  80         ▇▇▇▇▇ █████    │
│ 100┤█████·█████·█████·▄▄▄▄▄··70···█████·█████··· │
│    │█████ █████ █████ █████ █████ █████ █████    │
│  50┤█████·█████·█████·█████·█████·█████·█████··· │
│    │█████ █████ █████ █████ █████ █████ █████    │
│   0┤█████·█████·█████·█████·█████·█████·█████··· │
│    │                                             │
│    └───┬─────┬─────┬─────┬─────┬─────┬─────┬──── │
│       mon   tue   wed   thu   fri   sat   sun    │
└──────────────────────────────────────────────────┘
```

### Stacked bars

Add a second series and switch on `Stacked`:

```go
g := fixed(52)
b := tuichart.NewBarValues([]string{"q1", "q2", "q3", "q4"},
	[]float64{12, 15, 9, 14}).Title("north")
b.Add(tuichart.BarSeries{Name: "west", Values: []float64{5, 7, 6, 8}})
b.Stacked(true)
g.Add(b)
```

```
┌─ north ──────────────────────────────────────────┐
│                                                  │
│   │                                 ██   ██ west │
│ 20┤············█████████·············█████████·· │
│   │            █████████             █████████   │
│   │ █████████  ▅▅▅▅▅▅▅▅▅             █████████   │
│ 15┤·█████████··█████████··█████████··▁▁▁▁▁▁▁▁▁·· │
│   │ ▁▁▁▁▁▁▁▁▁  █████████  █████████  █████████   │
│ 10┤·█████████··█████████··▃▃▃▃▃▃▃▃▃··█████████·· │
│   │ █████████  █████████  █████████  █████████   │
│  5┤·█████████··█████████··█████████··█████████·· │
│   │ █████████  █████████  █████████  █████████   │
│  0┤·█████████··█████████··█████████··█████████·· │
│   │                                              │
│   └─────┬──────────┬──────────┬──────────┬────── │
│        q1         q2         q3         q4       │
└──────────────────────────────────────────────────┘
```

### Horizontal bars

```go
g := fixed(60)
b := tuichart.NewBarValues([]string{"orders", "returns", "refunds"},
	[]float64{90, 40, 15}).Title("volume")
b.Horizontal(true)
g.Add(b)
```

```
┌─ volume ─────────────────────────────────────────────────┐
│ orders ██████████████████████████████████████████████ 90 │
│returns ████████████████████▍ 40                          │
│refunds ███████▋ 15                                       │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

Key methods:

| Method                        | Effect                               |
| ----------------------------- | ------------------------------------ |
| `NewBarValues(cats, vals)`    | Create with a single series.         |
| `Add(BarSeries{Name,Values})`| Add additional series (stacked).     |
| `Stacked(true)`              | Stack values instead of grouping.    |
| `Horizontal(true)`           | Bars grow left-to-right.             |
| `ShowValues(true)`           | Draw numeric labels on bars.         |

---

## Pie / Donut chart

A circular layout: `Slice` adds a segment; `Donut` punches a hole.

```go
g := fixed(44)
g.Add(tuichart.NewPie().
	Slice("web", 400).
	Slice("ios", 300).
	Slice("android", 200).
	Slice("cli", 100).
	Donut(true))
```

```
┌──────────────────────────────────────────┐
│                                          │
│            #                             │
│       ooooo######                        │
│     *oooooo########                      │
│    ***ooooo#########                     │
│   *****oooo##########                    │
│  *******       #######                   │
│ ******           ######                  │
│ *****             #####  ## web 40%      │
│ *****             #####  @@ ios 30%      │
│******             ###### ** android 20%  │
│ *****             #####  oo cli 10%      │
│ ***@@             #####                  │
│ @@@@@@           ######                  │
│  @@@@@@@       #######                   │
│   @@@@@@@@@@@@@@#####                    │
│    @@@@@@@@@@@@@@###                     │
│     @@@@@@@@@@@@@@#                      │
│       @@@@@@@@@@@                        │
│            @                             │
└──────────────────────────────────────────┘
```

Key methods:

| Method               | Effect                                          |
| -------------------- | ----------------------------------------------- |
| `Slice(name, val)`   | Add a segment (call several times).             |
| `Donut(true)`        | Hollow center.                                  |
| `ShowValues(v)`      | Show segment labels.                            |
| `SliceColor(c)`      | Set the color for the *next* added slice.       |

---

## Heatmap

A color-mapped grid; the two `Colors` calls set the low and high end of the
gradient.

```go
g := fixed(48)
grid := make([][]float64, 5)
for y := range grid {
	grid[y] = make([]float64, 6)
	for x := range grid[y] {
		grid[y][x] = float64(x*10 + y*4)
	}
}
g.Add(tuichart.NewHeat(grid).Colors(tuichart.Blue, tuichart.BrightRed).Title("latency"))
```

```
┌─ latency ────────────────────────────────────┐
│       .......-------=======+++++++#######    │
│.......:::::::-------+++++++*******#######    │
│.......:::::::=======+++++++#######%%%%%%%    │
│:::::::-------=======*******#######%%%%%%%    │
│:::::::=======+++++++*******%%%%%%%@@@@@@@    │
│  ..:::--===+++**###%%@@ 0..66                │
└──────────────────────────────────────────────┘
```

Key methods:

| Method                | Effect                                    |
| --------------------- | ----------------------------------------- |
| `NewHeat(grid)`       | Build a heatmap from a `[][]float64`.     |
| `Colors(low, high)`   | Color ramp endpoints.                     |
| `ShowValues(v)`       | Print the raw numbers in each cell.       |
| `CellWidth(n)`        | Minimum column width per cell (default 2).|
| `RowLabels/ColLabels` | Axis labels.                              |

---

## Gauge — single-value indicators

```go
g := fixed(46)
g.Add(tuichart.NewGauge(72, 100).Title("cpu").Label("4 core"))
g.Add(tuichart.NewGauge(50, 100).Style(tuichart.GaugeSegments).ShowPercent(false).Color(tuichart.Lime))
g.Add(tuichart.NewGauge(38, 100).Style(tuichart.GaugeArrow).Color(tuichart.Cyan))
```

```
┌─ cpu ──────────────────────────────────────┐
│███████████████████████▊░░░░░░░░░ 72% 4 core│
└────────────────────────────────────────────┘

┌────────────────────────────────────────────┐
│▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱│
└────────────────────────────────────────────┘

┌────────────────────────────────────────────┐
│━━━━━━━━━━━━━━━╸──────────────────────── 38%│
└────────────────────────────────────────────┘
```

Key methods:

| Method                  | Effect                                           |
| ----------------------- | ------------------------------------------------ |
| `NewGauge(value, max)`  | Create a gauge.                                   |
| `Style(GaugeBlocks)`    | One of `GaugeBlocks`, `GaugeASCII`, `GaugeBrackets`, `GaugeArrow`, `GaugeSegments`. |
| `Label(s)`              | Extra text shown after the percentage.            |
| `ShowPercent(v)`        | Toggle the percentage display.                    |
| `Color(c)`              | Color of the filled portion.                      |

---

## Sparkline

Minimal one-row or multi-row status charts.

```go
g := fixed(50)
g.Add(tuichart.NewSpark(5, 4, 9, 8, 6, 7, 5, 3, 4, 6, 8, 7, 5, 2, 1).Title("packet loss"))
g.Add(tuichart.NewSpark(0, 1, 2, 3, 4, 5, 6, 7, 8, 7, 6, 5, 4, 3, 2, 1))
```

```
                    packet loss

▅▄█▇▅▆▅▃▄▅▇▆▅▂▁


▁▂▃▄▅▅▆▇█▇▆▅▅▄▃▂


```

Key methods:

| Method                   | Effect                       |
| ------------------------ | ---------------------------- |
| `NewSpark(vals...)`      | Create with initial data.    |
| `SetValues([]float64)`  | Replace data.                |
| `Title(s)`               | Set a title above the chart. |

---

## Timeline — annotated time axis

```go
g := fixed(52)
now := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
tl := tuichart.NewTimeline().Format("Jun 02")
tl.Event(now, "deploy")
tl.Event(now.Add(6*time.Hour), "canary")
tl.Events(tuichart.TimelineEvent{
	At: now.Add(20 * time.Hour), Label: "spike", Side: tuichart.SideBelow, Detail: "cpu 98%",
})
g.Add(tl)
```

```
┌──────────────────────────────────────────────────┐
│                                                  │
│                                                  │
│ deploy                                           │
│──◆┬───────────◆────────────┬────────────┬─────◆──│
│            canary                           spike│
│                                         cpu 98%  │
│Jun 02      Jun 02       Jun 02       Jun 02      │
└──────────────────────────────────────────────────┘
```

Key methods:

| Method                          | Effect                                             |
| ------------------------------- | -------------------------------------------------- |
| `Event(at, label)`              | Add an event above the line.                       |
| `Events(TimelineEvent{...})`    | Add events with detail text and `SideAbove/Below`. |
| `Format(layout)`                | `time.Format` layout for the axis labels.          |

---

## Gantt — activity bars over time

```go
g := fixed(52)
day := 24 * time.Hour
start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
g.Add(tuichart.NewGantt().Format("Jun 02").
	Bar("design", start, start.Add(day*3)).
	Bar("impl", start.Add(day*2), start.Add(day*6)).
	Bar("qa", start.Add(day*5), start.Add(day*8)))
```

```
┌──────────────────────────────────────────────────┐
│design  ████████████████                          │
│  impl            █████████████████████           │
│    qa                           ████████████████ │
│                                                  │
│             Jun 02     Jun 04     Jun 07         │
└──────────────────────────────────────────────────┘
```

Key methods:

| Method                               | Effect                                     |
| ------------------------------------ | ------------------------------------------ |
| `Bar(name, start, end)`              | Add an activity bar.                       |
| `Duration(name, start, duration)`    | Same, with `time.Duration`.                |
| `Format(layout)`                     | Axis label layout.                         |

---

## Candlestick — OHLC financial bars

```go
g := fixed(52)
base := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
g.Add(tuichart.NewCandlestick().Title("ACME").Format("Jun 02").
	Candle(base, 142.10, 145.80, 141.55, 144.90).
	Candle(base.AddDate(0,0,1), 145.00, 146.20, 142.30, 142.60).
	Candle(base.AddDate(0,0,2), 142.50, 143.90, 141.20, 143.40).
	Candle(base.AddDate(0,0,3), 143.30, 150.10, 143.10, 149.80))
```

```
┌─ ACME ───────────────────────────────────────────┐
│                                                  │
│    │       ·              ·         # up ·% down │
│ 150┤·········································│·· │
│    │       ·              ·              ####### │
│    │       ·              ·              ####### │
│    │       ·              ·              ####### │
│    │       ·       │      ·              ####### │
│    │  │    ·       │      ·              ####### │
│ 145┤#######····%%%%%%%%%·················####### │
│    │#######·   %%%%%%%%%  ·     │        ####### │
│    │#######·   %%%%%%%%%  · #########    ####### │
│    │#######·   %%%%%%%%%  · #########    ·       │
│    │       ·              ·     │        ·       │
│    │       ·              ·              ·       │
│    └───────┬──────────────┬──────────────┬────── │
│         Jun 25         Jun 26         Jun 27     │
└──────────────────────────────────────────────────┘
```

Key methods:

| Method                                       | Effect                                    |
| -------------------------------------------- | ----------------------------------------- |
| `Candle(at, open, high, low, close)`         | Add a single candle.                      |
| `Candles(cs ...Candle)`                      | Add several at once.                      |
| `UpColor(c)` / `DownColor(c)`               | Color for up/down moves.                  |

---

## TimeSeries — timestamps × values

A wrapper around `Plot` with `time.Time` X-axis values:

```go
g := fixed(52)
ts := tuichart.NewTimeSeries().Format("15:04").Title("requests")
base := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)
line := ts.Line("rps")
for i, v := range []float64{120, 90, 200, 260, 150, 180} {
	line.Add(base.Add(time.Duration(i)*time.Hour), v)
}
g.Add(ts)
```

```
┌─ requests ───────────────────────────────────────┐
│                                                  │
│    │ ·         ·          ·  ⢀       ·   ─── rps │
│ 250┤·······················⢀⠔⠉⢆················· │
│    │ ·         ·          ⡠⠊  ⠈⢆     ·           │
│    │ ·         ·        ⡠⠊·    ⠈⢆    ·           │
│    │ ·         ·      ⢀⠜  ·     ⠈⡆   ·           │
│ 200┤·················⡰⠁··········⠘⡄············· │
│    │ ·         ·    ⡰⠁    ·       ⠘⡄ ·      ⢀    │
│    │ ·         ·   ⡰⠁     ·        ⠘⡄·  ⢀⡠⠔⠊⠁    │
│    │ ·         ·  ⡰⠁      ·         ⠘⡤⠔⠊⠁        │
│ 150┤·············⢰⠁····························· │
│    │ ·         ·⢠⠃        ·          ·           │
│    │ ·⠈⠒⠤⡀     ⢠⠃         ·          ·           │
│    │ ·   ⠈⠑⠢⣀ ⢠⠃          ·          ·           │
│ 100┤·········⠉⠃································· │
│    │ ·         ·          ·          ·           │
│    └─┬─────────┬──────────┬──────────┬────────── │
│    11:46     13:10      14:33      15:56         │
└──────────────────────────────────────────────────┘
```

Key methods:

| Method                       | Effect                                    |
| ---------------------------- | ----------------------------------------- |
| `Line(name)`                 | Create a named series (returns `*TimeSeriesLine`). |
| `Format(layout)`             | `time.Format` layout for the axis.        |
| `SetXFormatter(f)`           | Override the axis formatter manually.     |

---

## Histogram — frequency distribution

Bins the input data and draws the frequency of each bin:

```go
g := fixed(52)
data := []float64{1, 2, 2, 3, 3, 3, 4, 4, 5, 5, 5, 5, 6, 6, 7, 7, 8, 8, 8, 8, 8, 9, 9, 10}
g.Add(tuichart.NewHistogram(data).Bins(5).Title("distribution"))
```

```
┌─ distribution ───────────────────────────────────┐
│                                                  │
│  │                                      ██ count │
│  │                            ███████            │
│ 6┤···················███████··███████··········· │
│  │          ▄▄▄▄▄▄▄  ███████  ███████            │
│  │          ███████  ███████  ███████            │
│ 4┤··········███████··███████··███████··········· │
│  │ ▆▆▆▆▆▆▆  ███████  ███████  ███████  ▆▆▆▆▆▆▆   │
│  │ ███████  ███████  ███████  ███████  ███████   │
│ 2┤·███████··███████··███████··███████··███████·· │
│  │ ███████  ███████  ███████  ███████  ███████   │
│ 0┤·███████··███████··███████··███████··███████·· │
│  │                                               │
│  └────┬────────┬────────┬────────┬────────┬───── │
│       1       2.8      4.6      6.4      8.2     │
└──────────────────────────────────────────────────┘
```

Key methods:

| Method               | Effect                                   |
| -------------------- | ---------------------------------------- |
| `Bins(n)`            | Number of bins.                          |
| `Orientation(o)`     | Vertical or horizontal.                  |

---

## Funnel — stacked conversion stages

Centered trapezoids whose width is proportional to each stage's value, with
a legend to the right. `ShowValues` adds the raw value next to each stage.

```go
g := fixed(52)
g.Add(tuichart.NewFunnel().
	Title("conversion funnel").
	Step("visitors", 100).
	Step("signups", 55).
	Step("paid", 22).
	ShowValues(true))
```

```
┌─ conversion funnel ──────────────────────────────┐
│#############################                     │
│  ########################     visitors 100 (100%)│
│    ####################                          │
│       @@@@@@@@@@@@@@@                            │
│         @@@@@@@@@@@              signups 55 (55%)│
│           ******                                 │
│             ***                     paid 22 (22%)│
│                                                  │
└──────────────────────────────────────────────────┘
```

Key methods:

| Method               | Effect                                        |
| -------------------- | --------------------------------------------- |
| `Step(name, val)`    | Add a stage (call several times).             |
| `StepColor(c)`       | Set the color for the most recent stage.      |
| `ShowValues(v)`      | Print each stage's value next to its name.    |

---

## Radar — multivariate comparison

Concentric rings with one spoke per variable; each series is a polygon
whose vertex on any spoke is its value on that variable.

```go
g := fixed(52)
g.Add(tuichart.NewRadar().Title("skill profile").
	Axes("frontend", "backend", "data", "testing", "devops", "ux").
	Series("you", 4, 6, 8, 5, 7, 3).
	Series("team", 5, 6, 6, 6, 5, 5))
```

```
┌─ skill profile ──────────────────────────────────┐
│                     frontend      ── you  ── team│
│                                                  │
│                     ⢀⡠⠔⠊⡏⠒⠤⣀                     │
│                 ⣀⠤⠒⠊⠁   ⡇   ⠉⠑⠢⢄⡀                │
│            ⢀⡠⠔⠒⠉        ⡇       ⠈⠑⠒⠤⣀            │
│        ⣀⠤⠔⠊⠁          ⢀⡠⡧⣀⣀          ⠉⠒⠢⢄⡀       │
│ ux  ⡤⣒⠉           ⣀⠤⠔⠊⢁⣠⡧⣤⣀⣉⣑⠒⠤⠤⣀⡀       backend │
│     ⡇ ⠉⠒⠤⣀⡀  ⢀⡠⠔⠒⠉⣀⡠⢔⠮⠋ ⡇ ⠉⠒⠤⢍⡉⠉⠒⠚⠛⠶⠶⠤⣄⡠⠔⠊⠁ ⡇    │
│     ⡇     ⠈⢱⠪⢅⣀⠤⠒⠉⡠⠒⠁   ⡇     ⠈⠑⠢⢄⣀⠤⠒⠉ ⣷    ⡇    │
│     ⡇      ⢸  ⡇⢉⠖⠭⣀     ⡇    ⢀⡠⠔⠊⠁⡇    ⡇⢇   ⡇    │
│     ⡇      ⢸  ⡧⠃   ⠉⠑⠢⢄⡀⡇⣀⠤⠒⠉⠁    ⡇    ⡇⠘⡄  ⡇    │
│     ⡇      ⢸⡠⠊⡇     ⢀⡠⠔⠊⡏⠒⠤⣀      ⡇    ⡇ ⢱  ⡇    │
│     ⡇     ⢀⢼  ⡇ ⣀⠤⠒⠉⠁   ⡇   ⠉⠑⠢⢄⡀ ⡇    ⡇ ⠈⡆ ⡇    │
│     ⡇   ⢀⠔⠁⢸⡠⠔⠓⠭⣀       ⡇      ⢀⡨⠕⠓⠤⣀  ⡇  ⠸⡀⡇    │
│     ⡇  ⣠⠧⠒⠊⠁⠉⠢⢄⡀ ⠉⠒⠤⢄⡀  ⡇  ⣀⠤⠒⠊⠁    ⢀⡩⠖⠣⢄⡀ ⢣⡇    │
│     ⠧⣒⠉⠉⠒⠒⠢⠤⠤⣀⣀⡈⠒⢄⡀  ⠈⠑⠢⡧⠒⠉     ⢀⡠⠔⠊⠁ ⣀⣀⣀⣈⢵⡮⠇    │
│devops ⠉⠒⠤⣀⡀    ⠈⠉⠉⠚⠒⢦⡤⠤⣀⣇⣀⠤⠤⣤⠤⠒⠛⠓⠒⠉⠉⠉⠉⣀⠤⠔⠊⠁ data │
│           ⠈⠑⠢⢄⣀      ⠈⠑⠤⣇⠤⠒⠉     ⢀⣀⠤⠒⠉           │
│                ⠉⠒⠤⣀     ⡇    ⣀⡠⠔⠊⠁               │
│                    ⠉⠑⠢⢄⡀⡇⣀⠤⠒⠉                    │
│                        ⠈⠉                        │
│                      testing                     │
└──────────────────────────────────────────────────┘
```

Key methods:

| Method                       | Effect                                    |
| ---------------------------- | ----------------------------------------- |
| `Axes(names ...)`            | Name each variable (one spoke each).      |
| `Series(name, vals ...)`     | Add a polygon across the axes.            |
| `SeriesColor(c)`             | Set the color for the most recent series. |
| `Max(m)`                     | Pin the shared scale (0 = autoscale).     |
| `Fill(true)`                 | Shade the interior of each polygon.       |
| `ShowValues(v)`              | Display value labels.                     |

---

## Treemap — hierarchical proportions

Nested rectangles sized by summed values; interior nodes lay out their
children and gain a caption row when there is room.

```go
g := fixed(52)
g.Add(tuichart.NewTreemap().Title("disk use").
	Add(&tuichart.TreemapNode{Name: "web", Children: []*tuichart.TreemapNode{
		{Name: "static", Value: 40},
		{Name: "logs", Value: 30},
	}}).
	Add(&tuichart.TreemapNode{Name: "mobile", Children: []*tuichart.TreemapNode{
		{Name: "ios", Value: 20},
		{Name: "android", Value: 15},
	}}).
	Item("docs", 10))
```

```
┌─ disk use ───────────────────────────────────────┐
│web                           mobile         docs#│
│static########################ios#################│
│##################################################│
│##################################################│
│##############################android#############│
│##################################################│
│##################################################│
│##################################################│
│logs##############################################│
│##################################################│
│##################################################│
│##################################################│
│##################################################│
│##################################################│
│##################################################│
│##################################################│
│##################################################│
│##################################################│
│##################################################│
│##################################################│
│##################################################│
│##################################################│
└──────────────────────────────────────────────────┘
```

Key methods:

| Method                        | Effect                                     |
| ----------------------------- | ------------------------------------------ |
| `Item(name, val)`             | Add a leaf node.                          |
| `Add(*TreemapNode)`           | Attach any node (children allowed).        |
| `ShowValues(v)`               | Include each node's value in its label.    |

---

## ASCII fallback — no Unicode, no color

When `WithUnicode(false)` or `WithNoColor()` is active the entire library
degrades to printable ASCII:

```go
g := tuichart.New(tuichart.WithWidth(52), tuichart.WithNoColor(), tuichart.WithUnicode(false))
p := tuichart.NewPlot().Title("sin")
v := make([]float64, 60)
for i := range v {
	v[i] = math.Sin(float64(i) / 9)
}
p.Add(tuichart.NewLineVals("sin", v))
g.Add(p)
fmt.Print(g.Render())
```

```
+- sin --------------------------------------------+
|                                                  |
|     |  .            .            .       --- sin |
|    1|.........-----|............................ |
|     |  .    -/     --|           .           .   |
|     |  .   -|       .\           .           .   |
|  0.5|...../...........-|........................ |
|     |  . /          .  -|        .           /   |
|     |  .-|          .   \        .          /.   |
|    0|../.................\................./.... |
|     |  .            .     \      .        /  .   |
|     |  .            .      \     .       -|  .   |
| -0.5|.......................\.........../....... |
|     |  .            .        \   .     -|    .   |
|     |  .            .         -| .   -/      .   |
|   -1|..........................-----/........... |
|     |  .            .            .           .   |
|     +--+------------+------------+-----------+-- |
|        0           20           40          60   |
+--------------------------------------------------+
```

All lines, bars, gauge segments, pie and heatmap slices, and box-drawing
borders fall back to distinct ASCII equivalents when the terminal or the
code disables Unicode. What is rendered in a real color-capable terminal:
braille dots become `/`, `\`, `.`, `|`, `-`; block `█` becomes `#`; gauge
braille characters (`▏▎▍▌▋▊▉█`) become ASCII blocks; heatmap cells map to a
text density ramp (`.`, `:`, `-`, `=`, `+`, `*`, `#`, `@`); pie slices
identify by glyphs (`#`, `@`, `*`, `o`). See [Styling & degradation](styling.md)
for the full mapping tables.