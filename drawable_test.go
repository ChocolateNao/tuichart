package tuichart

import "testing"

func TestCtxNextCycles(t *testing.T) {
	rc := NewRenderCtx(Info{Level: LevelNone})
	n := len(rc.Palette)

	seen := make([]Color, n)
	for i := 0; i < n; i++ {
		c, rc2 := rc.WithNextColor()
		seen[i] = c
		rc = rc2
	}

	wrapped, _ := rc.WithNextColor()
	if wrapped != seen[0] {
		t.Errorf("WithNextColor did not cycle: got %v, want %v", wrapped, seen[0])
	}
}

func TestCtxNextMultipleCycles(t *testing.T) {
	rc := NewRenderCtx(Info{Level: LevelNone})
	n := len(rc.Palette)

	first, rc2 := rc.WithNextColor()

	rc = rc2

	for i := 0; i < n-1; i++ {
		_, rc3 := rc.WithNextColor()

		rc = rc3
	}

	second, _ := rc.WithNextColor()
	if first != second {
		t.Errorf("second cycle start %v != first %v", second, first)
	}
}

func TestCtxWithNextColorDoesNotMutate(t *testing.T) {
	rc := NewRenderCtx(Info{Level: LevelNone})
	before := rc.next

	if _, _ = rc.WithNextColor(); rc.next != before {
		t.Errorf("WithNextColor mutated receiver: next=%d, want %d", rc.next, before)
	}
}

func TestCtxNextEmptyPalette(t *testing.T) {
	rc := NewRenderCtx(Info{})
	if len(rc.Palette) == 0 {
		t.Skip("default palette is empty, skip")
	}

	for i := 0; i < len(rc.Palette)*3; i++ {
		_, rc2 := rc.WithNextColor()
		rc = rc2
	}
}

func TestCtxPaletteDefault(t *testing.T) {
	rc := NewRenderCtx(Info{Level: LevelNone})
	if len(rc.Palette) != len(defaultPalette) {
		t.Errorf("palette len %d, want %d", len(rc.Palette), len(defaultPalette))
	}

	for i, c := range rc.Palette {
		if c != defaultPalette[i] {
			t.Errorf("palette[%d] = %v, want %v", i, c, defaultPalette[i])
		}
	}
}

func TestDiagramSatisfiedByBuiltins(t *testing.T) {
	// Every built-in diagram must satisfy Diagram, or Board.Diagrams silently
	// drops it.
	all := []Drawable{
		NewPlot(), NewBar([]string{"a"}), NewSpark(1, 2), NewGauge(0, 1),
		NewHeat([][]float64{{1}}), NewHistogram([]float64{1}), NewPie(),
		NewTimeline(), NewFunnel(), NewTreemap(), NewCandlestick(),
		NewTimeSeries(), NewFunction(func(float64) float64 { return 0 }),
		NewRadar(), NewGantt(),
	}

	b := New()

	for _, d := range all {
		row, ok := d.(Diagram)
		if !ok {
			t.Errorf("%T does not satisfy Diagram", d)
			continue
		}

		row.SetTitle("t-" + row.GetTitle())
		b.Add(d)
	}

	got := b.Diagrams()
	if len(got) != len(all) {
		t.Fatalf("Diagrams() returned %d, want %d", len(got), len(all))
	}

	for i, d := range got {
		if want := "t-"; len(d.GetTitle()) < len(want) || d.GetTitle()[:len(want)] != want {
			t.Errorf("diagram %d title = %q, want prefix %q", i, d.GetTitle(), want)
		}
	}
}

func TestBoardDiagramsSkipsNonDiagrams(t *testing.T) {
	b := New()
	b.Add(NewPlot().Title("kept"))
	b.Add(&plainDrawable{})

	got := b.Diagrams()
	if len(got) != 1 || got[0].GetTitle() != "kept" {
		t.Fatalf("Diagrams() = %v, want just the titled plot", got)
	}

	if b.Len() != 2 {
		t.Errorf("Len() = %d, want 2", b.Len())
	}
}

// plainDrawable implements Drawable but not Diagram.
type plainDrawable struct{}

func (plainDrawable) Draw(*Ctx, *Canvas) {}
func (plainDrawable) HeightHint(int) int { return 9 }
func (plainDrawable) WidthHint(int) int  { return 0 }
