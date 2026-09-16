package tuichart

import (
	"strings"
	"testing"
)

func TestFunnelBasic(t *testing.T) {
	f := NewFunnel().Title("funnel").
		Step("visitors", 100).
		Step("paid", 40).
		ShowValues(true)
	out := renderD(f, WithWidth(60))
	// mono mode (renderD) degrades fills to ASCII chars; find any fill glyph
	var glyph byte
	for _, g := range []byte("#@*o") {
		if strings.IndexByte(out, g) >= 0 {
			glyph = g
			break
		}
	}
	if glyph == 0 {
		t.Fatalf("no filled trapezoid in:\n%s", out)
	}
	for _, want := range []string{"visitors 100 (100%)", "paid 40 (40%)"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// first stage legend line must sit on a non-empty row
	found := 0
	for _, ln := range splitLines(out) {
		if strings.Contains(ln, "visitors 100 (100%)") && strings.IndexByte(ln, glyph) >= 0 {
			found++
		}
	}
	if found == 0 {
		t.Errorf("legend not aligned with trapezoid body")
	}
}

func TestFunnelNoData(t *testing.T) {
	out := renderD(NewFunnel(), WithWidth(30))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("expected no-data placeholder:\n%s", out)
	}
}

func TestFunnelNoDataCases(t *testing.T) {
	for _, name := range []string{"all zero", "all negative", "zero and negative", "too many stages for height"} {
		t.Run(name, func(t *testing.T) {
			out := renderD(newFunnelFor(name), WithWidth(60))
			if !strings.Contains(out, "(no data)") {
				t.Errorf("expected (no data):\n%s", out)
			}
		})
	}
	// Narrow canvas: a 8-wide canvas gives inner.W=6 < 8, the funnel's minimum.
	cv := NewCanvas(8, 10)
	NewFunnel().Step("a", 10).Draw(
		NewRenderCtx(Info{Level: LevelNone, Unicode: true}), cv)
	if !strings.Contains(cv.Plain(), "(no data") {
		t.Errorf("too narrow should be (no data):\n%s", cv.Plain())
	}
}

// newFunnelFor returns a funnel exercising a single no-data trigger.
func newFunnelFor(name string) Drawable {
	switch name {
	case "all zero":
		return NewFunnel().Step("a", 0).Step("b", 0).ShowValues(true)
	case "all negative":
		return NewFunnel().Step("a", -1).Step("b", -2)
	case "zero and negative":
		return NewFunnel().Step("a", 0).Step("b", -5)
	default:
		f := NewFunnel()
		for i := 0; i < 20; i++ {
			f.Step("s", float64(20-i))
		}
		return f
	}
}

func TestFunnelNegativeStageDoesNotPanic(t *testing.T) {
	f := NewFunnel().Title("f").Step("a", 100).Step("b", -5).Step("c", 40).ShowValues(true)
	out := renderD(f)
	if !strings.Contains(out, "b -5") {
		t.Errorf("negative stage should still appear in legend:\n%s", out)
	}
	if strings.Contains(out, "PANIC") {
		t.Errorf("panic leaked:\n%s", out)
	}
}

func TestFunnelSingleStage(t *testing.T) {
	out := renderD(NewFunnel().Step("solo", 42).ShowValues(true))
	if !strings.Contains(out, "solo 42 (100%)") {
		t.Errorf("single stage legend wrong:\n%s", out)
	}
}

func TestFunnelStepColorEmptyNoPanic(t *testing.T) {
	out := renderD(NewFunnel().StepColor(Red))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("empty + StepColor should be no-data:\n%s", out)
	}
}

func TestFunnelStepColorTargetsLatestStage(t *testing.T) {
	// StepColor colors only the stage added most recently; stages added
	// after it must still render with their own palette colors.
	f := NewFunnel().Title("f").
		Step("first", 80).StepColor(Red).
		Step("later", 40).
		ShowValues(true)
	out := renderD(f)
	for _, want := range []string{"first 80", "later 40"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "PANIC") {
		t.Errorf("panic leaked:\n%s", out)
	}
}

func TestFunnelLegendWithoutShowValues(t *testing.T) {
	// without ShowValues the legend shows name + pct only (no raw value).
	out := renderD(NewFunnel().Title("f").Step("a", 100).Step("b", 40))
	for _, want := range []string{"a 100%", "b 40%"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "100 (100%)") {
		t.Errorf("raw value shown without ShowValues:\n%s", out)
	}
}

func TestFunnelPctRelativeToTopStage(t *testing.T) {
	// shares are relative to the first (top) stage, so a later stage can
	// exceed 100%; the funnel width stays proportional to the widest stage.
	out := renderD(NewFunnel().Title("f").
		Step("first", 40).Step("second", 100).ShowValues(true))
	if !strings.Contains(out, "second 100 (250%)") {
		t.Errorf("pct should be relative to first stage:\n%s", out)
	}
}

func TestFunnelFloatValues(t *testing.T) {
	out := renderD(NewFunnel().Title("f").
		Step("a", 2.5).Step("b", 1.25).ShowValues(true))
	for _, want := range []string{"a 2.5 (100%)", "b 1.25 (50%)"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestFunnelHeightHintBounds(t *testing.T) {
	for _, tc := range []struct {
		steps int
		want  int
	}{
		{0, 6}, {1, 6}, {2, 8}, {7, 18}, {20, 18},
	} {
		f := NewFunnel()
		for i := 0; i < tc.steps; i++ {
			f.Step("s", 1)
		}
		if h := f.HeightHint(60); h != tc.want {
			t.Errorf("%d steps: HeightHint=%d, want %d", tc.steps, h, tc.want)
		}
	}
}

func TestFunnelLongNamesTruncatedToGutter(t *testing.T) {
	// legend gutter is capped at inner.W/2; a very long stage name must not
	// overflow the frame or push the plot off-canvas.
	f := NewFunnel().Step(strings.Repeat("x", 80), 10).Step("b", 5)
	out := renderD(f, WithWidth(30))
	if strings.Contains(out, "PANIC") {
		t.Fatalf("panic leaked:\n%s", out)
	}
	for _, ln := range splitLines(out) {
		if runeLen(ln) > 30 {
			t.Errorf("row wider than chart: %q", ln)
		}
	}
}
