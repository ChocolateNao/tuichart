package tuichart

import (
	"math"
	"strings"
)

var sparkASCII = []rune{'_', '.', '-', '=', '+', '*', '#', '%', '@'}

// Sparkline renders a compact one-row line of block characters representing
// a sequence of values — useful for inline or embedded mini-charts.
type Sparkline struct {
	vals []float64
	DisplayConfig
	AxisConfig
	color Color
}

// NewSpark creates a Sparkline with the given initial values.
func NewSpark(vals ...float64) *Sparkline {
	return &Sparkline{AxisConfig: newAxisConfig(), DisplayConfig: newDisplayConfig(), vals: vals}
}

// Title is accepted for symmetry with other diagrams; it renders above the
// sparkline when placed in a Board row by itself only if height allows.
func (s *Sparkline) Title(t string) *Sparkline { s.SetTitle(t); return s }

// Values appends additional data points to the sparkline.
func (s *Sparkline) Values(
	vals ...float64,
) *Sparkline {
	s.vals = append(s.vals, vals...)
	return s
}

// SetValues replaces all data points in the sparkline.
func (s *Sparkline) SetValues(vals []float64) *Sparkline {
	s.vals = vals
	return s
}

// Color sets the sparkline fill color; defaults to SkyBlue.
func (s *Sparkline) Color(c Color) *Sparkline { s.color = c; return s }

// HeightHint returns the suggested height in rows (1, or 2 with a title).
func (s *Sparkline) HeightHint(int) int {
	if s.title != "" {
		return 2
	}

	return 1
}

// WidthHint returns the suggested width for the given height.
func (s *Sparkline) WidthHint(height int) int {
	// Sparklines are typically 1 row high; width is flexible.
	return height * 10
}

// Draw renders the sparkline into the canvas.
func (s *Sparkline) Draw(rc *Ctx, cv *Canvas) {
	row := cv.Height() - 1
	if s.title != "" && cv.Height() >= 2 {
		cv.TextCenter(
			cv.Width()/2,
			0,
			ellipTrunc(s.title, cv.Width(), rc.Info.Unicode),
			NewStyle(Gray),
		)
	}

	line := renderSpark(s.vals, rc.Info.Unicode, s.color, rc)
	for x, r := range line {
		if x >= cv.Width() {
			break
		}

		cv.Set(x, row, r.r, NewStyle(r.c))
	}
}

type sparkRune struct {
	r rune
	c Color
}

func renderSpark(vals []float64, uni bool, c Color, rc *Ctx) []sparkRune {
	n := len(vals)
	out := make([]sparkRune, 0, n)

	blocks := barEighths
	if !uni {
		blocks = sparkASCII
	}

	lo, hi := math.Inf(1), math.Inf(-1)

	for _, v := range vals {
		if math.IsNaN(v) {
			continue
		}

		lo = math.Min(lo, v)
		hi = math.Max(hi, v)
	}

	col := c
	if col.IsZero() && rc != nil {
		col = SkyBlue
	}

	for i := 0; i < n; i++ {
		v := vals[i]
		if math.IsNaN(v) || hi == lo {
			out = append(out, sparkRune{r: ' ', c: col})
			continue
		}

		out = append(out, sparkRune{r: blocks[rampIdx((v-lo)/(hi-lo), len(blocks))], c: col})
	}

	return out
}

// Spark renders values as a one-line sparkline using the current terminal profile.
func Spark(vals []float64) string {
	info := Detect()
	line := renderSpark(vals, info.Unicode, SkyBlue, newCtx(info))

	cv := NewCanvasWithInfo(len(line), 1, info)
	for x, sr := range line {
		cv.Set(x, 0, sr.r, NewStyle(sr.c))
	}

	return strings.TrimSuffix(cv.Render(info.Level), "\n")
}

// Reset restores every configurable property to its default, keeping the
// title.
func (s *Sparkline) Reset() { s.AxisConfig.Reset(); s.DisplayConfig.Reset() }
