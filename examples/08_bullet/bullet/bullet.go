// Package bullet implements a custom diagram type for tuichart (see
// docs/extending.md): a "bullet graph" that shows one measured value against
// a target and the qualitative zones of its scale. It uses only the public
// Canvas/Ctx API — no library internals.
package bullet

import (
	"unicode/utf8"

	"github.com/ChocolateNao/tuichart"
)

// Bullet renders one value against a target and qualitative zones (Stephen
// Few's bullet graph): a faint band shows how the full scale is divided into
// zones, a bold bar shows the measured value, and a marker shows the target.
type Bullet struct {
	title  string
	name   string
	zones  []float64
	value  float64
	target float64
	max    float64
	height int
	color  tuichart.Color
}

// NewBullet creates a Bullet named name measuring value against target on a
// scale of 0..max.
func NewBullet(name string, value, target, max float64) *Bullet {
	return &Bullet{name: name, value: value, target: target, max: max, height: 5}
}

// Title sets the frame title above the diagram.
func (b *Bullet) Title(t string) *Bullet { b.title = t; return b }

// Color sets the value-bar color; zero uses the chart palette.
func (b *Bullet) Color(c tuichart.Color) *Bullet { b.color = c; return b }

// Zone marks a qualitative boundary on the scale (e.g. an "alert" threshold).
// Each boundary adds a visually distinct segment to the background band.
// Call it in ascending order.
func (b *Bullet) Zone(boundary float64) *Bullet { b.zones = append(b.zones, boundary); return b }

// HeightHint is part of the Drawable contract. The chart gives us at most
// this many rows and possibly fewer when space is tight.
func (b *Bullet) HeightHint(int) int { return b.height }

// Draw paints the bullet graph into the canvas area we were given.
func (b *Bullet) Draw(rc *tuichart.Ctx, cv *tuichart.Canvas) {
	uni := rc.Info.Unicode
	zoneChs := []rune{'░', '▒', '▓'}
	valCh, markCh := '█', '┃'
	if !uni {
		zoneChs = []rune{'.', '.', '.'}
		valCh, markCh = '#', '|'
	}
	zoneCh := func(depth int) rune {
		if depth >= len(zoneChs) {
			depth = len(zoneChs) - 1
		}
		if depth < 0 {
			depth = 0
		}
		return zoneChs[depth]
	}

	cv.Border(tuichart.NewStyle(tuichart.Gray), uni)
	if b.title != "" {
		cv.Text(2, 0, " "+b.title+" ", tuichart.NewStyle(tuichart.Default).Bolder())
	}
	inner := tuichart.Rect{X: 1, Y: 1, W: cv.Width() - 2, H: cv.Height() - 2}
	barRow := inner.Y
	tickRow := inner.Y + inner.H - 1
	if tickRow <= barRow {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no room)", tuichart.NewStyle(tuichart.Gray))
		return
	}

	valStyle := tuichart.NewStyle(b.color)
	if b.color.IsZero() {
		valStyle = tuichart.NewStyle(rc.Palette[0])
	}

	labelCols := 0
	if b.name != "" {
		labelCols = utf8.RuneCountInString(b.name) + 1
	}
	barX := inner.X + labelCols
	barW := inner.X2() - barX + 1

	frac := func(v float64) int {
		f := v / b.max
		if f < 0 {
			f = 0
		}
		if f > 1 {
			f = 1
		}
		return int(f*float64(barW) + 0.5)
	}

	// Qualitative zones first, as a faint band spanning the whole scale. Each
	// boundary in b.zones shifts the depth (and so the glyph) of the cells to
	// its right.
	for x := 0; x < barW; x++ {
		colV := (float64(x) + 0.5) / float64(barW) * b.max
		depth := 0
		for _, bd := range b.zones {
			if colV >= bd {
				depth++
			}
		}
		cv.Set(barX+x, barRow, zoneCh(depth), tuichart.NewStyle(tuichart.DimGray))
	}
	// The measured value as a bold block bar (covers the zone glyphs).
	for x := 0; x < frac(b.value); x++ {
		cv.Set(barX+x, barRow, valCh, valStyle)
	}
	// The target as a vertical rule pinned to its column.
	cv.Set(barX+frac(b.target), barRow, markCh, tuichart.NewStyle(tuichart.Default).Bolder())
	if b.name != "" {
		cv.Text(inner.X, barRow, b.name, tuichart.NewStyle(tuichart.Default))
	}

	// Tick row: 0, middle, maximum — plus the target readout on the right.
	cv.Text(barX, tickRow, tuichart.FormatValue(0), tuichart.NewStyle(tuichart.Gray))
	mid := " " + tuichart.FormatValue(b.max/2) + " "
	cv.Text(
		barX+barW/2-(utf8.RuneCountInString(mid)/2),
		tickRow,
		mid,
		tuichart.NewStyle(tuichart.Gray),
	)
	cv.TextRight(inner.X2(), tickRow, tuichart.FormatValue(b.max), tuichart.NewStyle(tuichart.Gray))
	cv.TextRight(
		inner.X2(),
		barRow,
		"target "+tuichart.FormatValue(b.target),
		tuichart.NewStyle(tuichart.DimGray),
	)
}
