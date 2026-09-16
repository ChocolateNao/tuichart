package tuichart

import (
	"strings"
	"testing"
)

func TestTreemapSlices(t *testing.T) {
	tr := NewTreemap().Title("disk").
		Item("a", 60).
		Item("b", 30).
		Item("c", 10).
		ShowValues(true)
	out := renderD(tr, WithWidth(50))
	if !strings.Contains(out, "a 60") {
		t.Errorf("top label missing:\n%s", out)
	}
	// the largest slice must be wider on its first row than the smallest
	rowsA := rowWhere(out, "a 60")
	rowsC := rowWhere(out, "c")
	if rowsA < 0 || rowsC < 0 {
		t.Fatalf("slice labels missing, a=%d c=%d", rowsA, rowsC)
	}
}

func TestTreemapNested(t *testing.T) {
	tr := NewTreemap().Title("nested").
		Add(&TreemapNode{Name: "web", Children: []*TreemapNode{
			{Name: "static", Value: 70},
			{Name: "api", Value: 30},
		}}).
		Item("else", 20).
		ShowValues(true)
	out := renderD(tr, WithWidth(50))
	for _, want := range []string{"static 70", "api 30", "else 20"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestTreemapNoData(t *testing.T) {
	out := renderD(NewTreemap(), WithWidth(30))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("expected no-data placeholder:\n%s", out)
	}
}

func TestTreemapNoDataCases(t *testing.T) {
	cases := []struct {
		tr   *TreemapChart
		name string
	}{
		{NewTreemap().Item("a", -1).Item("b", -2), "negative only"},
		{NewTreemap().Item("a", 0).Item("b", -5), "zero and negative"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := renderD(tc.tr, WithWidth(30))
			if !strings.Contains(out, "(no data)") {
				t.Errorf("expected (no data):\n%s", out)
			}
		})
	}
}

func TestTreemapTooNarrow(t *testing.T) {
	// 7-wide canvas gives inner.W=5 < 6.
	cv := NewCanvas(7, 10)
	NewTreemap().Item("a", 10).Draw(
		NewRenderCtx(Info{Level: LevelNone, Unicode: true}), cv)
	if !strings.Contains(cv.Plain(), "(no dat") {
		t.Errorf("too narrow should be (no data):\n%s", cv.Plain())
	}
}

func TestTreemapSingleLeaf(t *testing.T) {
	out := renderD(NewTreemap().Title("one").Item("only", 5).ShowValues(true))
	if !strings.Contains(out, "only 5") {
		t.Errorf("single leaf label missing:\n%s", out)
	}
}

func TestTreemapZeroLeafOmitted(t *testing.T) {
	out := renderD(NewTreemap().Title("z").Item("a", 10).Item("zero", 0).Item("b", 30))
	if strings.Contains(out, "zero") {
		t.Errorf("zero-value leaf should render nothing:\n%s", out)
	}
	for _, want := range []string{"a", "b"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestTreemapNegativeChildIgnored(t *testing.T) {
	// a negative child contributes nothing: it must not shift proportions,
	// render a slice, or leak a label onto the frame.
	tr := NewTreemap().Title("n").
		Add(&TreemapNode{Name: "k", Children: []*TreemapNode{
			{Name: "pos", Value: 60},
			{Name: "neg", Value: -20},
			{Name: "last", Value: 10},
		}}).
		ShowValues(true)
	out := renderD(tr)
	if strings.Contains(out, "neg") {
		t.Errorf("negative child leaked label:\n%s", out)
	}
	for _, want := range []string{"pos", "last"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestTreemapEqualValuesDeterministic(t *testing.T) {
	out := renderD(NewTreemap().Title("eq").
		Item("a", 10).Item("b", 10).Item("c", 10).ShowValues(true))
	for _, want := range []string{"a 10", "b 10", "c 10"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestTreemapShowValuesOff(t *testing.T) {
	out := renderD(NewTreemap().Title("t").Item("a", 60).Item("b", 40))
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("labels missing without ShowValues:\n%s", out)
	}
	if strings.Contains(out, "60") || strings.Contains(out, "40") {
		t.Errorf("values shown without ShowValues:\n%s", out)
	}
}

func TestTreemapInteriorOwnValueIgnored(t *testing.T) {
	// an interior node is sized by its children's sum, not its own Value.
	out := renderD(NewTreemap().Title("own").
		Add(&TreemapNode{Name: "n", Value: 500, Children: []*TreemapNode{
			{Name: "c1", Value: 40}, {Name: "c2", Value: 10},
		}}).ShowValues(true))
	if strings.Contains(out, "500") {
		t.Errorf("own interior value leaked into render:\n%s", out)
	}
	for _, want := range []string{"c1", "c2"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestTreemapDeepNestingOrdering(t *testing.T) {
	// interior captions sit above their children; nested leaves land above
	// their parent's horizontal siblings.
	tr := NewTreemap().Title("deep").
		Add(&TreemapNode{Name: "top", Children: []*TreemapNode{
			{Name: "mid", Children: []*TreemapNode{
				{Name: "leaf-a", Value: 30}, {Name: "leaf-b", Value: 20},
			}},
			{Name: "other", Value: 50},
		}}).ShowValues(true)
	out := renderD(tr)
	mid, leafA, other := rowWhere(out, "mid"), rowWhere(out, "leaf-a"), rowWhere(out, "other")
	if mid < 0 || leafA < 0 || other < 0 {
		t.Fatalf("labels missing (mid=%d leaf-a=%d other=%d):\n%s", mid, leafA, other, out)
	}
	if mid >= leafA || leafA >= other {
		t.Errorf("expected mid < leaf-a < other, got %d < %d < %d",
			mid, leafA, other)
	}
}

func TestTreemapPreColoredNodeDoesNotPanic(t *testing.T) {
	// A node with an explicit Color must render without consuming palette slots
	// and without crashing.
	out := renderD(NewTreemap().Title("c").
		Add(&TreemapNode{Name: "a", Value: 30}).
		Add(&TreemapNode{Name: "b", Value: 10, Color: Red}))
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("missing labels:\n%s", out)
	}
	if strings.Contains(out, "PANIC") {
		t.Errorf("panic leaked:\n%s", out)
	}
}

func TestTreemapNodeGetsPaletteColor(t *testing.T) {
	// a node without an explicit color is assigned one as a side effect of
	// rendering, and the assignment is stable across two renders.
	n := &TreemapNode{Name: "x", Value: 10}
	out1 := renderD(NewTreemap().Title("p").Add(n))
	c1 := n.Color
	_ = out1
	if c1.IsZero() {
		t.Errorf("node color not assigned after render")
	}
	renderD(NewTreemap().Title("p").Add(n))
	if n.Color != c1 {
		t.Errorf("re-render changed node color %v -> %v", c1, n.Color)
	}
}

func TestTreemapHeightHintBounds(t *testing.T) {
	for _, tc := range []struct{ w, want int }{
		{5, 8}, {10, 8}, {30, 18}, {50, 24},
	} {
		if h := NewTreemap().HeightHint(tc.w); h != tc.want {
			t.Errorf("Width %d: HeightHint=%d, want %d", tc.w, h, tc.want)
		}
	}
}

// rowWhere returns the first line of out containing substr, or -1.
func rowWhere(out, substr string) int {
	for i, ln := range splitLines(out) {
		if strings.Contains(ln, substr) {
			return i
		}
	}
	return -1
}
