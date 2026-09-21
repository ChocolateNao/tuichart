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

// chartBase holds configuration shared by all built-in diagrams, including
// the Set*/Reset* API for size, scale and axis marks.
type chartBase struct {
	yFmt       func(float64) string
	xFmt       func(float64) string
	title      string
	xLabel     string
	yLabel     string
	yTicks     []Tick
	xTicks     []Tick
	cellW      int
	tickN      int
	y1         float64
	height     int
	y0         float64
	x1         float64
	x0         float64
	grid       bool
	titleAlign Align
	xSet       bool
	ySet       bool
	frame      bool
	legend     bool
	orient     Orientation
	showVals   bool
}

func newChartBase() chartBase {
	return chartBase{grid: true, frame: true, legend: true, tickN: defaultTickTarget}
}

// SetTitle sets the diagram's frame title.
func (b *chartBase) SetTitle(t string) { b.title = t }

// ResetTitle clears the diagram's frame title.
func (b *chartBase) ResetTitle() { b.title = "" }

// SetShowValues toggles numeric value annotations (bar tops, heatmap
// cells, pie legend values).
func (b *chartBase) SetShowValues(v bool) { b.showVals = v }

// ResetShowValues disables numeric value annotations.
func (b *chartBase) ResetShowValues() { b.showVals = false }

// SetCellWidth fixes the heatmap cell width in columns; 0 restores the
// automatic stretch that fills the frame width.
func (b *chartBase) SetCellWidth(n int) { b.cellW = n }

// ResetCellWidth restores automatic heatmap cell width.
func (b *chartBase) ResetCellWidth() { b.cellW = 0 }

// SetTitleAlign sets where the diagram title sits within its frame.
func (b *chartBase) SetTitleAlign(a Align) { b.titleAlign = a }

// ResetTitleAlign restores left alignment.
func (b *chartBase) ResetTitleAlign() { b.titleAlign = AlignLeft }

// SetOrientation switches the value axis; see Orientation for semantics.
// Diagrams that cannot swap axes ignore it.
func (b *chartBase) SetOrientation(o Orientation) { b.orient = o }

// ResetOrientation restores each diagram's default axis layout.
func (b *chartBase) ResetOrientation() { b.orient = OrientAuto }

// SetXLabel sets the label text displayed below the X axis.
func (b *chartBase) SetXLabel(l string) { b.xLabel = l }

// SetYLabel sets the label text displayed beside the Y axis.
func (b *chartBase) SetYLabel(l string) { b.yLabel = l }

// ResetLabels clears both axis labels.
func (b *chartBase) ResetLabels() { b.xLabel, b.yLabel = "", "" }

// SetXTicks overrides the auto-generated X-axis ticks with explicit values.
func (b *chartBase) SetXTicks(t []Tick) { b.xTicks = t }

// SetYTicks overrides the auto-generated Y-axis ticks with explicit values.
func (b *chartBase) SetYTicks(t []Tick) { b.yTicks = t }

// ResetTicks clears custom ticks, restoring auto-generation.
func (b *chartBase) ResetTicks() { b.xTicks, b.yTicks = nil, nil }

// SetXFormatter provides a custom formatter for X-axis tick labels.
func (b *chartBase) SetXFormatter(f func(float64) string) { b.xFmt = f }

// SetYFormatter provides a custom formatter for Y-axis tick labels.
func (b *chartBase) SetYFormatter(f func(float64) string) { b.yFmt = f }

// ResetFormatters clears custom formatters, restoring default formatting.
func (b *chartBase) ResetFormatters() { b.xFmt, b.yFmt = nil, nil }

// SetScale fixes both axis ranges, disabling auto-scaling.
func (b *chartBase) SetScale(x0, x1, y0, y1 float64) {
	b.x0, b.x1, b.y0, b.y1 = x0, x1, y0, y1
	b.xSet, b.ySet = true, true
}

// SetXRange fixes the X axis range, disabling auto-scaling on that axis.
func (b *chartBase) SetXRange(lo, hi float64) { b.x0, b.x1, b.xSet = lo, hi, true }

// SetYRange fixes the Y axis range, disabling auto-scaling on that axis.
func (b *chartBase) SetYRange(lo, hi float64) { b.y0, b.y1, b.ySet = lo, hi, true }

// ResetScale re-enables automatic scaling from data.
func (b *chartBase) ResetScale() {
	b.xSet, b.ySet = false, false
	b.xTicks, b.yTicks = nil, nil
}

// SetSize pins this diagram's height in rows when rendered in a Board.
func (b *chartBase) SetSize(rows int) { b.height = rows }

// ResetSize clears the pinned height, letting the Board container decide.
func (b *chartBase) ResetSize() { b.height = 0 }

// HeightHint returns the pinned height or 0 to let the container decide.
func (b *chartBase) HeightHint(int) int {
	if b.height > 0 {
		return b.height
	}

	return 0
}

// SetGrid enables or disables the background grid lines.
func (b *chartBase) SetGrid(on bool) { b.grid = on }

// SetBorder enables or disables the frame border around the diagram.
func (b *chartBase) SetBorder(on bool) { b.frame = on }

// SetLegend enables or disables the legend box.
func (b *chartBase) SetLegend(on bool) { b.legend = on }

// SetTickCount sets the target number of ticks per axis; minimum is 2.
func (b *chartBase) SetTickCount(n int) {
	if n < 2 {
		n = 2
	}

	b.tickN = n
}

// Reset restores every configurable property to its default.
func (b *chartBase) Reset() {
	title := b.title
	*b = newChartBase()
	b.title = title
}

func (b *chartBase) frameTitle(cv *Canvas, uni bool) Rect {
	r := cv.Rect()
	if !b.frame {
		if b.title != "" && r.H > 1 {
			drawAlignedText(
				cv,
				0,
				ellipTrunc(b.title, r.W, uni),
				NewStyle(Default).Bolder(),
				b.titleAlign,
				r.W,
			)

			return Rect{X: 0, Y: 1, W: r.W, H: r.H - 1}.clip(r)
		}

		return r
	}

	st := NewStyle(Gray)
	cv.Border(st, uni)

	if b.title != "" && r.H > 2 {
		t := " " + b.title + " "
		if runeLen(t)+4 > r.W {
			t = " " + ellipTrunc(b.title, max(r.W-6, 1), uni) + " "
		}

		drawAlignedText(cv, 0, t, NewStyle(Default).Bolder(), b.titleAlign, r.W)

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

// plotHeightHint returns the height hint shared by plot-like diagrams
// (Plot, TimeSeries, Candlestick) when no explicit height is set.
func (b *chartBase) plotHeightHint(width int) int {
	if h := b.HeightHint(width); h > 0 {
		return h
	}

	return clampInt(width*2/5, 9, 24)
}

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
