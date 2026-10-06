package tuichart

// Ctx carries rendering context passed to Drawables.
type Ctx struct {
	Palette []Color
	Info    Info
	next    int
}

// newCtx creates a new Ctx with the given info and default palette.
func newCtx(info Info) *Ctx {
	pal := defaultPalette
	return &Ctx{Info: info, Palette: pal}
}

// NewRenderCtx builds a Ctx manually, e.g. when rendering a Drawable
// directly into a Canvas without going through a Board.
func NewRenderCtx(info Info) *Ctx { return newCtx(info) }

// WithNextColor returns the next palette color and a new Ctx with the
// cycled index. This is immutable - the original Ctx is not mutated.
func (rc *Ctx) WithNextColor() (Color, *Ctx) {
	c := rc.Palette[rc.next%len(rc.Palette)]
	return c, &Ctx{Palette: rc.Palette, Info: rc.Info, next: rc.next + 1}
}

// LegendEntry is one row of a chart legend.
type LegendEntry struct {
	Label string
	Glyph string
	Style Style
}

// Drawable is implemented by everything that can be placed in a Board.
// Implement it to add custom diagram types; use Canvas primitives inside
// Draw to paint into the area you are given.
type Drawable interface {
	Draw(rc *Ctx, cv *Canvas)
	HeightHint(width int) int
	WidthHint(height int) int
}

// Diagram is a Drawable whose title can be read and written. Every built-in
// diagram satisfies it by embedding DisplayConfig, which is why it is spelled
// GetTitle rather than Title: the fluent setter is Title(string) *T on each
// diagram type, and a Title() string getter on the embedded config would be
// shadowed by it. Board.Diagrams returns them as Diagram.
type Diagram interface {
	Drawable
	GetTitle() string
	SetTitle(string)
}

func drawLegendInside(cv *Canvas, r Rect, entries []LegendEntry, unicode bool) {
	if len(entries) == 0 || r.W < 12 || r.H < 2 {
		return
	}

	const sep = "  "

	x := r.X2()
	y := r.Y

	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]

		glyph := e.Glyph
		if glyph == "" {
			glyph = "██"
			if !unicode {
				glyph = "##"
			}
		}

		w := runeLen(glyph) + 1 + runeLen(e.Label)
		if x-w < r.X {
			x = r.X2()

			y++
			if y > r.Y2() {
				return
			}
		}

		cv.TextRight(x, y, e.Label, e.Style)
		x -= runeLen(e.Label) + 1
		cv.TextRight(x, y, glyph, e.Style)
		x -= runeLen(glyph) + len(sep)
	}
}
