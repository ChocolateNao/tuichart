package tuichart

import (
	"fmt"
	"math"
)

// GaugeStyle selects the visual language of a Gauge bar.
type GaugeStyle uint8

const (
	// GaugeBlocks draws ████░░░ with eighth-block precision at the tip.
	GaugeBlocks GaugeStyle = iota
	// GaugeASCII is the plain-terminal variant: ####-----.
	GaugeASCII
	// GaugeBrackets wraps the bar in brackets: [██████░░░░].
	GaugeBrackets
	// GaugeArrow draws ━━━━╸─── with an arrowhead tip.
	GaugeArrow
	// GaugeSegments renders ten discrete segments: ▰▰▰▰▱▱▱▱▱▱.
	GaugeSegments
	// GaugeBraille renders with braille patterns: [ ⣿⣿⣿⣿⣿⣿⣿⣿⣀⣀ ] 75%.
	GaugeBraille
	// GaugeCircles renders with circles: [ ●●●●●●○○○○ ] 60%.
	GaugeCircles
	// GaugeDots renders with dots: [ ••••••••••·········· ] 50%.
	GaugeDots
	// GaugeSquares renders with squares: [ ■■■■■■■■■■□□□□□□□□□□ ] 50%.
	GaugeSquares
)

// Gauge shows a single value against its maximum as a one-row progress
// bar. Works in live dashboards (mutate Value between frames) and embeds
// like any other Drawable.
type Gauge struct {
	label string
	configs
	value            float64
	max              float64
	color            Color
	style            GaugeStyle
	showPct          bool
	gradient         bool
	gradientEnd      Color
	brailleEmptyRune rune
}

// NewGauge creates a Gauge that displays value as a fraction of max.
func NewGauge(value, max float64) *Gauge {
	return &Gauge{
		AxisConfig: newAxisConfig(), DisplayConfig: newDisplayConfig(),
		value:   value,
		max:     max,
		showPct: true,
	}
}

// Title sets the chart title.
func (g *Gauge) Title(s string) *Gauge { g.SetTitle(s); return g }

// Value sets the current value; it is clamped to [0, max].
func (g *Gauge) Value(v float64) *Gauge { g.value = v; return g }

// Max sets the scale maximum (values above are clamped to full).
func (g *Gauge) Max(m float64) *Gauge { g.max = m; return g }

// Style selects one of the GaugeStyle renderings.
func (g *Gauge) Style(s GaugeStyle) *Gauge { g.style = s; return g }

// Label appends a text suffix after the percent readout.
func (g *Gauge) Label(s string) *Gauge { g.label = s; return g }

// Color fixes the fill color; defaults to the first palette entry.
func (g *Gauge) Color(c Color) *Gauge { g.color = c; return g }

// Gradient enables a color gradient from the start color to the end color.
// The start color is set via Color() (or palette), and the end color via GradientEnd().
// If GradientEnd is not set, it defaults to the start color with 50% brightness.
func (g *Gauge) Gradient(on bool) *Gauge { g.gradient = on; return g }

// GradientEnd sets the ending color for the gradient.
// Use with Gradient(true) to enable.
func (g *Gauge) GradientEnd(c Color) *Gauge { g.gradientEnd = c; return g }

// BrailleEmptyRune sets the rune used for empty/unfilled portions in Braille style.
// Default is '⠀' (empty braille). Use '⣀' (upper half block) for a more filled look.
func (g *Gauge) BrailleEmptyRune(r rune) *Gauge { g.brailleEmptyRune = r; return g }

// ShowPercent toggles the numeric percentage suffix (default on).
func (g *Gauge) ShowPercent(on bool) *Gauge { g.showPct = on; return g }

// HeightHint returns the suggested height in rows for the given width.
func (g *Gauge) HeightHint(width int) int {
	if h := g.DisplayConfig.HeightHint(width); h > 0 {
		return h
	}

	if g.title != "" {
		return 3
	}

	return 1
}

// WidthHint returns the suggested width for the given height.
func (g *Gauge) WidthHint(height int) int {
	// Gauge is typically 1 row high; width is flexible.
	return height * 10
}

// Draw renders the gauge bar into the canvas area.
func (g *Gauge) Draw(rc *Ctx, cv *Canvas) {
	inner := frameTitle(cv, rc.Info.Unicode, g.title, g.titleAlign, g.frame)
	if inner.W < 3 || inner.H < 1 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", NewStyle(Gray))
		return
	}

	row := inner.Y + inner.H/2

	st := NewStyle(g.color)
	if g.color.IsZero() {
		st = NewStyle(rc.Palette[0])
	}

	suffix := ""

	pct := 0.0
	if g.max > 0 && !math.IsNaN(g.value) {
		pct = math.Max(0, math.Min(1, g.value/g.max)) * 100
	}

	if g.showPct {
		suffix = fmt.Sprintf(" %.0f%%", pct)
	} else if g.label != "" {
		suffix = " " + g.label
	}

	if g.label != "" && g.showPct {
		suffix += " " + g.label
	}

	suffix = ellipTrunc(suffix, max(inner.W/2-1, 0), rc.Info.Unicode)

	barW := inner.W - runeLen(suffix)
	if barW < 1 {
		return
	}

	g.drawBar(cv, row, inner.X, barW, pct/100, st, rc.Info.Unicode, rc)

	if suffix != "" {
		cv.Text(inner.X+barW, row, suffix, NewStyle(Default))
	}
}

func (g *Gauge) drawBar(cv *Canvas, row, x, w int, frac float64, st Style, uni bool, rc *Ctx) {
	var startColor, endColor Color
	if g.gradient {
		startColor = g.color
		if startColor.IsZero() {
			startColor = rc.Palette[0]
		}

		endColor = g.gradientEnd
		if endColor.IsZero() {
			endColor = startColor
		}
	} else {
		startColor = g.color
		if startColor.IsZero() {
			startColor = rc.Palette[0]
		}

		endColor = startColor
	}

	getColorAt := func(i, total int) Style {
		if total <= 1 || !g.gradient {
			return NewStyle(startColor)
		}

		t := float64(i) / float64(total-1)
		c := mix(startColor, endColor, t)

		return NewStyle(c)
	}

	switch g.style {
	case GaugeASCII:
		gaugeFillRunesGradient(cv, row, x, w, frac, getColorAt, '#', '-')
	case GaugeBrackets:
		if w < 5 {
			if uni {
				gaugeFillEighthsGradient(cv, row, x, w, frac, getColorAt, '█', '░')
			} else {
				gaugeFillRunesGradient(cv, row, x, w, frac, getColorAt, '#', '-')
			}

			return
		}

		br := NewStyle(Default)
		cv.Set(x, row, '[', br)
		cv.Set(x+w-1, row, ']', br)

		if uni {
			gaugeFillEighthsGradient(cv, row, x+1, w-2, frac, getColorAt, '█', '░')
		} else {
			gaugeFillRunesGradient(cv, row, x+1, w-2, frac, getColorAt, '#', '-')
		}
	case GaugeArrow:
		if uni {
			const fill, tip, track = '━', '╸', '─'

			full := int(math.Ceil(frac * float64(w)))
			if frac > 0 && full == 0 {
				full = 1
			}

			for i := 0; i < w; i++ {
				chSt := getColorAt(i, w)
				switch {
				case i < full-1:
					cv.Set(x+i, row, fill, chSt)
				case i == full-1:
					cv.Set(x+i, row, tip, chSt.Bolder())
				default:
					cv.Set(x+i, row, track, NewStyle(DimGray))
				}
			}
		} else {
			gaugeFillRunesGradient(cv, row, x, w, frac, getColorAt, '=', '-')
		}
	case GaugeSegments:
		n := 10
		if w < n {
			n = w
		}

		base := w / n
		rem := w - base*n
		filled := int(math.Round(frac * float64(n)))
		cur := x

		for i := 0; i < n; i++ {
			chunk := base
			if i < rem {
				chunk++
			}

			fullCh, emptyCh := '▰', '▱'
			if !uni {
				fullCh, emptyCh = '#', '-'
			}

			ch := emptyCh
			if i < filled {
				ch = fullCh
			}

			segColor := getColorAt(cur, w)
			for j := 0; j < chunk; j++ {
				cv.Set(cur, row, ch, segColor)
				cur++
			}
		}
	case GaugeBraille:
		emptyBraille := g.brailleEmptyRune
		if emptyBraille == 0 {
			emptyBraille = '⠀'
		}

		gaugeFillBrailleGradient(cv, row, x, w, frac, getColorAt, uni, emptyBraille)
	case GaugeCircles:
		gaugeFillCirclesGradient(cv, row, x, w, frac, getColorAt, uni)
	case GaugeDots:
		gaugeFillDotsGradient(cv, row, x, w, frac, getColorAt, uni)
	case GaugeSquares:
		gaugeFillSquaresGradient(cv, row, x, w, frac, getColorAt, uni)
	default:
		if uni {
			gaugeFillEighthsGradient(cv, row, x, w, frac, getColorAt, '█', '░')
		} else {
			gaugeFillRunesGradient(cv, row, x, w, frac, getColorAt, '#', '-')
		}
	}
}

// gaugeFillRunesGradient fills with a gradient using the provided color function.
func gaugeFillRunesGradient(cv *Canvas, row, x, w int, frac float64,
	getColorAt func(int, int) Style, fill, empty rune,
) {
	full := int(frac * float64(w))
	for i := 0; i < w; i++ {
		ch := empty
		if i < full {
			ch = fill
		}

		cv.Set(x+i, row, ch, getColorAt(i, w))
	}
}

func gaugeFillEighthsGradient(cv *Canvas, row, x, w int, frac float64,
	getColorAt func(int, int) Style, fullRune, empty rune,
) {
	exact := frac * float64(w)
	nFull := int(exact)

	for i := 0; i < w; i++ {
		ch := empty
		if i < nFull {
			ch = fullRune
		} else if i == nFull {
			idx := int((exact - float64(nFull)) * 8)
			if idx > 0 {
				ch = barEighthsH[idx-1]
			}
		}

		cv.Set(x+i, row, ch, getColorAt(i, w))
	}
}

func gaugeFillBrailleGradient(cv *Canvas, row, x, w int, frac float64,
	getColorAt func(int, int) Style, uni bool, emptyBraille rune,
) {
	if !uni {
		gaugeFillRunesGradient(cv, row, x, w, frac, getColorAt, '#', '-')
		return
	}

	br := NewStyle(Default)
	cv.Set(x, row, '[', br)
	cv.Set(x+w-1, row, ']', br)

	barW := w - 2
	if barW < 1 {
		return
	}

	fullBraille := '⣿'

	full := int(frac * float64(barW))
	for i := 0; i < barW; i++ {
		ch := emptyBraille
		if i < full {
			ch = fullBraille
		}

		cv.Set(x+1+i, row, ch, getColorAt(x+1+i, w))
	}
}

func gaugeFillCirclesGradient(cv *Canvas, row, x, w int, frac float64,
	getColorAt func(int, int) Style, uni bool,
) {
	fullCh, emptyCh := '●', '○'
	if !uni {
		fullCh, emptyCh = '#', '-'
	}

	br := NewStyle(Default)
	cv.Set(x, row, '[', br)
	cv.Set(x+w-1, row, ']', br)

	barW := w - 2
	if barW < 1 {
		return
	}

	full := int(frac * float64(barW))
	for i := 0; i < barW; i++ {
		ch := emptyCh
		if i < full {
			ch = fullCh
		}

		cv.Set(x+1+i, row, ch, getColorAt(x+1+i, w))
	}
}

func gaugeFillDotsGradient(cv *Canvas, row, x, w int, frac float64,
	getColorAt func(int, int) Style, uni bool,
) {
	fullCh, emptyCh := '•', '·'
	if !uni {
		fullCh, emptyCh = '#', '-'
	}

	br := NewStyle(Default)
	cv.Set(x, row, '[', br)
	cv.Set(x+w-1, row, ']', br)

	barW := w - 2
	if barW < 1 {
		return
	}

	full := int(frac * float64(barW))
	for i := 0; i < barW; i++ {
		ch := emptyCh
		if i < full {
			ch = fullCh
		}

		cv.Set(x+1+i, row, ch, getColorAt(x+1+i, w))
	}
}

func gaugeFillSquaresGradient(cv *Canvas, row, x, w int, frac float64,
	getColorAt func(int, int) Style, uni bool,
) {
	fullCh, emptyCh := '■', '□'
	if !uni {
		fullCh, emptyCh = '#', '-'
	}

	br := NewStyle(Default)
	cv.Set(x, row, '[', br)
	cv.Set(x+w-1, row, ']', br)

	barW := w - 2
	if barW < 1 {
		return
	}

	full := int(frac * float64(barW))
	for i := 0; i < barW; i++ {
		ch := emptyCh
		if i < full {
			ch = fullCh
		}

		cv.Set(x+1+i, row, ch, getColorAt(x+1+i, w))
	}
}
