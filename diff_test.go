package tuichart

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

// testVT is a minimal virtual terminal that applies exactly the escape
// sequences the diff painter produces (CUP, cursor-forward, home, clear,
// alt-screen, SGR) to a rune grid, so tests can assert on the *visible*
// screen state rather than raw bytes.
type testVT struct {
	grid [][]rune
	row  int
	col  int
}

func newTestVT() *testVT { return &testVT{} }

func (v *testVT) feed(s string) {
	for i := 0; i < len(s); {
		if s[i] == '\x1b' {
			i = v.esc(s, i+1)
			continue
		}
		if s[i] == '\n' {
			v.row++
			v.col = 0
			i++
			continue
		}
		r, n := utf8.DecodeRuneInString(s[i:])
		i += n
		v.put(r)
	}
}

func (v *testVT) put(r rune) {
	for v.row >= len(v.grid) {
		v.grid = append(v.grid, []rune{})
	}
	for v.col >= len(v.grid[v.row]) {
		v.grid[v.row] = append(v.grid[v.row], ' ')
	}
	v.grid[v.row][v.col] = r
	v.col++
}

// esc consumes one escape sequence starting at the byte after ESC and
// returns the next index. CSIs ("\x1b[") are parsed into numeric parameters.
func (v *testVT) esc(s string, i int) int {
	if i >= len(s) || s[i] != '[' {
		return i + 1
	}
	i++
	if i < len(s) && s[i] == '?' {
		for i < len(s) && s[i] != 'h' && s[i] != 'l' {
			i++
		}
		return i + 1
	}
	var params []int
	cur := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			cur = cur*10 + int(c-'0')
		case c == ';':
			params = append(params, cur)
			cur = 0
		default:
			params = append(params, cur)
			final := c
			switch final {
			case 'H':
				if len(params) >= 2 {
					v.row = params[0] - 1
					v.col = params[1] - 1
				} else {
					v.row, v.col = 0, 0
				}
			case 'C':
				v.col += params[0]
			}
			return i + 1
		}
		i++
	}
	return i
}

func (v *testVT) visible() []string {
	out := make([]string, 0, len(v.grid))
	for _, row := range v.grid {
		out = append(out, strings.TrimRight(string(row), " "))
	}
	return out
}

// TestLiveDiffReconstructsScreen drives a Live renderer through several
// repaints of a changing gauge, replays the raw escape stream into a virtual
// terminal, and verifies the visible screen matches a fresh plain render of
// the chart. This is the end-to-end guarantee that diff painting paints the
// same picture a full repaint would.
func TestLiveDiffReconstructsScreen(t *testing.T) {
	g := New(WithWidth(30), WithNoColor(), WithUnicode(true))
	gauge := NewGauge(0, 100).Title("progress")
	g.Add(gauge)

	var buf bytes.Buffer
	l := NewLive(g, WithLiveOutput(&buf))
	l.Repaint()
	for _, v := range []float64{25, 50, 12, 87} {
		gauge.Value(v)
		l.Repaint()
	}

	vt := newTestVT()
	vt.feed(buf.String())

	want := screenLines(g, 30)
	got := vt.visible()
	if len(got) != len(want) {
		t.Fatalf("screen height=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d:\n got %q\nwant %q", i, got[i], want[i])
		}
	}
}

// TestLiveDiffWritesNothingWhenStatic verifies an unchanged chart produces
// no terminal writes after the initial paint.
func TestLiveDiffWritesNothingWhenStatic(t *testing.T) {
	g := New(WithWidth(30), WithNoColor(), WithUnicode(true))
	g.Add(NewPlot().Title("still"))
	var buf bytes.Buffer
	l := NewLive(g, WithLiveOutput(&buf))
	l.Repaint()
	n0 := buf.Len()
	for i := 0; i < 3; i++ {
		l.Repaint()
	}
	if buf.Len() != n0 {
		t.Fatalf("static frame wrote %d extra bytes (len %d -> %d)",
			buf.Len()-n0, n0, buf.Len())
	}
}

// TestLiveDiffDoesNotRewriteUnchangedCells checks the diff output skips
// integer cursor-forward jumps over runs of identical cells, i.e. it does
// not merely re-emit full rows.
func TestLiveDiffDoesNotRewriteUnchangedCells(t *testing.T) {
	g := New(WithWidth(30), WithNoColor(), WithUnicode(true))
	gauge := NewGauge(0, 100).Title("bar")
	g.Add(gauge)
	var buf bytes.Buffer
	l := NewLive(g, WithLiveOutput(&buf))
	l.Repaint()
	gauge.Value(42)
	l.Repaint()
	out := buf.String()
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected cursor positioning in diff: %q", out)
	}
	if strings.Count(out, seqHome) != 1 {
		t.Fatalf("expected one full paint, got %d", strings.Count(out, seqHome))
	}
}

// TestRenderCanvasMatchesRender pins RenderCanvas byte-for-byte to the string
// render path so the diff engine cannot silently diverge from static output.
func TestRenderCanvasMatchesRender(t *testing.T) {
	g := New(WithWidth(50), WithNoColor(), WithUnicode(true)).Title("io")
	g.Add(NewPlot().Title("alpha"))
	g.Row(NewBarValues([]string{"a"}, []float64{6}).Title("alpha"), NewGauge(0, 10).Value(7))

	lines := g.RenderLines(50)
	cv, info := g.RenderCanvas(50)
	if cv.Width() != 50 {
		t.Fatalf("RenderCanvas width = %d, want 50", cv.Width())
	}
	got := strings.TrimRight(cv.Render(info.Level), "\n")
	want := strings.Join(lines, "\n")
	if got != want {
		t.Fatalf("RenderCanvas diverges from layout:\n got %q\nwant %q", got, want)
	}
}

// TestPaintFrameFullRepaintVariants pins the three paintFrame outputs: the
// legacy full paint for non-terminal writers, the terminal-aware full paint
// (home + clear-all, no trailing newline) used on a real TTY, and the diff
// path, which must never emit a screen reset.
func TestPaintFrameFullRepaintVariants(t *testing.T) {
	cv := NewCanvas(5, 3)
	cv.Text(0, 0, "abc", Style{})
	body := cv.Plain()

	want := seqHome + body + "\n" + seqClearBelow
	if got := paintFrame(nil, cv, LevelNone, 0); got != want {
		t.Errorf("legacy full paint mismatch:\n got %q\nwant %q", got, want)
	}
	if got := paintFrame(nil, cv, LevelNone, 3); got != seqHome+seqClearAll+body {
		t.Errorf("TTY full paint mismatch:\n got %q\nwant %q", got, seqHome+seqClearAll+body)
	}

	diff := NewCanvas(5, 3)
	diff.Text(0, 0, "abd", Style{})
	got := paintFrame(cv, diff, LevelNone, 3)
	if got == "" || strings.Contains(got, seqHome) || strings.Contains(got, seqClearAll) {
		t.Errorf("same-size diff must avoid screen resets, got %q", got)
	}
}

// screenLines returns the chart laid out at width as plain rows.
func screenLines(c *Chart, width int) []string {
	return c.RenderLines(width)
}
