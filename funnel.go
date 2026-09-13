package tuichart

import (
	"math"
)

type funnelStep struct {
	name  string
	val   float64
	color Color
}

// FunnelChart renders a process as stacked, centered trapezoids whose width
// is proportional to each stage's value, with a legend to the right.
type FunnelChart struct {
	steps []funnelStep
	chartBase
}

// NewFunnel creates an empty funnel chart.
func NewFunnel() *FunnelChart {
	return &FunnelChart{chartBase: newChartBase()}
}

// Step appends a stage; the widest stage sets the funnel's full width.
func (f *FunnelChart) Step(name string, val float64) *FunnelChart {
	f.steps = append(f.steps, funnelStep{name: name, val: val})
	return f
}

// StepColor overrides the color of the most recently added stage.
func (f *FunnelChart) StepColor(c Color) *FunnelChart {
	if len(f.steps) > 0 {
		f.steps[len(f.steps)-1].color = c
	}
	return f
}

// Title sets the chart title.
func (f *FunnelChart) Title(t string) *FunnelChart { f.SetTitle(t); return f }

// ShowValues includes each stage's raw value in the legend next to its
// share of the top stage.
func (f *FunnelChart) ShowValues(v bool) *FunnelChart { f.SetShowValues(v); return f }

// HeightHint returns the suggested height in rows for the given width.
func (f *FunnelChart) HeightHint(width int) int {
	if h := f.chartBase.HeightHint(width); h > 0 {
		return h
	}
	h := len(f.steps)*2 + 4
	if h < 6 {
		h = 6
	}
	if h > 18 {
		h = 18
	}
	return h
}

// Draw renders the trapezoid stages and legend into the canvas.
func (f *FunnelChart) Draw(rc *Ctx, cv *Canvas) {
	inner := f.frameTitle(cv, rc.Info.Unicode)
	uni := rc.Info.Unicode
	mono := rc.Info.Level == LevelNone

	n := len(f.steps)
	if n == 0 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", NewStyle(Gray))
		return
	}
	maxVal := 0.0
	ci := 0
	for i := range f.steps {
		if f.steps[i].color.IsZero() {
			f.steps[i].color = rc.Palette[ci%len(rc.Palette)]
			ci++
		}
		if f.steps[i].val > maxVal {
			maxVal = f.steps[i].val
		}
	}
	if maxVal <= 0 || inner.W < 8 || inner.H < n {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", NewStyle(Gray))
		return
	}

	gutter := 0
	for i := range f.steps {
		pct := pctOfTop(f.steps[i].val, f.steps[0].val)
		lbl := f.steps[i].name + " " + FormatValue(pct) + "%"
		if f.showVals {
			lbl = f.steps[i].name + " " + FormatValue(
				f.steps[i].val,
			) + " (" + FormatValue(
				pct,
			) + "%)"
		}
		if nl := runeLen(lbl); nl > gutter {
			gutter = nl
		}
	}
	gutter = min(gutter+2, inner.W/2)
	plotW := inner.W - gutter
	if plotW < 4 {
		plotW = 4
	}

	// Stage boundary widths and rows: width i tapers toward width i+1,
	// and the last stage tapers to a point at the bottom. Negative values
	// clamp to a point so the geometry stays inside the plot.
	bw := make([]float64, n+1)
	for i := 0; i < n; i++ {
		if f.steps[i].val > 0 {
			bw[i] = float64(plotW) * f.steps[i].val / maxVal
		} else {
			bw[i] = 0
		}
	}
	bw[n] = 0
	bi := make([]int, n+1)
	for i := 0; i <= n; i++ {
		bi[i] = inner.Y + int(math.Round(float64(inner.H)*float64(i)/float64(n)))
	}
	if bi[n] > inner.Y2() {
		bi[n] = inner.Y2()
	}

	for i := 0; i < n; i++ {
		st := NewStyle(f.steps[i].color)
		fill := '█'
		if mono || !uni {
			fill = pieASCIIChars[i%len(pieASCIIChars)]
		}
		y0, y1 := bi[i], bi[i+1]
		if y1 <= y0 {
			y1 = y0 + 1
		}
		span := y1 - y0
		// The taper row belongs to the stage it tapers into.
		for y := y0; y < y1; y++ {
			frac := float64(y-y0) / float64(span)
			w := bw[i] + (bw[i+1]-bw[i])*frac
			half := int(w / 2)
			x0 := inner.X + plotW/2 - half
			for x := 0; x < int(w); x++ {
				cv.Set(x0+x, y, fill, st)
			}
		}
	}

	// Legend column, right-aligned against the frame.
	for i := 0; i < n; i++ {
		y := bi[i] + (bi[i+1]-bi[i])/2
		if y < inner.Y {
			y = inner.Y
		}
		if y > inner.Y2() {
			y = inner.Y2()
		}
		pct := pctOfTop(f.steps[i].val, f.steps[0].val)
		lbl := f.steps[i].name + " " + FormatValue(pct) + "%"
		if f.showVals {
			lbl = f.steps[i].name + " " + FormatValue(
				f.steps[i].val,
			) + " (" + FormatValue(
				pct,
			) + "%)"
		}
		cv.TextRight(inner.X2(), y, ellipTrunc(lbl, gutter, uni), NewStyle(f.steps[i].color))
	}
}

// pctOfTop reports v as a percentage of top (100 when equal; 0 when top is 0).
func pctOfTop(v, top float64) float64 {
	if top <= 0 {
		return 0
	}
	return v / top * 100
}
