package tuichart

import "math"

type pieSlice struct {
	name  string
	val   float64
	color Color
}

// PieChart renders slices as a pie (or donut) chart with a legend.
type PieChart struct {
	slices []pieSlice
	chartBase
	donut bool
}

// NewPie creates an empty PieChart.
func NewPie() *PieChart {
	return &PieChart{chartBase: newChartBase(), donut: false}
}

// Slice appends a slice; the color is assigned from the palette when omitted.
func (p *PieChart) Slice(name string, val float64) *PieChart {
	p.slices = append(p.slices, pieSlice{name: name, val: val})
	return p
}

// SliceColor overrides the color of the most recently added slice.
func (p *PieChart) SliceColor(c Color) *PieChart {
	if len(p.slices) > 0 {
		p.slices[len(p.slices)-1].color = c
	}

	return p
}

// Donut enables the donut style (hollow center) when on is true.
func (p *PieChart) Donut(on bool) *PieChart { p.donut = on; return p }

// ShowValues includes each slice's raw value in the legend next to its
// percentage (terminal pie legends are the practical spot for numbers).
func (p *PieChart) ShowValues(v bool) *PieChart { p.SetShowValues(v); return p }

// Title sets the chart title.
func (p *PieChart) Title(t string) *PieChart { p.SetTitle(t); return p }

// HeightHint returns the suggested height in rows for the given width.
func (p *PieChart) HeightHint(width int) int {
	if h := p.chartBase.HeightHint(width); h > 0 {
		return h
	}

	return clampInt(width*3/5, 8, 22)
}

var pieASCIIChars = []rune{'#', '@', '*', 'o', '=', '+', '~', '%'}

// Draw renders the pie sectors and legend into the canvas.
func (p *PieChart) Draw(rc *Ctx, cv *Canvas) {
	total := 0.0

	for i := range p.slices {
		if p.slices[i].color.IsZero() {
			p.slices[i].color = rc.Next()
		}

		total += math.Max(p.slices[i].val, 0)
	}

	inner := p.frameTitle(cv, rc.Info.Unicode)

	if total <= 0 || inner.W < 6 || inner.H < 4 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", NewStyle(Gray))
		return
	}

	reserve := (inner.W*2)/5 + 1
	if reserve > len(p.slices)*14+6 {
		reserve = len(p.slices)*14 + 6
	}

	cx := inner.X + (inner.W-reserve)/2
	cy := inner.Y + inner.H/2
	rx := (inner.W - reserve - 1) / 2

	ry := (inner.H - 1) / 2
	if rx < 1 || ry < 1 {
		rx = max(rx, 1)
		ry = max(ry, 1)
	}

	var entries []LegendEntry

	acc := 0.0

	for i := range p.slices {
		s := &p.slices[i]
		v := math.Max(s.val, 0)
		from := acc / total
		acc += v
		to := acc / total
		st := NewStyle(s.color)
		mono := rc.Info.Level == LevelNone

		fillCh := '█'
		if mono || !rc.Info.Unicode {
			fillCh = pieASCIIChars[i%len(pieASCIIChars)]
		}

		drawPieSector(cv, cx, cy, rx, ry, from, to, fillCh, st, p.donut)

		pct := v / total * 100
		pctStr := FormatValue(math.Round(pct*10) / 10)

		label := s.name + " " + pctStr + "%"
		if p.showVals {
			label = s.name + " " + FormatValue(v) + " (" + pctStr + "%)"
		}

		glyph := "██"
		if mono || !rc.Info.Unicode {
			glyph = string(fillCh) + string(fillCh)
		}

		entries = append(entries, LegendEntry{
			Label: label,
			Style: st,
			Glyph: glyph,
		})
	}

	lx := min(cx+rx+2, inner.X2()-12)

	for i, e := range entries {
		y := cy - len(entries)/2 + i
		if y > inner.Y2() || y < 0 {
			continue
		}

		n := cv.Text(lx, y, e.Glyph, e.Style)
		cv.Text(lx+n+1, y, ellipTrunc(e.Label, inner.X2()-lx+1-n-2, rc.Info.Unicode), e.Style)
	}
}

func drawPieSector(
	cv *Canvas,
	cx, cy, rx, ry int,
	from, to float64,
	ch rune,
	st Style,
	donut bool,
) {
	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			nx := float64(dx) / float64(rx)
			ny := float64(dy) / float64(ry)

			d := nx*nx + ny*ny
			if d > 1 {
				continue
			}

			if donut && d < 0.30 {
				continue
			}

			ang := math.Atan2(nx, -ny)
			if ang < 0 {
				ang += 2 * math.Pi
			}

			t := ang / (2 * math.Pi)
			if t >= from-1e-9 && t < to+1e-9 {
				cv.Set(cx+dx, cy+dy, ch, st)
			}
		}
	}
}
