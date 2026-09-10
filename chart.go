package tuichart

import (
	"fmt"
	"io"
	"strings"
)

// Options configures a Chart at construction time.
type Options struct {
	palette         []Color
	levelOverride   Level
	width           int
	gap             int
	diagramHeight   int
	hasLevel        bool
	unicodeOverride int8
}

// Option mutates Chart options.
type Option func(*Options)

// WithWidth overrides the detected terminal width with a fixed column count.
func WithWidth(w int) Option {
	return func(o *Options) {
		if w > 0 {
			o.width = w
		}
	}
}

// WithGap sets the number of blank rows between diagram rows.
func WithGap(rows int) Option { return func(o *Options) { o.gap = rows } }

// WithPalette replaces the default color palette with the provided colors.
func WithPalette(cs ...Color) Option { return func(o *Options) { o.palette = cs } }

// WithDiagramHeight sets the default height in rows for diagrams that do not
// provide their own.
func WithDiagramHeight(rows int) Option {
	return func(o *Options) {
		if rows > 2 {
			o.diagramHeight = rows
		}
	}
}

// WithUnicode overrides automatic unicode detection: true forces braille
// and box-drawing glyphs, false forces pure ASCII.
func WithUnicode(on bool) Option {
	return func(o *Options) {
		if on {
			o.unicodeOverride = 1
		} else {
			o.unicodeOverride = 0
		}
	}
}

// WithProfile overrides automatic color-level detection with a fixed Level.
func WithProfile(l Level) Option {
	return func(o *Options) { o.levelOverride, o.hasLevel = l, true }
}

// WithNoColor forces monochrome output with no ANSI escape sequences.
func WithNoColor() Option { return WithProfile(LevelNone) }

// WithColor16 forces 16-color ANSI output.
func WithColor16() Option { return WithProfile(Level16) }

// WithColor256 forces 256-color ANSI output.
func WithColor256() Option { return WithProfile(Level256) }

// WithTrueColor forces 24-bit truecolor ANSI output.
func WithTrueColor() Option { return WithProfile(LevelTrue) }

func (o *Options) apply(info Info) Info {
	if o.hasLevel {
		info.Level = o.levelOverride
	}
	if o.unicodeOverride >= 0 {
		info.Unicode = o.unicodeOverride == 1
	}
	return info
}

type rowEntry struct{ d Drawable }

// Chart renders multiple diagrams into one string, stacked vertically or
// arranged side by side with Row.
type Chart struct {
	title      string
	rows       [][]rowEntry
	opts       Options
	titleAlign Align
}

// New creates an empty chart container. Options configure rendering behavior;
// they override automatic terminal detection.
func New(opts ...Option) *Chart {
	c := &Chart{opts: Options{unicodeOverride: -1}, titleAlign: AlignCenter}
	for _, opt := range opts {
		opt(&c.opts)
	}
	return c
}

// Title sets the container title rendered above all diagrams.
func (c *Chart) Title(t string) *Chart { c.title = t; return c }

// TitleAlign sets where the container title sits: AlignLeft, AlignCenter
// (default) or AlignRight.
func (c *Chart) TitleAlign(a Align) *Chart { c.titleAlign = a; return c }

// Add appends a diagram on its own row.
func (c *Chart) Add(d Drawable) *Chart {
	return c.Row(d)
}

// Row places diagrams side by side, splitting available width evenly.
func (c *Chart) Row(ds ...Drawable) *Chart {
	entry := make([]rowEntry, len(ds))
	for i, d := range ds {
		entry[i] = rowEntry{d: d}
	}
	c.rows = append(c.rows, entry)
	return c
}

// Clear removes all diagrams and the title from the chart.
func (c *Chart) Clear() *Chart { c.rows = nil; return c }

// Len returns the total number of diagrams across all rows.
func (c *Chart) Len() int {
	n := 0
	for _, r := range c.rows {
		n += len(r)
	}
	return n
}

// Reset removes all diagrams but keeps options.
func (c *Chart) Reset() *Chart { c.Clear(); return c }

func defaultDiagramHeight(width int) int {
	h := width / 3
	if h > 20 {
		h = 20
	}
	if h < 9 {
		h = 9
	}
	return h
}

// resolveWidthInfo resolves the effective rendering width and the terminal
// info (after option overrides). A width <= 0 falls back to the detected
// terminal width, then the chart's WithWidth option.
func (c *Chart) resolveWidthInfo(width int) (int, Info) {
	info := c.opts.apply(Detect())
	w := width
	if w <= 0 {
		w = info.W
		if c.opts.width > 0 {
			w = c.opts.width
		}
	}
	if w < 10 {
		w = 10
	}
	return w, info
}

// diagramHeight determines the draw height for a single diagram using the
// diagram's own hint, the chart's default height, and the WithDiagramHeight
// option.
func (c *Chart) diagramHeight(d Drawable, seg, w int) int {
	h := d.HeightHint(seg)
	if h <= 0 {
		h = defaultDiagramHeight(w)
	}
	if c.opts.diagramHeight > 0 && d.HeightHint(seg) == 0 {
		h = c.opts.diagramHeight
	}
	if h < 3 {
		h = 3
	}
	return h
}

// segmentWidth is the per-diagram width when k diagrams share a row, leaving
// a 2-cell gap between neighbors.
func segmentWidth(w, k int) int {
	if k < 1 {
		return 0
	}
	const sepW = 2
	seg := (w - (k-1)*sepW) / k
	if seg < 8 {
		seg = 8
	}
	return seg
}

// layout resolves the effective width and renders every diagram row into
// a slice of finished line strings (ANSI-styled per the resolved profile).
func (c *Chart) layout(width int) []string {
	w, info := c.resolveWidthInfo(width)
	rc := newCtx(info)
	if len(c.opts.palette) > 0 {
		rc.Palette = c.opts.palette
	}

	var lines []string
	if c.title != "" {
		lines = append(lines, alignStyled(c.title, w, c.titleAlign, info.Unicode))
		lines = append(lines, "")
	}

	for ri, row := range c.rows {
		if ri > 0 || len(lines) > 0 && c.title != "" {
			for g := 0; g < max(c.opts.gap, 1); g++ {
				lines = append(lines, "")
			}
		}
		k := len(row)
		seg := segmentWidth(w, k)
		if seg < 8 {
			seg = 8
		}
		heights := make([]int, k)
		canvases := make([]*Canvas, k)
		maxLines := 0
		for i, e := range row {
			h := c.diagramHeight(e.d, seg, w)
			heights[i] = h
			cv := NewCanvas(seg, h)
			e.d.Draw(rc, cv)
			canvases[i] = cv
		}
		rendered := make([][]string, k)
		for i, cv := range canvases {
			s := cv.Render(info.Level)
			s = strings.TrimRight(s, "\n")
			rendered[i] = splitLines(s)
			if len(rendered[i]) > maxLines {
				maxLines = len(rendered[i])
			}
		}
		for ln := 0; ln < maxLines; ln++ {
			var sb []byte
			for i := 0; i < k; i++ {
				if i > 0 {
					sb = append(sb, ' ', ' ')
				}
				if ln < len(rendered[i]) {
					sb = append(sb, rendered[i][ln]...)
				} else {
					sb = append(sb, strings.Repeat(" ", seg+2)...)
				}
			}
			lines = append(lines, strings.TrimRight(string(sb), " "))
		}
	}
	return lines
}

// Render lays out every added diagram. With no argument the detected
// terminal width is used.
func (c *Chart) Render(width ...int) string {
	w := 0
	if len(width) > 0 {
		w = width[0]
	}
	lines := c.layout(w)
	var out []byte
	for i, l := range lines {
		if i > 0 {
			out = append(out, '\n')
		}
		out = append(out, l...)
	}
	out = append(out, '\n')
	return string(out)
}

// RenderLines is the embedding-friendly form of Render: it returns one
// string per terminal row, ANSI-styled according to the resolved profile,
// without trailing newlines. Feed the result to frameworks that manage
// their own line buffers (Bubble Tea views, tview TextView, ...).
func (c *Chart) RenderLines(width ...int) []string {
	w := 0
	if len(width) > 0 {
		w = width[0]
	}
	return c.layout(w)
}

// RenderCanvas renders the chart into a single Canvas preserving per-cell
// style information, plus the Info the render resolved. It mirrors layout
// exactly — same width, title, gaps, and diagram placement — so the result
// serializes to the same output as Render. It exists for incremental
// painters (the Live renderer) that diff frames cell-by-cell instead of
// rewriting the whole screen.
func (c *Chart) RenderCanvas(width int) (*Canvas, Info) {
	w, info := c.resolveWidthInfo(width)
	rc := newCtx(info)
	if len(c.opts.palette) > 0 {
		rc.Palette = c.opts.palette
	}
	const sepW = 2

	totalH := 0
	if c.title != "" {
		totalH = 2
	}
	rowHeights := make([]int, len(c.rows))
	for ri, row := range c.rows {
		if ri > 0 || c.title != "" {
			totalH += max(c.opts.gap, 1)
		}
		seg := segmentWidth(w, len(row))
		mh := 0
		for _, e := range row {
			if h := c.diagramHeight(e.d, seg, w); h > mh {
				mh = h
			}
		}
		rowHeights[ri] = mh
		totalH += mh
	}

	cv := NewCanvas(w, totalH)
	y := 0
	if c.title != "" {
		cx := 0
		switch c.titleAlign {
		case AlignRight:
			cx = w - runeLen(c.title)
		case AlignLeft:
			cx = 0
		default:
			cx = max((w-runeLen(c.title))/2, 0)
		}
		cv.Text(cx, 0, c.title, Style{})
		y = 2
	}
	for ri, row := range c.rows {
		if ri > 0 || c.title != "" {
			y += max(c.opts.gap, 1)
		}
		seg := segmentWidth(w, len(row))
		x := 0
		for _, e := range row {
			h := c.diagramHeight(e.d, seg, w)
			dcv := NewCanvas(seg, h)
			e.d.Draw(rc, dcv)
			cv.Blit(dcv, x, y)
			x += seg + sepW
		}
		y += rowHeights[ri]
	}
	return cv, info
}

// String renders at the detected terminal width.
func (c *Chart) String() string { return c.Render(0) }

// RenderTo writes the rendered chart to any io.Writer — stdout, a file, a
// network connection, an http.ResponseWriter — instead of returning a
// string. The width follows the variadic convention of Render: detected
// terminal width when omitted, explicit otherwise. Styling is decided by
// the chart's own options (WithNoColor / WithProfile), so the same call
// serves ANSI terminals and plain-text sinks like log files.
//
// The chart is rendered fully in memory and then written; a short write
// or any writer error aborts and is returned wrapped with context.
// Rendering itself never fails.
func (c *Chart) RenderTo(w io.Writer, width ...int) error {
	out := c.Render(width...)
	n, err := w.Write([]byte(out))
	if err == nil && n != len(out) {
		err = io.ErrShortWrite
	}
	if err != nil {
		return fmt.Errorf("tuichart: rendering to writer failed: %w", err)
	}
	return nil
}

// WriteTo implements io.WriterTo so charts compose with io.Copy and any
// other writer-based plumbing. It renders at the detected terminal width;
// use RenderTo for an explicit width or wrapped errors.
func (c *Chart) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(c.Render()))
	return int64(n), err
}

// Reader returns an io.Reader over the rendered chart (detected terminal
// width), so charts plug into reader-based APIs: io.Copy, http response
// bodies, multipart writers, and friends. The content is fully rendered
// up front; reading never fails.
func (c *Chart) Reader() io.Reader {
	return strings.NewReader(c.Render())
}

func alignStyled(s string, w int, a Align, uni bool) string {
	n := runeLen(s)
	if n >= w {
		return ellipTrunc(s, w, uni)
	}
	switch a {
	case AlignRight:
		return strings.Repeat(" ", w-n) + s
	case AlignLeft:
		return s
	default:
		pad := (w - n) / 2
		if pad < 0 {
			pad = 0
		}
		return strings.Repeat(" ", pad) + s
	}
}

func splitLines(s string) []string {
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}
