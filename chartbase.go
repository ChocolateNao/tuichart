package tuichart

import "time"

const (
	defaultTickTarget = 5
	defaultPad        = 0.06
)

// Align controls where a title is placed along its row.
type Align uint8

const (
	// AlignLeft places the title at the start of its row.
	AlignLeft Align = iota
	// AlignCenter places the title in the middle of its row.
	AlignCenter
	// AlignRight places the title at the end of its row.
	AlignRight
)

// Orientation selects which axis carries values.
type Orientation uint8

const (
	// OrientAuto keeps each diagram's default (bars grow upward,
	// plots map x→columns).
	OrientAuto Orientation = iota
	// OrientHorizontal puts values along X: bars grow rightward, plots
	// swap axes so y=f(x) becomes x=f(y).
	OrientHorizontal
	// OrientVertical puts values along Y: bars grow upward, plots keep
	// the standard orientation even if legacy flags said otherwise.
	OrientVertical
)

// AxisConfig holds the axis, scale and frame configuration shared by
// diagrams that plot values against axes: axis labels, ticks, formatters,
// explicit ranges and the grid/frame/legend toggles. Embed it to get the
// axis half of the Set*/Reset* API promoted onto your type.
type AxisConfig struct {
	xFmt   func(float64) string
	yFmt   func(float64) string
	xLabel string
	yLabel string
	xTicks []Tick
	yTicks []Tick
	x0     float64
	x1     float64
	y0     float64
	y1     float64
	tickN  int
	grid   bool
	frame  bool
	legend bool
	xSet   bool
	ySet   bool
}

func newAxisConfig() AxisConfig {
	return AxisConfig{grid: true, frame: true, legend: true, tickN: defaultTickTarget}
}

// SetXLabel sets the label text displayed below the X axis.
func (a *AxisConfig) SetXLabel(l string) { a.xLabel = l }

// SetYLabel sets the label text displayed beside the Y axis.
func (a *AxisConfig) SetYLabel(l string) { a.yLabel = l }

// ResetLabels clears both axis labels.
func (a *AxisConfig) ResetLabels() { a.xLabel, a.yLabel = "", "" }

// SetXTicks overrides the auto-generated X-axis ticks with explicit values.
func (a *AxisConfig) SetXTicks(t []Tick) { a.xTicks = t }

// SetYTicks overrides the auto-generated Y-axis ticks with explicit values.
func (a *AxisConfig) SetYTicks(t []Tick) { a.yTicks = t }

// ResetTicks clears custom ticks, restoring auto-generation.
func (a *AxisConfig) ResetTicks() { a.xTicks, a.yTicks = nil, nil }

// SetXFormatter provides a custom formatter for X-axis tick labels.
func (a *AxisConfig) SetXFormatter(f func(float64) string) { a.xFmt = f }

// SetYFormatter provides a custom formatter for Y-axis tick labels.
func (a *AxisConfig) SetYFormatter(f func(float64) string) { a.yFmt = f }

// ResetFormatters clears custom formatters, restoring default formatting.
func (a *AxisConfig) ResetFormatters() { a.xFmt, a.yFmt = nil, nil }

// SetScale fixes both axis ranges, disabling auto-scaling.
func (a *AxisConfig) SetScale(x0, x1, y0, y1 float64) {
	a.x0, a.x1, a.y0, a.y1 = x0, x1, y0, y1
	a.xSet, a.ySet = true, true
}

// SetXRange fixes the X axis range, disabling auto-scaling on that axis.
func (a *AxisConfig) SetXRange(lo, hi float64) { a.x0, a.x1, a.xSet = lo, hi, true }

// SetYRange fixes the Y axis range, disabling auto-scaling on that axis.
func (a *AxisConfig) SetYRange(lo, hi float64) { a.y0, a.y1, a.ySet = lo, hi, true }

// ResetScale re-enables automatic scaling from data.
func (a *AxisConfig) ResetScale() {
	a.xSet, a.ySet = false, false
	a.xTicks, a.yTicks = nil, nil
}

// SetGrid enables or disables the background grid lines.
func (a *AxisConfig) SetGrid(on bool) { a.grid = on }

// SetBorder enables or disables the frame border around the diagram.
func (a *AxisConfig) SetBorder(on bool) { a.frame = on }

// SetLegend enables or disables the legend box.
func (a *AxisConfig) SetLegend(on bool) { a.legend = on }

// SetTickCount sets the target number of ticks per axis; minimum is 2.
func (a *AxisConfig) SetTickCount(n int) {
	if n < 2 {
		n = 2
	}

	a.tickN = n
}

// Reset restores every axis property to its default.
func (a *AxisConfig) Reset() { *a = newAxisConfig() }

// DisplayConfig holds presentation configuration that applies to every
// diagram whether or not it has axes: frame title and alignment, orientation,
// value annotations, cell width and pinned height. Embed it to get the
// presentation half of the Set*/Reset* API promoted onto your type.
type DisplayConfig struct {
	title      string
	titleAlign Align
	orient     Orientation
	showVals   bool
	cellW      int
	height     int
}

func newDisplayConfig() DisplayConfig { return DisplayConfig{} }

// SetTitle sets the diagram's frame title.
func (d *DisplayConfig) SetTitle(t string) { d.title = t }

// GetTitle returns the frame title. It makes every diagram embedding
// DisplayConfig satisfy the Diagram interface.
func (d *DisplayConfig) GetTitle() string { return d.title }

// ResetTitle clears the diagram's frame title.
func (d *DisplayConfig) ResetTitle() { d.title = "" }

// SetShowValues toggles numeric value annotations (bar tops, heatmap
// cells, pie legend values).
func (d *DisplayConfig) SetShowValues(v bool) { d.showVals = v }

// ResetShowValues disables numeric value annotations.
func (d *DisplayConfig) ResetShowValues() { d.showVals = false }

// SetCellWidth fixes the heatmap cell width in columns; 0 restores the
// automatic stretch that fills the frame width.
func (d *DisplayConfig) SetCellWidth(n int) { d.cellW = n }

// ResetCellWidth restores automatic heatmap cell width.
func (d *DisplayConfig) ResetCellWidth() { d.cellW = 0 }

// SetTitleAlign sets where the diagram title sits within its frame.
func (d *DisplayConfig) SetTitleAlign(a Align) { d.titleAlign = a }

// ResetTitleAlign restores left alignment.
func (d *DisplayConfig) ResetTitleAlign() { d.titleAlign = AlignLeft }

// SetOrientation switches the value axis; see Orientation for semantics.
// Diagrams that cannot swap axes ignore it.
func (d *DisplayConfig) SetOrientation(o Orientation) { d.orient = o }

// ResetOrientation restores each diagram's default axis layout.
func (d *DisplayConfig) ResetOrientation() { d.orient = OrientAuto }

// SetSize pins this diagram's height in rows when rendered in a Board.
func (d *DisplayConfig) SetSize(rows int) { d.height = rows }

// ResetSize clears the pinned height, letting the Board container decide.
func (d *DisplayConfig) ResetSize() { d.height = 0 }

// HeightHint returns the pinned height or 0 to let the container decide.
func (d *DisplayConfig) HeightHint(int) int {
	if d.height > 0 {
		return d.height
	}

	return 0
}

// WidthHint reports the width this diagram would like for the given height.
// Zero means "no preference"; horizontal layouts may negotiate with it.
func (d *DisplayConfig) WidthHint(int) int { return 0 }

// plotHeightHint returns the height hint shared by plot-like diagrams
// (Plot, TimeSeries, Candlestick) when no explicit height is set.
func (d *DisplayConfig) plotHeightHint(width int) int {
	if h := d.HeightHint(width); h > 0 {
		return h
	}

	return clampInt(width*2/5, 9, 24)
}

// Reset restores every presentation property to its default, keeping the
// title so Reset can be used to clear a board's styling in one call.
func (d *DisplayConfig) Reset() {
	title := d.title
	*d = newDisplayConfig()
	d.title = title
}

// frameTitle paints the diagram frame and title, returning the rectangle
// left for content. frame comes from AxisConfig, title and align from
// DisplayConfig.
func frameTitle(cv *Canvas, uni bool, title string, align Align, frame bool) Rect {
	r := cv.Rect()
	if !frame {
		if title != "" && r.H > 1 {
			drawAlignedText(
				cv,
				0,
				ellipTrunc(title, r.W, uni),
				NewStyle(Default).Bolder(),
				align,
				r.W,
			)

			return Rect{X: 0, Y: 1, W: r.W, H: r.H - 1}.clip(r)
		}

		return r
	}

	st := NewStyle(Gray)
	cv.Border(st, uni)

	if title != "" && r.H > 2 {
		t := " " + title + " "
		if runeLen(t)+4 > r.W {
			t = " " + ellipTrunc(title, max(r.W-6, 1), uni) + " "
		}

		drawAlignedText(cv, 0, t, NewStyle(Default).Bolder(), align, r.W)

		return Rect{X: 1, Y: 1, W: r.W - 2, H: r.H - 2}.clip(r)
	}

	return Rect{X: 1, Y: 1, W: r.W - 2, H: r.H - 2}.clip(r)
}

// drawAlignedText places s on row y honoring a within width w. Left-aligned
// text keeps a small margin so it clears frame borders.
func drawAlignedText(cv *Canvas, y int, s string, st Style, a Align, w int) {
	n := runeLen(s)

	switch a {
	case AlignRight:
		x := w - n - 1
		if x < 0 {
			x = 0
		}

		cv.Text(x, y, s, st)
	case AlignCenter:
		cv.TextCenter(w/2, y, s, st)
	default:
		x := 0
		if w > n+3 {
			x = 2
		}

		cv.Text(x, y, s, st)
	}
}

func (r Rect) clip(o Rect) Rect {
	if r.X < o.X {
		r.W -= o.X - r.X
		r.X = o.X
	}

	if r.Y < o.Y {
		r.H -= o.Y - r.Y
		r.Y = o.Y
	}

	if r.W > o.W-(r.X-o.X) {
		r.W = o.W - (r.X - o.X)
	}

	if r.H > o.H-(r.Y-o.Y) {
		r.H = o.H - (r.Y - o.Y)
	}

	if r.W < 0 {
		r.W = 0
	}

	if r.H < 0 {
		r.H = 0
	}

	return r
}

// clampInt bounds v to [lo, hi].
func clampInt(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

// rampIdx maps a [0,1] fraction to a clamped ramp index.
func rampIdx(t float64, n int) int { return clampInt(int(t*float64(n-1)+0.5), 0, n-1) }

// horizontalBarHeight sizes a left-to-right bar chart: one row per
// category plus border/margin allowance, bounded like the other diagrams.
func horizontalBarHeight(cats int) int {
	return clampInt(cats+3, 6, 30)
}

// writeLabel draws a centered label at the given column, clamped to the
// frame: it shifts sideways near the edges before falling back to
// truncation, so full text survives whenever it fits somewhere on the row.
func writeLabel(cv *Canvas, row, col int, text string, st Style, inner Rect, uni bool) {
	if row < inner.Y || row > inner.Y2() {
		return
	}

	lbl := ellipTrunc(text, inner.W-2, uni)

	start := max(col-runeLen(lbl)/2, inner.X)

	if end := start + runeLen(lbl) - 1; end > inner.X2() {
		// Prefer shifting left over truncating.
		start = inner.X2() - runeLen(lbl) + 1
		if start < inner.X {
			start = inner.X
			lbl = ellipTrunc(text, inner.X2()-inner.X+1, uni)
		}
	}

	if runeLen(lbl) == 0 {
		return
	}

	clearAndWrite(cv, row, start, lbl, st)
}

// clearAndWrite overwrites the target segment so labels never mix with
// leftover glyphs.
func clearAndWrite(cv *Canvas, y, x int, s string, st Style) {
	for i := 0; i < runeLen(s); i++ {
		cv.Set(x+i, y, ' ', Style{})
	}

	cv.Text(x, y, s, st)
}

func autoTimeLayout(spanSec float64) string {
	switch {
	case spanSec < 90:
		return "15:04:05"
	case spanSec < 24*3600:
		return "15:04"
	case spanSec < 30*24*3600:
		return "Jan 02 15:04"
	default:
		return "2006-01-02"
	}
}

// timeTickFmt resolves the X-tick formatter for time-based diagrams: the
// caller's own formatter wins; otherwise the configured Go layout; finally
// autoTimeLayout for the data span.
func timeTickFmt(xFmt func(float64) string, layout string, span float64) func(float64) string {
	if xFmt != nil {
		return xFmt
	}

	if layout == "" {
		layout = autoTimeLayout(span)
	}

	return func(v float64) string { return time.Unix(int64(v), 0).Format(layout) }
}
