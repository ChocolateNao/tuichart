package tuichart

import (
	"math"
	"strings"
)

// Braille blocks for SparkBraille style.
var braillePositive = []rune{'⣀', '⣄', '⣤', '⣆', '⣧', '⣷', '⣸', '⣿'}

// TODO: build from top to bottom.
var brailleNegative = []rune{'⢀', '⢸', '⣀', '⣸', '⣠', '⣰', '⣤', '⣿'}

var sparkASCII = []rune{'_', '.', '-', '=', '+', '*', '#', '%', '@'}

// SparkStyle selects the visual language of a Sparkline.
type SparkStyle uint8

const (
	// SparkBlocks draws with eighth-block precision (default).
	SparkBlocks SparkStyle = iota
	// SparkBraille draws with braille patterns for higher vertical resolution.
	SparkBraille
)

// Sparkline renders a compact one-row line of block characters representing
// a sequence of values — useful for inline or embedded mini-charts.
type Sparkline struct {
	vals []float64
	configs
	color       Color
	style       SparkStyle
	gradient    bool
	gradientEnd Color
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

// Style selects the sparkline rendering style.
func (s *Sparkline) Style(st SparkStyle) *Sparkline { s.style = st; return s }

// Gradient enables a color gradient from the start color to the end color.
// The start color is set via Color() (or palette), and the end color via GradientEnd().
func (s *Sparkline) Gradient(on bool) *Sparkline { s.gradient = on; return s }

// GradientEnd sets the ending color for the gradient.
// Use with Gradient(true) to enable.
func (s *Sparkline) GradientEnd(c Color) *Sparkline { s.gradientEnd = c; return s }

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

	// Pass len(s.vals) to maintain 1:1 mapping (backward compatible)
	line := renderSpark(s.vals, len(s.vals), rc.Info.Unicode,
		s.color, s.gradient, s.gradientEnd, s.style, rc)
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

func renderSpark(
	vals []float64,
	width int,
	uni bool,
	c Color,
	gradient bool,
	gradientEnd Color,
	style SparkStyle,
	rc *Ctx,
) []sparkRune {
	n := len(vals)
	out := make([]sparkRune, 0, n)

	if n == 0 {
		return out
	}

	availableWidth := width
	if availableWidth <= 0 {
		availableWidth = n
	}

	scale := float64(n) / float64(availableWidth)

	lo, hi := math.Inf(1), math.Inf(-1)

	for _, v := range vals {
		if math.IsNaN(v) {
			continue
		}

		lo = math.Min(lo, v)
		hi = math.Max(hi, v)
	}

	// Handle all-same values: output space (backward compatible)
	allSame := hi == lo

	col := c
	if col.IsZero() && rc != nil {
		col = SkyBlue
	}

	getColorAt := func(i, total int) Color {
		if !gradient || total <= 1 {
			return col
		}

		start := col
		if start.IsZero() {
			start = rc.Palette[0]
		}

		end := gradientEnd
		if end.IsZero() {
			end = col
		}

		t := float64(i) / float64(total-1)

		return mix(start, end, t)
	}

	for x := 0; x < availableWidth; x++ {
		idx := int(float64(x) * scale)
		if idx >= n {
			idx = n - 1
		}

		v := vals[idx]

		var (
			r   rune
			clr Color
		)

		if math.IsNaN(v) || allSame {
			r = ' '
			clr = col
		} else if v >= 0 {
			t := (v - 0) / (hi - 0)
			if hi <= 0 {
				t = 1
			}

			t = math.Max(0, math.Min(1, t))
			if uni && style == SparkBraille {
				idx := int(t * float64(len(braillePositive)-1))
				r = braillePositive[idx]
			} else {
				blocks := barEighths
				if !uni {
					blocks = sparkASCII
				}

				r = blocks[rampIdx(t, len(blocks))]
			}

			clr = getColorAt(int(t*float64(availableWidth-1)), availableWidth)
		} else {
			t := (v - lo) / (0 - lo)
			if lo >= 0 {
				t = 1
			}

			t = math.Max(0, math.Min(1, t))
			if uni && style == SparkBraille {
				idx := int(t * float64(len(brailleNegative)-1))
				r = brailleNegative[idx]
			} else {
				blocks := barEighths
				if !uni {
					blocks = sparkASCII
				}

				r = blocks[rampIdx(t, len(blocks))]
			}

			clr = getColorAt(int(t*float64(availableWidth-1)), availableWidth)
		}

		out = append(out, sparkRune{r: r, c: clr})
	}

	return out
}

// Spark renders values as a one-line sparkline using the current terminal profile.
func Spark(vals []float64) string {
	info := Detect()
	line := renderSpark(
		vals,
		len(vals),
		info.Unicode,
		SkyBlue,
		false,
		Color{},
		SparkBlocks,
		newCtx(info),
	)

	cv := NewCanvasWithInfo(len(line), 1, info)
	for x, sr := range line {
		cv.Set(x, 0, sr.r, NewStyle(sr.c))
	}

	return strings.TrimSuffix(cv.Render(info.Level), "\n")
}
