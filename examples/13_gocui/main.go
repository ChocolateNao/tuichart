// Command 13_gocui renders a tuichart board inside a real gocui window. The
// chart lives in a framed, titled View — an actual window instead of bare
// cells painted over the screen. gocui draws the focused window's frame in an
// accent color (and fills the focused footer button), so focus is always
// visible where it sits. Mouse support drives the footer buttons and lets the
// wheel cycle datasets.
//
// Controls:
//
//	Tab / Shift+Tab   move focus (chart window, then each footer button)
//	Enter / Space     activate the focused footer button
//	click             focus the clicked window; clicks on buttons activate them
//	mouse wheel       cycle the dataset (requests/s, memory, cpu)
//	d                 cycle the dataset
//	c                 cycle the line colour (cyan, yellow, green, red, blue, fuchsia)
//	u                 toggle unicode / ascii rendering
//	q or Ctrl+C       quit
//
// The letter keys also work when pressed with Shift (their capitals), since
// gocui delivers those as the uppercase rune.
//
// The board is laid out through board.RenderLayout, so the per-diagram
// rectangle (the chart's on-screen size) is available to the app: the status
// strip shows the live chart geometry from the reported entry.
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/ChocolateNao/tuichart"
	"github.com/awesome-gocui/gocui"
)

var (
	palette = []tuichart.Color{
		tuichart.Cyan,
		tuichart.Yellow,
		tuichart.Green,
		tuichart.Red,
		tuichart.Blue,
		tuichart.Fuchsia,
	}
	palNames = []string{"cyan", "yellow", "green", "red", "blue", "fuchsia"}
	dss      = [][]float64{
		{120, 132, 141, 176, 210, 331, 146, 120, 241, 193, 221, 268, 308, 341, 402, 26},
		{32, 62, 8, 158, 65, 3, 745, 229, 326, 435, 72, 655, 5, 690},
		{28, 61, 72, 9, 85, 326, 402, 736, 749, 81, 100, 130, 900, 620, 385, 41},
	}
	dsNames = []string{"requests/s", "memory", "cpu"}
)

const (
	// chrome colours the non-focused windows; accent marks the focused one,
	// both its frame (via Gui.Highlight) and its button fill.
	chrome = gocui.ColorCyan
	accent = gocui.ColorGreen

	// minX/minY keep every footer button visible (focus order assumes it).
	minX, minY = 52, 12
)

// buttons are the footer controls; the label embeds the quick key.
var buttons = []struct{ name, label string }{
	{"dataset", "d dataset"},
	{"colour", "c colour"},
	{"unicode", "u unicode"},
	{"quit", "q quit"},
}

// ui owns all mutable state. gocui runs every handler (events, mouse and the
// layout manager) on a single goroutine, so no locking is required.
type ui struct {
	gui    *gocui.Gui
	boards map[string]*tuichart.Board
	ds     int
	col    int
	uni    bool
}

// board returns the cached Board for the current selection, building it on
// first use so keypresses swap instantly on the next layout pass.
func (u *ui) board() *tuichart.Board {
	key := fmt.Sprintf("%d:%d:%t", u.ds, u.col, u.uni)
	if b, ok := u.boards[key]; ok {
		return b
	}

	b := tuichart.New(tuichart.WithUnicode(u.uni), tuichart.WithTrueColor())
	plot := tuichart.NewPlot().Title(dsNames[u.ds])
	plot.Add(tuichart.NewLineVals("line", dss[u.ds]).Color(palette[u.col]))
	b.Add(plot)
	u.boards[key] = b

	return b
}

// bar pads s to exactly w cells with the ─/- filler used by the other
// integration examples, so the status strip spans the full window width.
func (u *ui) bar(s string, w int) string {
	r := []rune(s)
	if n := len(r); n > w {
		return string(r[:w])
	}

	fill := "─"
	if !u.uni {
		fill = "-"
	}

	return string(r) + strings.Repeat(fill, w-len(r))
}

func (u *ui) dash(n int) string {
	if u.uni {
		return strings.Repeat("─", n)
	}

	return strings.Repeat("-", n)
}

// gocuiSGR rewrites gocui-hostile SGR in the rendered canvas. tuichart emits
// named colors as bright 16-color codes (`\x1b[96m`, `\x1b[93m`, ...) even in
// truecolor mode, but the escape interpreter only understands the 30-37/40-47
// ranges — 90-97/100-107 fall into its default branch and the color is lost.
// Rewriting them to the 38;5;/48;5; form gocui does parse keeps every palette
// color visible.
//
// note: the implementation presented in this example only remaps bright fg/bg, the only form
// tuichart emits here; a
// combined 1;38;5;N sequence would need splitting, add when a bold line style
// actually merges.
func gocuiSGR(s string) string {
	var b strings.Builder

	for i := 0; i < len(s); {
		if s[i] != '\x1b' || i+1 >= len(s) || s[i+1] != '[' {
			b.WriteByte(s[i])
			i++

			continue
		}

		j := strings.IndexByte(s[i+2:], 'm')

		if j < 0 {
			b.WriteByte(s[i])
			i++

			continue
		}

		parts := strings.Split(s[i+2:i+2+j], ";")

		for k, p := range parts {
			if len(p) == 2 && p[0] == '9' && p[1] >= '0' && p[1] <= '7' {
				parts[k] = "38;5;" + strconv.Itoa(int(p[1]-'0')+8)
			} else if len(p) == 3 && p[0] == '1' && p[1] == '0' && p[2] >= '0' && p[2] <= '7' {
				parts[k] = "48;5;" + strconv.Itoa(int(p[2]-'0')+8)
			}
		}

		b.WriteString("\x1b[")
		b.WriteString(strings.Join(parts, ";"))
		b.WriteString("m")

		i += 2 + j + 1
	}

	return b.String()
}

// paintBoard renders the chart into the gocui view. RenderLayout produces the
// board at the view's interior width and reports each diagram's rectangle;
// Render() emits the canvas as ANSI SGR text that gocui's escape interpreter
// (OutputTrue) parses into per-cell colors, so the board flows through
// gocui's own content buffer like any other view. The entry rects are
// returned so the status strip can show the live chart geometry.
func (u *ui) paintBoard(v *gocui.View) ([]tuichart.LayoutEntry, error) {
	w, h := v.Size()

	if w < 10 || h < 3 {
		return nil, nil
	}

	entries, cv, info := u.board().RenderLayout(w)

	v.Clear()
	v.WriteString(gocuiSGR(cv.Render(info.Level)))

	return entries, nil
}

func (u *ui) layout(gui *gocui.Gui) error {
	// gocui draws its window frames in `-|+` when told to, matching the
	// board's own glyph choice.
	gui.ASCII = !u.uni
	maxX, maxY := gui.Size()

	if maxX < minX || maxY < minY {
		return nil // too cramped to lay out anything useful
	}

	// Chart window: the border and title belong to gocui; tuichart fills the
	// interior through the view's content buffer.
	chart, err := gui.SetView("chart", 0, 0, maxX-1, maxY-6, 0)

	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}

	chart.Title = fmt.Sprintf(" tuichart x gocui · %s ", dsNames[u.ds])
	chart.FrameColor = chrome
	chart.TitleColor = chrome

	entries, paintErr := u.paintBoard(chart)

	if paintErr != nil {
		return paintErr
	}

	// Status strip under the window (the view needs two rows so its single
	// content row lands one line below the chart's frame).
	status, err := gui.SetView("status", 0, maxY-6, maxX-1, maxY-4, 0)

	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}

	status.Frame = false
	status.FgColor = chrome
	status.Clear()
	status.WriteString(u.bar(fmt.Sprintf(
		"%s %s · colour %s · unicode %v · chart %dx%d ",
		u.dash(3),
		dsNames[u.ds],
		palNames[u.col],
		u.uni,
		entries[0].W,
		entries[0].H,
	), maxX-2))

	// Footer buttons: framed views along the bottom edge. The focused button
	// fills with the accent color, so focus is visible on the button itself.
	x := 1

	for _, b := range buttons {
		w := len(b.label) + 2 // one padding cell either side of the label
		if x+w+2 > maxX {
			break
		}

		v, err := gui.SetView(b.name, x, maxY-3, x+w+1, maxY-1, 0)
		if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		if v == gui.CurrentView() {
			v.FgColor, v.BgColor = gocui.ColorBlack, accent
		} else {
			v.FgColor, v.BgColor = gocui.ColorDefault, gocui.ColorDefault
		}

		v.Clear()
		v.WriteString(" " + b.label + " ")

		x += w + 3
	}

	if gui.CurrentView() == nil {
		if _, err := gui.SetCurrentView("chart"); err != nil {
			return err
		}
	}

	return nil
}

// focus moves keyboard focus through the windows in layout order, wrapping at
// both ends.
func (u *ui) focus(dir int) error {
	names := make([]string, 0, 1+len(buttons))

	names = append(names, "chart")
	for _, b := range buttons {
		names = append(names, b.name)
	}

	idx, cur := 0, u.gui.CurrentView()

	if cur != nil {
		for i, n := range names {
			if cur.Name() == n {
				idx = i
				break
			}
		}
	}

	idx = (idx + dir + len(names)) % len(names)
	_, err := u.gui.SetCurrentView(names[idx])

	return err
}

// activate runs the action of a footer button.
func (u *ui) activate(btn string) error {
	switch btn {
	case "dataset":
		u.ds = (u.ds + 1) % len(dss)
	case "colour":
		u.col = (u.col + 1) % len(palette)
	case "unicode":
		u.uni = !u.uni
	case "quit":
		return gocui.ErrQuit
	}

	return nil
}

func (u *ui) nextDataset(dir int) error {
	u.ds = (u.ds + dir + len(dss)) % len(dss)
	return nil
}

// key binds h for both the lowercase and uppercase (Shift+key) forms of r.
// gocui's driver turns Shift+letter into the capitalized rune and drops the
// shift modifier, so mirroring the binding keeps Shift-presses working.
func (u *ui) key(gui *gocui.Gui, r rune, h func(g *gocui.Gui, v *gocui.View) error) {
	_ = gui.SetKeybinding("", r, gocui.ModNone, h)
	if up := unicode.ToUpper(r); up != r {
		_ = gui.SetKeybinding("", up, gocui.ModNone, h)
	}
}

func (u *ui) bind(gui *gocui.Gui) {
	quit := func(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit }

	_ = gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	u.key(gui, 'q', quit)
	_ = gui.SetKeybinding("", gocui.KeyTab, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error { return u.focus(1) })
	_ = gui.SetKeybinding("", gocui.KeyTab, gocui.ModShift,
		func(g *gocui.Gui, v *gocui.View) error { return u.focus(-1) })
	u.key(gui, 'd',
		func(g *gocui.Gui, v *gocui.View) error { return u.nextDataset(1) })
	u.key(gui, 'c',
		func(g *gocui.Gui, v *gocui.View) error { u.col = (u.col + 1) % len(palette); return nil })
	u.key(gui, 'u',
		func(g *gocui.Gui, v *gocui.View) error { u.uni = !u.uni; return nil })

	// A wheel binding carries no rune, so gocui matches it for whatever window
	// is under the cursor.
	_ = gui.SetKeybinding("", gocui.MouseWheelUp, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error { return u.nextDataset(1) })
	_ = gui.SetKeybinding("", gocui.MouseWheelDown, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error { return u.nextDataset(-1) })

	// Clicking the chart window just moves focus to it.
	_ = gui.SetKeybinding("chart", gocui.MouseLeft, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			_, err := gui.SetCurrentView("chart")
			return err
		})

	// Footer buttons: click focuses and activates; Enter/Space activate.
	for _, b := range buttons {
		name := b.name
		act := func(g *gocui.Gui, v *gocui.View) error { return u.activate(name) }
		click := func(g *gocui.Gui, v *gocui.View) error {
			if _, err := gui.SetCurrentView(name); err != nil {
				return err
			}

			return u.activate(name)
		}
		_ = gui.SetKeybinding(name, gocui.MouseLeft, gocui.ModNone, click)
		_ = gui.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, act)
		_ = gui.SetKeybinding(name, gocui.KeySpace, gocui.ModNone, act)
	}
}

func main() {
	u := &ui{boards: map[string]*tuichart.Board{}, uni: true, ds: 0, col: 0}

	gui, err := gocui.NewGui(gocui.OutputTrue, true)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer gui.Close()

	u.gui = gui

	gui.Mouse = true
	gui.Highlight = true
	gui.SelFgColor, gui.SelFrameColor = accent, accent

	gui.SetManagerFunc(u.layout)
	u.bind(gui)

	if err := gui.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
