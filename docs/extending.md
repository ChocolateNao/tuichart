# Writing your own diagrams

tuichart's `Drawable` interface is deliberately tiny:

```go
type Drawable interface {
	Draw(rc *Ctx, cv *Canvas)
	HeightHint(width int) int
}
```

`HeightHint` tells the chart how many rows you would like; `Draw` paints
into whatever canvas the chart gives you. That is the whole contract.

This page walks you through building a complete custom diagram — a **bullet
graph** (Stephen Few's compact single-value indicator) — from a one-line
stub to a fully polished package that integrates with charts, degrades to
ASCII, and plays nicely with the palette.

Every stage below was generated from real code and captured with
`WithNoColor()` and `WithUnicode(true)` so you can reproduce it exactly.

---

## Stage 1 — make it visible

The quickest way to confirm that your `Drawable` works: fill a row with a
block bar.

```go
package bullet

import "github.com/ChocolateNao/tuichart"

// HeightHint tells the chart how many rows we want.
func (b *Bullet) HeightHint(int) int { return 1 }

func (b *Bullet) Draw(rc *tuichart.Ctx, cv *tuichart.Canvas) {
	frac := int(float64(cv.Width()) * b.value / b.max)
	for x := 0; x < frac; x++ {
		cv.Set(x, 0, '█', tuichart.S(tuichart.Lime))
	}
}
```

Output (44 cols, one row, plain text):

```
████████████████
```

The chart allocates a canvas of `HeightHint` rows, and we paint a block bar
into it. Already you can tell the fraction is working.

---

## Stage 2 — frame and tick labels

A bare row is useful, but a real diagram needs a title, a border, and tick
labels so the reader knows what they are looking at. With a fixed height
of 5 rows the tick labels fit *inside* the frame (the frame takes rows 0
and 4, leaving rows 1–3 for content):

```go
func (b *Bullet) Draw(rc *tuichart.Ctx, cv *tuichart.Canvas) {
	uni := rc.Info.Unicode
	barW := cv.Width() - 2 // reserve space for the border
	frac := int(float64(barW) * b.value / b.max)

	cv.Border(tuichart.S(tuichart.Gray), uni)
	cv.Text(2, 0, " p95 latency (ms) ", tuichart.S(tuichart.Default).Bolder())
	for x := 0; x < frac; x++ {
		cv.Set(x+1, 1, '█', tuichart.S(tuichart.Lime))
	}
	cv.Text(1, 3, tuichart.FormatValue(0), tuichart.S(tuichart.Gray))
	cv.Text(barW/2, 3, tuichart.FormatValue(b.max/2), tuichart.S(tuichart.Gray))
	cv.TextRight(barW, 3, tuichart.FormatValue(b.max), tuichart.S(tuichart.Gray))
}
```

```
┌─ p95 latency (ms) ───────────────────────┐
│███████████████                           │
│                                          │
│0                   250                500│
└──────────────────────────────────────────┘
```

The frame row is painted by `cv.Border`, and we overwrite cell (2, 0) with
the title — exactly the trick `chartBase` uses internally.

---

## Stage 3 — qualitative zones and a target marker

A bullet graph's signature is the faint band behind the value bar showing
"qualitative zones" (safe / concern / danger), and a vertical rule marking
the target. We paint zones first (so the value bar covers them), then the
target on top:

```go
func (b *Bullet) Draw(rc *tuichart.Ctx, cv *tuichart.Canvas) {
	uni := rc.Info.Unicode
	barW := cv.Width() - 2
	frac := func(v float64) int { return int(v/b.max*float64(barW) + 0.5) }

	cv.Border(tuichart.S(tuichart.Gray), uni)
	cv.Text(2, 0, " p95 latency (ms) ", tuichart.S(tuichart.Default).Bolder())

	// qualitative zones — faint band spanning the full scale
	for x := 0; x < barW; x++ {
		cv.Set(x+1, 1, '░', tuichart.S(tuichart.DimGray))
	}
	// a danger zone in the top 20%
	for x := frac(b.max * 0.8); x < barW; x++ {
		cv.Set(x+1, 1, '░', tuichart.S(tuichart.Maroon))
	}
	// measured value
	for x := 0; x < frac(b.value); x++ {
		cv.Set(x+1, 1, '█', tuichart.S(tuichart.Lime))
	}
	// target rule
	cv.Set(frac(b.target)+1, 1, '┃', tuichart.S(tuichart.Default).Bolder())

	// ticks
	cv.Text(1, 3, tuichart.FormatValue(0), tuichart.S(tuichart.Gray))
	cv.Text(barW/2, 3, tuichart.FormatValue(b.max/2), tuichart.S(tuichart.Gray))
	cv.TextRight(barW, 3, tuichart.FormatValue(b.max), tuichart.S(tuichart.Gray))
}
```

```
┌─ p95 latency (ms) ───────────────────────┐
│███████████████░░░░░░┃░░░░░░░░░░░░░░░░░░░░│
│                                          │
│0                   250                500│
└──────────────────────────────────────────┘
```

The zones are painted behind the value bar, and the target rule is drawn on
top. In a color terminal the two zone colors are visible; in no-color mode
they collapse to the same glyph.

---

## Stage 4 — fluent setters, ASCII fallback, and a working chart

The final step turns the loose functions into a proper type with fluent
setters, a `Zone` method, a `Title` setter, and a graceful ASCII fallback:

```go
package bullet

import (
	"unicode/utf8"

	"github.com/ChocolateNao/tuichart"
)

type Bullet struct {
	title  string
	name   string
	value  float64
	target float64
	max    float64
	zones  []float64
	color  tuichart.Color
	height int
}

func NewBullet(name string, value, target, max float64) *Bullet {
	return &Bullet{name: name, value: value, target: target, max: max, height: 5}
}

func (b *Bullet) Title(t string) *Bullet   { b.title = t; return b }
func (b *Bullet) Color(c tuichart.Color) *Bullet { b.color = c; return b }
func (b *Bullet) Zone(boundary float64) *Bullet  { b.zones = append(b.zones, boundary); return b }

func (b *Bullet) HeightHint(int) int { return b.height }

func (b *Bullet) Draw(rc *tuichart.Ctx, cv *tuichart.Canvas) {
	uni := rc.Info.Unicode
	zoneChs := []rune{'░', '▒', '▓'}
	valCh, markCh := '█', '┃'
	if !uni {
		zoneChs = []rune{'.', '.', '.'}
		valCh, markCh = '#', '|'
	}
	zoneCh := func(d int) rune {
		if d >= len(zoneChs) { d = len(zoneChs) - 1 }
		if d < 0 { d = 0 }
		return zoneChs[d]
	}

	cv.Border(tuichart.S(tuichart.Gray), uni)
	if b.title != "" {
		cv.Text(2, 0, " "+b.title+" ", tuichart.S(tuichart.Default).Bolder())
	}
	inner := tuichart.Rect{X: 1, Y: 1, W: cv.Width()-2, H: cv.Height()-2}
	barRow := inner.Y
	tickRow := inner.Y + inner.H - 1
	if tickRow <= barRow {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no room)", tuichart.S(tuichart.Gray))
		return
	}

	valStyle := tuichart.S(b.color)
	if b.color.IsZero() {
		valStyle = tuichart.S(rc.Palette[0])
	}

	labelCols := 0
	if b.name != "" {
		labelCols = utf8.RuneCountInString(b.name) + 1
	}
	barX := inner.X + labelCols
	barW := inner.X2() - barX + 1

	frac := func(v float64) int {
		f := v / b.max
		if f < 0 { f = 0 }
		if f > 1 { f = 1 }
		return int(f*float64(barW) + 0.5)
	}

	// qualitative zones
	for x := 0; x < barW; x++ {
		colV := (float64(x) + 0.5) / float64(barW) * b.max
		depth := 0
		for _, bd := range b.zones {
			if colV >= bd { depth++ }
		}
		cv.Set(barX+x, barRow, zoneCh(depth), tuichart.S(tuichart.DimGray))
	}
	// value bar
	for x := 0; x < frac(b.value); x++ {
		cv.Set(barX+x, barRow, valCh, valStyle)
	}
	// target rule
	cv.Set(barX+frac(b.target), barRow, markCh, tuichart.S(tuichart.Default).Bolder())
	if b.name != "" {
		cv.Text(inner.X, barRow, b.name, tuichart.S(tuichart.Default))
	}
	// ticks
	cv.Text(barX, tickRow, tuichart.FormatValue(0), tuichart.S(tuichart.Gray))
	mid := " " + tuichart.FormatValue(b.max/2) + " "
	cv.Text(barX+barW/2-(utf8.RuneCountInString(mid)/2), tickRow, mid, tuichart.S(tuichart.Gray))
	cv.TextRight(inner.X2(), tickRow, tuichart.FormatValue(b.max), tuichart.S(tuichart.Gray))
	cv.TextRight(inner.X2(), barRow, "target "+tuichart.FormatValue(b.target), tuichart.S(tuichart.DimGray))
}
```

Drop it into a chart:

```go
package main

import (
	"fmt"

	"github.com/ChocolateNao/tuichart"
	"github.com/ChocolateNao/tuichart/_example/08_bullet/bullet"
)

func main() {
	g := tuichart.New(tuichart.WithWidth(58), tuichart.WithNoColor(), tuichart.WithUnicode(true))
	g.Title("service health")
	g.Add(bullet.NewBullet("latency", 180, 250, 500).
		Title("p95 latency (ms)").Zone(350).Zone(450))
	g.Row(
		bullet.NewBullet("cpu", 62, 50, 100).Zone(80),
		bullet.NewBullet("orders", 3800, 4500, 6000).Zone(5000),
	)
	fmt.Print(g.Render())
}
```

```
                      service health


┌─ p95 latency (ms) ─────────────────────────────────────┐
│latency █████████████████░░░░░░░┃░░░░░░░░░▒▒▒▒target 250│
│                                                        │
│        0                      250                   500│
└────────────────────────────────────────────────────────┘

┌──────────────────────────┐  ┌──────────────────────────┐
│cpu ███████████┃█target 50│  │orders ████████target 4500│
│                          │  │                          │
│    0         50       100│  │       0      3000    6000│
└──────────────────────────┘  └──────────────────────────┘
```

Because `Name()` is set, each bullet shows its own name; because `Zone` is
called, the faint background shifts through distinct glyphs at each boundary
threshold.

---

## Mixing with built-in diagrams

Your custom type is a `Drawable`, so `Row` works just as it does for any
other diagram:

```go
g := tuichart.New(tuichart.WithWidth(58), tuichart.WithNoColor(), tuichart.WithUnicode(true))
g.Row(
	bullet.NewBullet("cpu", 62, 50, 100).Title("cpu"),
	tuichart.NewGauge(62, 100).Title("gauge"),
)
```

```
┌─ cpu ────────────────────┐  ┌─ gauge ──────────────────┐
│cpu ███████████┃█target 50│  │█████████████▋░░░░░░░░ 62%│
│                          │  └──────────────────────────┘
│    0         50       100│
└──────────────────────────┘
```

The chart treats your diagram identically to a built-in: it asks for
`HeightHint`, allocates a canvas, calls `Draw`, and paints the border.

---

## Degrading gracefully

Because we check `rc.Info.Unicode`, the same `Draw` method degrades to
ASCII when unicode is disabled:

```
+------------------------------------------+
|latency ############.....|......target 250|
|                                          |
|        0               250            500|
+------------------------------------------+
```

`'░'` → `.`, `'█'` → `#`, `'┃'` → `|`, box drawing → ASCII border. No
code changes, no extra branches — just the `uni` check we wrote once.

---

## Where to go next

- See the full, polished version in `_example/08_bullet/` — the tutorial
  code with all setters, error handling, and the `Zone` feature fully
  exercised.
- Read [Charts: composition & layout](charts.md) for the `Row` and `Add`
  rules.
- Read [Styling & degradation](styling.md) for color profiles and the full
  ASCII-fallback mapping tables.
- See [Embedding in other UIs](embedding.md) for how to plug your custom
  diagrams into bubbletea, tview, or a plain HTTP server.