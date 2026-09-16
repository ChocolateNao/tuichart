package tuichart

import (
	"strings"
	"testing"
)

func TestRadarBasic(t *testing.T) {
	r := NewRadar().Title("radar").
		Axes("a", "b", "c").
		Series("s1", 10, 5, 8).
		Series("s2", 3, 9, 4)
	out := renderD(r)
	for _, want := range []string{"s1", "s2", "b"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestRadarMaxOutOfRangeClamped(t *testing.T) {
	r := NewRadar().Axes("a", "b", "c").
		Series("s", 100, 1, 1).Max(50)
	out := renderD(r, WithWidth(30))
	// must not panic; a value above the pinned max clamps to the outer ring
	if strings.Contains(out, "PANIC") {
		t.Errorf("panic leaked")
	}
}

func TestRadarNoData(t *testing.T) {
	out := renderD(NewRadar().Axes("a", "b"), WithWidth(30)) // <3 axes
	if !strings.Contains(out, "(no data)") {
		t.Errorf("expected no-data placeholder:\n%s", out)
	}
}

func TestRadarNoSeriesNoData(t *testing.T) {
	out := renderD(NewRadar().Axes("a", "b", "c"))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("3 axes but no series should be no-data:\n%s", out)
	}
}

func TestRadarAllZeroSeriesNoData(t *testing.T) {
	out := renderD(NewRadar().Axes("a", "b", "c").Series("s", 0, 0, 0))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("all-zero series should be no-data:\n%s", out)
	}
}

func TestRadarNegativeValuesClamp(t *testing.T) {
	out := renderD(NewRadar().Title("r").Axes("a", "b", "c").
		Series("s", -5, 3, 1))
	if strings.Contains(out, "(no data)") {
		t.Fatalf("mixed negatives must not blank the chart:\n%s", out)
	}
	if strings.Contains(out, "PANIC") {
		t.Errorf("panic leaked:\n%s", out)
	}
}

func TestRadarAutoscaleMax(t *testing.T) {
	// Max(0) or a negative Max let the scale derive from the series values.
	for _, m := range []float64{0, -1} {
		out := renderD(NewRadar().Title("r").Axes("a", "b", "c").
			Series("s", 7, 99, 3).Max(m))
		if strings.Contains(out, "(no data)") {
			t.Errorf("Max(%g) should autoscale:\n%s", m, out)
		}
	}
}

func TestRadarSeriesShorterAndLongerThanAxes(t *testing.T) {
	// shorter: missing axes treated as 0; longer: extras ignored.
	out := renderD(NewRadar().Title("r").Axes("a", "b", "c", "d", "e").
		Series("s", 1, 9, 2))
	if !strings.Contains(out, "s") {
		t.Errorf("series missing from render:\n%s", out)
	}
	out = renderD(NewRadar().Title("r").Axes("a", "b", "c").
		Series("s", 1, 2, 3, 4, 5))
	if strings.Contains(out, "PANIC") {
		t.Errorf("panic leaked for oversized series:\n%s", out)
	}
}

func TestRadarEmptyAxesNoData(t *testing.T) {
	out := renderD(NewRadar().Series("s", 1, 2, 3))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("no axes should be no-data:\n%s", out)
	}
}

func TestRadarTooNarrowNoData(t *testing.T) {
	out := renderD(NewRadar().Axes("a", "b", "c").Series("s", 1, 2, 3), WithWidth(8))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("too narrow should be no-data:\n%s", out)
	}
}

func TestRadarFillAddsBraille(t *testing.T) {
	base := renderD(NewRadar().Title("r").Axes("a", "b", "c").
		Series("s", 3, 5, 4))
	filled := renderD(NewRadar().Title("r").Axes("a", "b", "c").
		Series("s", 3, 5, 4).Fill(true))
	if filled == base {
		t.Errorf("Fill(true) should change the rendered polygon")
	}
	if brailleCount(filled) <= brailleCount(base) {
		t.Errorf("fill should add interior braille dots (base=%d filled=%d)",
			brailleCount(base), brailleCount(filled))
	}
}

func TestRadarFillAsciiHasNoBraille(t *testing.T) {
	out := renderD(NewRadar().Title("rf").Axes("a", "b", "c").
		Series("s", 3, 5, 4).Fill(true), WithUnicode(false))
	if brailleCount(out) != 0 {
		t.Errorf("braille leaked into ascii fill render")
	}
}

func TestRadarLegendGlyphs(t *testing.T) {
	uni := renderD(NewRadar().Title("r").Axes("a", "b", "c").Series("s", 1, 2, 3))
	ascii := renderD(NewRadar().Title("r").Axes("a", "b", "c").Series("s", 1, 2, 3),
		WithUnicode(false))
	if !strings.Contains(uni, "──") {
		t.Errorf("unicode legend should use the long-dash glyph:\n%s", uni)
	}
	if !strings.Contains(ascii, "==") {
		t.Errorf("ascii legend should use the == glyph:\n%s", ascii)
	}
}

func TestRadarManyAxesNoPanic(t *testing.T) {
	names := make([]string, 12)
	vals := make([]float64, 12)
	for i := range names {
		names[i] = "ax"
		vals[i] = float64((i % 5) + 1)
	}
	out := renderD(NewRadar().Title("r").Axes(names...).Series("s", vals...))
	if strings.Contains(out, "(no data)") {
		t.Errorf("12 axes should render:\n%s", out)
	}
	if strings.Contains(out, "PANIC") {
		t.Errorf("panic leaked:\n%s", out)
	}
}

func TestRadarLongAxisLabelTruncated(t *testing.T) {
	out := renderD(NewRadar().Title("r").Axes("very-long-axis-name", "b", "c").
		Series("s", 1, 2, 3), WithWidth(24))
	if strings.Contains(out, "PANIC") {
		t.Fatalf("panic leaked:\n%s", out)
	}
	for _, ln := range splitLines(out) {
		if runeLen(ln) > 24 {
			t.Errorf("row wider than chart: %q", ln)
		}
	}
}

func TestRadarHeightHintBounds(t *testing.T) {
	for _, tc := range []struct{ w, want int }{
		{5, 8}, {10, 9}, {30, 19}, {80, 24},
	} {
		if h := NewRadar().HeightHint(tc.w); h != tc.want {
			t.Errorf("Width %d: HeightHint=%d, want %d", tc.w, h, tc.want)
		}
	}
}

// brailleCount counts U+2800..28FF braille cells in a rendered string.
func brailleCount(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x2800 && r <= 0x28ff {
			n++
		}
	}
	return n
}
