package tuichart

import (
	"math"
	"slices"
)

type radarSeries struct {
	name  string
	vals  []float64
	color Color
}

// RadarChart compares several series across the same set of variables,
// drawn as concentric rings with one spoke per variable. Each series is a
// polygon whose vertex on every spoke is its value along that variable.
type RadarChart struct {
	axes   []string
	series []radarSeries
	chartBase
	max  float64
	fill bool
}

// NewRadar creates an empty radar chart.
func NewRadar() *RadarChart {
	return &RadarChart{chartBase: newChartBase()}
}

// Axes names the variables; each series must provide one value per axis.
func (r *RadarChart) Axes(names ...string) *RadarChart {
	r.axes = names
	return r
}

// Series appends a polygon across the axes. Values below the shared max
// scale shrink the vertex toward the center.
func (r *RadarChart) Series(name string, vals ...float64) *RadarChart {
	r.series = append(r.series, radarSeries{name: name, vals: vals})
	return r
}

// SeriesColor overrides the color of the most recently added series.
func (r *RadarChart) SeriesColor(c Color) *RadarChart {
	if len(r.series) > 0 {
		r.series[len(r.series)-1].color = c
	}
	return r
}

// Title sets the chart title.
func (r *RadarChart) Title(t string) *RadarChart { r.SetTitle(t); return r }

// Max pins the shared axis scale; when 0 (default) it derives from the data.
func (r *RadarChart) Max(m float64) *RadarChart { r.max = m; return r }

// Fill shades the inside of each polygon behind its outline.
func (r *RadarChart) Fill(on bool) *RadarChart { r.fill = on; return r }

// ShowValues enables or disables the numeric value labels.
func (r *RadarChart) ShowValues(v bool) *RadarChart { r.SetShowValues(v); return r }

// HeightHint returns the suggested height in rows for the given width.
func (r *RadarChart) HeightHint(width int) int {
	if h := r.chartBase.HeightHint(width); h > 0 {
		return h
	}
	h := width/2 + 4
	if h > 24 {
		h = 24
	}
	if h < 8 {
		h = 8
	}
	return h
}

type radarPt struct {
	x, y float64
	xi   int
	yi   int
}

// Draw renders the rings, spokes, series polygons, and labels.
func (r *RadarChart) Draw(rc *Ctx, cv *Canvas) {
	inner := r.frameTitle(cv, rc.Info.Unicode)
	uni := rc.Info.Unicode

	if len(r.axes) < 3 || inner.W < 10 || inner.H < 6 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", NewStyle(Gray))
		return
	}

	scale := r.max
	if scale <= 0 {
		for _, s := range r.series {
			for _, v := range s.vals {
				if v > scale {
					scale = v
				}
			}
		}
	}
	if scale <= 0 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", NewStyle(Gray))
		return
	}

	cx, cy := inner.X+inner.W/2, inner.Y+inner.H/2
	rx := inner.W/2 - 2
	ry := inner.H/2 - 2
	if rx < 2 {
		rx = 2
	}
	if ry < 2 {
		ry = 2
	}

	n := len(r.axes)

	// Ring polygons and spokes: braille dots when Unicode is on, slope
	// glyphs otherwise so ASCII stays pure.
	ring := make([]radarPt, n)
	ringHalf := make([]radarPt, n)
	for i := 0; i < n; i++ {
		a := ringAngle(i, n)
		ring[i] = polarToPt(cx, cy, rx, ry, 1.0, a)
		ringHalf[i] = polarToPt(cx, cy, rx, ry, 0.5, a)
	}
	dim := DimGray
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		if uni {
			dotLine(cv, ring[i].xi, ring[i].yi, ring[j].xi, ring[j].yi, dim, false)
			dotLine(cv, ringHalf[i].xi, ringHalf[i].yi, ringHalf[j].xi, ringHalf[j].yi, dim, false)
			dotLine(cv, cx*2, cy*4, ring[i].xi, ring[i].yi, dim, false)
		} else {
			asciiLine(cv, ring[i].cellX(), ring[i].cellY(),
				ring[j].cellX(), ring[j].cellY(), NewStyle(dim))
			asciiLine(cv, ringHalf[i].cellX(), ringHalf[i].cellY(),
				ringHalf[j].cellX(), ringHalf[j].cellY(), NewStyle(dim))
			asciiLine(cv, cx, cy, ring[i].cellX(), ring[i].cellY(), NewStyle(dim))
		}
	}

	ci := 0
	for si := range r.series {
		s := &r.series[si]
		if s.color.IsZero() {
			s.color = rc.Palette[ci%len(rc.Palette)]
			ci++
		}
	}

	for si := range r.series {
		s := &r.series[si]
		var pts []radarPt
		for i := 0; i < n; i++ {
			f := 0.0
			if i < len(s.vals) && s.vals[i] > 0 {
				f = s.vals[i] / scale
			}
			if f > 1 {
				f = 1
			}
			pts = append(pts, polarToPt(cx, cy, rx, ry, f, ringAngle(i, n)))
		}
		if uni {
			if r.fill {
				fillDotPolygon(cv, pts, s.color)
			}
			for i := 0; i < n; i++ {
				j := (i + 1) % n
				dotLine(cv, pts[i].xi, pts[i].yi, pts[j].xi, pts[j].yi, s.color, false)
			}
		} else {
			for i := 0; i < n; i++ {
				j := (i + 1) % n
				asciiLine(cv, pts[i].cellX(), pts[i].cellY(),
					pts[j].cellX(), pts[j].cellY(), NewStyle(s.color))
			}
		}
	}

	// Axis labels just outside the outer ring, clamped inside the frame.
	for i := 0; i < n; i++ {
		a := ringAngle(i, n)
		lx := cx + int(float64(rx+3)*math.Cos(a))
		ly := cy + int(float64(ry+2)*math.Sin(a))
		lbl := ellipTrunc(r.axes[i], max(inner.W/3, 1), uni)
		w := runeLen(lbl)
		x := lx
		if math.Abs(math.Cos(a)) < 0.3 {
			x = lx - w/2
		} else if math.Cos(a) < 0 {
			x = lx - w
		}
		x = max(inner.X, min(inner.X2()-w, x))
		ly = max(inner.Y, min(inner.Y2(), ly))
		cv.Text(x, ly, lbl, NewStyle(Silver))
	}

	// Legend in monochrome-safe form: series name + glyph.
	glyph := "──"
	if !uni {
		glyph = "=="
	}
	var entries []LegendEntry
	for si := range r.series {
		entries = append(entries, LegendEntry{
			Label: r.series[si].name,
			Style: NewStyle(r.series[si].color),
			Glyph: glyph,
		})
	}
	drawLegendInside(cv, inner, entries, uni)
}

func ringAngle(i, n int) float64 {
	return -math.Pi/2 + float64(i)*2*math.Pi/float64(n)
}

func polarToPt(cx, cy, rx, ry int, frac float64, a float64) radarPt {
	x := float64(cx) + float64(rx)*frac*math.Cos(a)
	y := float64(cy) + float64(ry)*frac*math.Sin(a)
	return radarPt{
		x:  x,
		y:  y,
		xi: int(math.Round(x * 2)),
		yi: int(math.Round(y * 4)),
	}
}

func (p radarPt) cellX() int { return int(math.Round(p.x)) }
func (p radarPt) cellY() int { return int(math.Round(p.y)) }

// fillDotPolygon shades the interior of a polygon by scanfilling each dot
// row between the polygon's left and right edges.
func fillDotPolygon(cv *Canvas, pts []radarPt, c Color) {
	minY, maxY := pts[0].yi, pts[0].yi
	for _, p := range pts {
		minY = min(minY, p.yi)
		maxY = max(maxY, p.yi)
	}
	for gy := minY; gy <= maxY; gy++ {
		if gy < 0 || gy >= cv.h*4 {
			continue
		}
		var xs []float64
		for i := 0; i < len(pts); i++ {
			p1, p2 := pts[i], pts[(i+1)%len(pts)]
			if p1.yi == p2.yi {
				continue
			}
			if (p1.yi <= gy && p2.yi > gy) || (p2.yi <= gy && p1.yi > gy) {
				t := float64(gy-p1.yi) / float64(p2.yi-p1.yi)
				xs = append(xs, p1.x+(p2.x-p1.x)*t)
			}
		}
		if len(xs) < 2 {
			continue
		}
		slices.Sort(xs)
		for i := 0; i+1 < len(xs); i += 2 {
			x1 := int(math.Round(xs[i] * 2))
			x2 := int(math.Round(xs[i+1] * 2))
			for gxd := x1; gxd <= x2; gxd++ {
				setDot(cv, gxd, gy, c)
			}
		}
	}
}
