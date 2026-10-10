package tuichart

import "math"

// Point is a single data point with X and Y coordinates.
type Point struct{ X, Y float64 }

// Seq creates a sequence of Points with X = 0, 1, 2, ... and the given y values.
func Seq(vals ...float64) []Point {
	out := make([]Point, len(vals))
	for i, v := range vals {
		out[i] = Point{X: float64(i), Y: v}
	}

	return out
}

// Line is a series rendered as connected line segments.
type Line struct {
	name   string
	pts    []Point
	color  Color
	marker rune
	dashed bool
}

// NewLine creates a named Line series from the given points.
func NewLine(name string, pts ...Point) *Line {
	return &Line{name: name, pts: pts}
}

// NewLineVals creates a named Line series from y values, assigning X = 0, 1, 2, ...
func NewLineVals(name string, vals []float64) *Line {
	return &Line{name: name, pts: Seq(vals...)}
}

// Points appends one or more Points to the line.
func (l *Line) Points(pts ...Point) *Line { l.pts = append(l.pts, pts...); return l }

// Values appends y values (with sequential X) to the line.
func (l *Line) Values(vals ...float64) *Line { l.pts = append(l.pts, Seq(vals...)...); return l }

// SetValues replaces all points with the given y values at x = 0..n-1.
// Intended for live charts: assign Ring.Values() on every tick.
func (l *Line) SetValues(vals []float64) *Line { l.pts = Seq(vals...); return l }

// Color sets the line color.
func (l *Line) Color(c Color) *Line { l.color = c; return l }

// Marker sets a glyph rendered at each data point.
func (l *Line) Marker(m rune) *Line { l.marker = m; return l }

// Dashed toggles a dashed line style instead of a solid line.
func (l *Line) Dashed(d bool) *Line { l.dashed = d; return l }

// SetName changes the series legend label.
func (l *Line) SetName(s string) *Line { l.name = s; return l }

func (l *Line) bounds(db *dataBounds) {
	for _, p := range l.pts {
		db.add(p.X, p.Y)
	}
}
func (l *Line) hasColor() bool { return !l.color.IsZero() }
func (l *Line) setColor(c Color) {
	if l.color.IsZero() {
		l.color = c
	}
}
func (l *Line) colorOf() Color { return l.color }

type projPt struct{ x, y float64 }

func project(fr frame, p Point) projPt {
	return projPt{x: fr.mx(p.X), y: fr.my(p.Y)}
}

func (l *Line) draw(cv *Canvas, fr frame, st Style) {
	if len(l.pts) == 0 {
		return
	}

	proj := make([]projPt, len(l.pts))
	for i, p := range l.pts {
		proj[i] = project(fr, p)
	}

	area := fr.area

	for i := 0; i < len(proj)-1; i++ {
		a, b := proj[i], proj[i+1]
		if math.IsNaN(a.x) || math.IsNaN(a.y) || math.IsNaN(b.x) || math.IsNaN(b.y) {
			continue
		}

		if fr.uni && fr.ysc.Kind == Linear && fr.xsc.Kind == Linear {
			gx0 := int(math.Round(
				math.Max(float64(area.X-1), math.Min(a.x, float64(area.X2()+1))) * 2,
			))
			gy0 := int(math.Round(
				math.Max(float64(area.Y-1), math.Min(a.y, float64(area.Y2()+1))) * 4,
			))
			gx1 := int(math.Round(
				math.Max(float64(area.X-1), math.Min(b.x, float64(area.X2()+1))) * 2,
			))
			gy1 := int(math.Round(
				math.Max(float64(area.Y-1), math.Min(b.y, float64(area.Y2()+1))) * 4,
			))
			dotLine(cv, gx0, gy0, gx1, gy1, st.Fg, l.dashed)
		} else {
			x0 := int(math.Round(
				math.Max(float64(area.X), math.Min(a.x, float64(area.X2()))),
			))
			y0 := int(math.Round(
				math.Max(float64(area.Y), math.Min(a.y, float64(area.Y2()))),
			))
			x1 := int(math.Round(
				math.Max(float64(area.X), math.Min(b.x, float64(area.X2()))),
			))
			y1 := int(math.Round(
				math.Max(float64(area.Y), math.Min(b.y, float64(area.Y2()))),
			))
			asciiLine(cv, x0, y0, x1, y1, st)
		}
	}

	l.drawMarkers(cv, fr, st)
}

func (l *Line) drawMarkers(cv *Canvas, fr frame, st Style) {
	if l.marker == 0 {
		return
	}

	m := l.marker

	area := fr.area
	for _, p := range l.pts {
		q := project(fr, p)
		if math.IsNaN(q.x) || math.IsNaN(q.y) {
			continue
		}

		x := int(math.Round(
			math.Max(float64(area.X), math.Min(q.x, float64(area.X2()))),
		))
		y := int(math.Round(
			math.Max(float64(area.Y), math.Min(q.y, float64(area.Y2()))),
		))
		cv.Set(x, y, m, st)
	}
}

func (l *Line) legendEntry(st Style, uni bool) LegendEntry {
	glyph := "───"
	if !uni {
		glyph = "---"
	}

	return LegendEntry{Label: l.name, Style: st, Glyph: glyph}
}

// Scatter is a series rendered as discrete point markers.
type Scatter struct {
	name   string
	pts    []Point
	color  Color
	marker rune
}

// NewScatter creates a named Scatter series from the given points.
func NewScatter(name string, pts ...Point) *Scatter {
	return &Scatter{name: name, pts: pts}
}

// NewScatterVals creates a named Scatter series from y values, assigning X = 0, 1, 2, ...
func NewScatterVals(name string, vals []float64) *Scatter {
	return &Scatter{name: name, pts: Seq(vals...)}
}

// Points appends one or more Points to the scatter series.
func (s *Scatter) Points(pts ...Point) *Scatter { s.pts = append(s.pts, pts...); return s }

// Color sets the marker color.
func (s *Scatter) Color(c Color) *Scatter { s.color = c; return s }

// Marker sets the glyph used to render each data point.
func (s *Scatter) Marker(m rune) *Scatter { s.marker = m; return s }

// SetName changes the series legend label.
func (s *Scatter) SetName(n string) *Scatter { s.name = n; return s }

func (s *Scatter) bounds(db *dataBounds) {
	for _, p := range s.pts {
		db.add(p.X, p.Y)
	}
}
func (s *Scatter) hasColor() bool { return !s.color.IsZero() }
func (s *Scatter) setColor(c Color) {
	if s.color.IsZero() {
		s.color = c
	}
}
func (s *Scatter) colorOf() Color { return s.color }

func (s *Scatter) draw(cv *Canvas, fr frame, st Style) {
	area := fr.area

	m := s.marker
	if m == 0 {
		if fr.uni {
			m = '●'
		} else {
			m = 'o'
		}
	}

	for _, p := range s.pts {
		q := project(fr, p)
		if math.IsNaN(q.x) || math.IsNaN(q.y) {
			continue
		}

		x := int(math.Round(
			math.Max(float64(area.X), math.Min(q.x, float64(area.X2()))),
		))
		y := int(math.Round(
			math.Max(float64(area.Y), math.Min(q.y, float64(area.Y2()))),
		))
		cv.Set(x, y, m, st)
	}
}

func (s *Scatter) legendEntry(st Style, uni bool) LegendEntry {
	glyph := "●●"
	if !uni {
		glyph = "oo"
	}

	return LegendEntry{Label: s.name, Style: st, Glyph: glyph}
}

// Plot is an XY chart containing line and scatter series.
type Plot struct {
	order []any // *Line or *Scatter
	configs
	xKind Kind
	yKind Kind
}

// NewPlot creates a new Plot.
func NewPlot() *Plot {
	return &Plot{AxisConfig: newAxisConfig(), DisplayConfig: newDisplayConfig()}
}

// Title sets the plot title.
func (p *Plot) Title(t string) *Plot { p.SetTitle(t); return p }

// XLabel sets the x axis label.
func (p *Plot) XLabel(l string) *Plot { p.SetXLabel(l); return p }

// YLabel sets the y axis label.
func (p *Plot) YLabel(l string) *Plot { p.SetYLabel(l); return p }

// Grid toggles grid lines.
func (p *Plot) Grid(on bool) *Plot { p.SetGrid(on); return p }

// Legend toggles the series legend.
func (p *Plot) Legend(on bool) *Plot { p.SetLegend(on); return p }

// Height pins the plot height in rows.
func (p *Plot) Height(rows int) *Plot { p.SetSize(rows); return p }

// Orientation swaps the axes when given OrientHorizontal; OrientVertical
// or OrientAuto restore the standard x→columns layout.
func (p *Plot) Orientation(o Orientation) *Plot { p.SetOrientation(o); return p }

// Add appends one or more *Line or *Scatter series to the plot.
func (p *Plot) Add(ss ...any) *Plot {
	for _, s := range ss {
		switch v := s.(type) {
		case *Line:
			p.order = append(p.order, v)
		case *Scatter:
			p.order = append(p.order, v)
		}
	}

	return p
}

// LogX switches the x axis to a logarithmic scale.
func (p *Plot) LogX(on bool) *Plot {
	if on {
		p.xKind = Logarithmic
	} else {
		p.xKind = Linear
	}

	return p
}

// LogY switches the y axis to a logarithmic scale.
func (p *Plot) LogY(on bool) *Plot {
	if on {
		p.yKind = Logarithmic
	} else {
		p.yKind = Linear
	}

	return p
}

// HeightHint returns the suggested height for the given width.
func (p *Plot) HeightHint(width int) int {
	return p.plotHeightHint(width)
}

// WidthHint returns the suggested width for the given height.
func (p *Plot) WidthHint(height int) int {
	return max(10, min(height*2, 100))
}

// Draw renders the plot onto the canvas.
func (p *Plot) Draw(rc *Ctx, cv *Canvas) {
	ax := &p.AxisConfig
	dc := &p.DisplayConfig
	xk, yk := p.xKind, p.yKind

	swap := dc.orient == OrientHorizontal
	if swap {
		// Present the transposed view: series points are swapped and all
		// per-axis configuration follows them.
		cp := *ax
		cp.xLabel, cp.yLabel = ax.yLabel, ax.xLabel
		cp.xTicks, cp.yTicks = ax.yTicks, ax.xTicks
		cp.xFmt, cp.yFmt = ax.yFmt, ax.xFmt
		cp.x0, cp.x1 = ax.y0, ax.y1
		cp.y0, cp.y1 = ax.x0, ax.x1
		cp.xSet, cp.ySet = ax.ySet, ax.xSet
		xk, yk = yk, xk
		ax = &cp
	}

	var db dataBounds

	db.empty = true

	for _, s := range p.order {
		switch v := s.(type) {
		case *Line:
			maybeSwapLine(v, swap).bounds(&db)
		case *Scatter:
			maybeSwapScatter(v, swap).bounds(&db)
		}
	}

	for _, s := range p.order {
		switch v := s.(type) {
		case *Line:
			if !v.hasColor() {
				c, rc2 := rc.WithNextColor()
				v.setColor(c)

				rc = rc2
			}
		case *Scatter:
			if !v.hasColor() {
				c, rc2 := rc.WithNextColor()
				v.setColor(c)

				rc = rc2
			}
		}
	}

	fr := prepareFrame(cv, rc, ax, dc, db, xk, yk, true)

	for _, s := range p.order {
		switch v := s.(type) {
		case *Line:
			st := NewStyle(v.colorOf())
			maybeSwapLine(v, swap).draw(cv, fr, st)
		case *Scatter:
			st := NewStyle(v.colorOf())
			maybeSwapScatter(v, swap).draw(cv, fr, st)
		}
	}

	entries := make([]LegendEntry, 0, len(p.order))
	for _, s := range p.order {
		switch v := s.(type) {
		case *Line:
			entries = append(entries,
				maybeSwapLine(v, swap).legendEntry(NewStyle(v.colorOf()), rc.Info.Unicode),
			)
		case *Scatter:
			entries = append(entries,
				maybeSwapScatter(v, swap).legendEntry(NewStyle(v.colorOf()), rc.Info.Unicode),
			)
		}
	}

	drawLegendInside(cv, fr.area, entries, rc.Info.Unicode)

	if db.empty && p.title == "" && len(p.order) == 0 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", NewStyle(Gray))
	}
}

// maybeSwapLine returns a shallow copy of the Line with transposed points
// when swap is set.
func maybeSwapLine(s *Line, swap bool) *Line {
	if !swap {
		return s
	}

	c := *s
	c.pts = transposePts(s.pts)

	return &c
}

// maybeSwapScatter returns a shallow copy of the Scatter with transposed points
// when swap is set.
func maybeSwapScatter(s *Scatter, swap bool) *Scatter {
	if !swap {
		return s
	}

	c := *s
	c.pts = transposePts(s.pts)

	return &c
}

func transposePts(pts []Point) []Point {
	out := make([]Point, len(pts))
	for i, p := range pts {
		out[i] = Point{X: p.Y, Y: p.X}
	}

	return out
}
