package main

import (
	"strings"
	"testing"

	"github.com/ChocolateNao/tuichart"
	gocui "github.com/awesome-gocui/gocui"
)

// newSimUI wires the app exactly like main() but on gocui's simulated screen,
// so the whole layout and keybindings are testable without a terminal.
func newSimUI(t *testing.T) (*ui, gocui.TestingScreen, func()) {
	t.Helper()
	gui, err := gocui.NewGui(gocui.OutputSimulator, true)
	if err != nil {
		t.Fatalf("NewGui: %v", err)
	}
	u := &ui{gui: gui, boards: map[string]*tuichart.Board{}}
	gui.Mouse = true
	gui.Highlight = true
	gui.SelFgColor, gui.SelFrameColor = accent, accent
	gui.SetManagerFunc(u.layout)
	u.bind(gui)
	ts := gui.GetTestingScreen()
	cleanup := ts.StartGui()
	t.Cleanup(func() { cleanup(); gui.Close() })
	return u, ts, cleanup
}

func TestLayoutRendersWindows(t *testing.T) {
	u, ts, _ := newSimUI(t)

	chart, err := ts.GetViewContent("chart")
	if err != nil {
		t.Fatalf("chart view: %v", err)
	}
	if !strings.Contains(chart, "requests/s") {
		t.Errorf("chart content missing plot title:\n%s", chart)
	}
	if !strings.ContainsAny(chart, "+┌╭") {
		t.Errorf("chart view should contain its window border:\n%s", chart)
	}
	status, err := ts.GetViewContent("status")
	if err != nil {
		t.Fatalf("status view: %v", err)
	}
	if !strings.Contains(status, "requests/s") || !strings.Contains(status, "cyan") {
		t.Errorf("status strip missing state: %q", status)
	}
	for _, b := range []string{"dataset", "colour", "unicode", "quit"} {
		if _, err := u.gui.View(b); err != nil {
			t.Errorf("button view %q missing: %v", b, err)
		}
	}
}

func TestFocusAndActivation(t *testing.T) {
	u, ts, _ := newSimUI(t)

	if v := u.gui.CurrentView(); v == nil || v.Name() != "chart" {
		t.Fatalf("initial focus = %v, want chart", v)
	}

	ts.SendKeySync(gocui.KeyTab) // -> dataset
	if v := u.gui.CurrentView(); v.Name() != "dataset" {
		t.Fatalf("focus after Tab = %q, want dataset", v.Name())
	}
	bv, _ := u.gui.View("dataset")
	if bv.BgColor != accent {
		t.Errorf("focused button should be filled with accent, got %v", bv.BgColor)
	}

	ts.SendKeySync(gocui.KeyTab) // -> colour
	if v := u.gui.CurrentView(); v.Name() != "colour" {
		t.Fatalf("focus after 2nd Tab = %q, want colour", v.Name())
	}
	bv, _ = u.gui.View("dataset")
	if bv.BgColor != gocui.ColorDefault {
		t.Errorf("unfocused button should reset bg, got %v", bv.BgColor)
	}

	ts.SendKeySync(gocui.KeyEnter) // activate the focused "colour" button
	if u.col != 1 {
		t.Errorf("Enter on colour should cycle colour, got %d", u.col)
	}
}

// pure state logic needs no gui loop, so it is safe to call directly.
func TestStateTransitions(t *testing.T) {
	u := &ui{boards: map[string]*tuichart.Board{}}

	for _, tt := range []int{1, -2, 3} {
		if err := u.nextDataset(tt); err != nil {
			t.Fatalf("nextDataset(%d): %v", tt, err)
		}
	}
	if u.ds != 2 { // (0+1-2+3) wrapped to 2 over 3 datasets
		t.Errorf("dataset after wheel turns = %d, want 2", u.ds)
	}

	if err := u.activate("unicode"); err != nil {
		t.Fatalf("activate unicode: %v", err)
	}
	if !u.uni {
		t.Error("unicode toggle should flip on")
	}
	if err := u.activate("quit"); err != gocui.ErrQuit {
		t.Errorf("activate quit = %v, want ErrQuit", err)
	}
}

func TestUnicodeModeRendersBraille(t *testing.T) {
	// Start with unicode ON so the very first layout runs in braille mode:
	// toggling mid-loop is covered by TestStateTransitions, and reading the
	// simulator mid-loop is not deterministic.
	u := &ui{uni: true, boards: map[string]*tuichart.Board{}}
	gui, err := gocui.NewGui(gocui.OutputSimulator, true)
	if err != nil {
		t.Fatalf("NewGui: %v", err)
	}
	u.gui = gui
	gui.Mouse = true
	gui.Highlight = true
	gui.SelFgColor, gui.SelFrameColor = accent, accent
	gui.SetManagerFunc(u.layout)
	u.bind(gui)
	ts := gui.GetTestingScreen()
	cleanup := ts.StartGui()
	t.Cleanup(func() { cleanup(); gui.Close() })

	chart, err := ts.GetViewContent("chart")
	if err != nil {
		t.Fatalf("chart view: %v", err)
	}
	if !strings.ContainsAny(chart, "⠁⠂⠄⡀⢀⠃") || !strings.Contains(chart, "┌") {
		t.Errorf("unicode chart should use braille/box glyphs:\n%s", chart)
	}
	status, _ := ts.GetViewContent("status")
	if !strings.Contains(status, "unicode true") {
		t.Errorf("status should reflect unicode on: %q", status)
	}
}

func TestGocuiSGSNormalizesBright16Color(t *testing.T) {
	cases := map[string]string{
		"\x1b[96m":         "\x1b[38;5;14m", // bright white -> index 14
		"\x1b[93m":         "\x1b[38;5;11m",
		"\x1b[104m":        "\x1b[48;5;12m",
		"\x1b[32mA\x1b[0m": "\x1b[32mA\x1b[0m", // dark range untouched
		"\x1b[38;2;1;2;3m": "\x1b[38;2;1;2;3m", // truecolor untouched
		"":                 "",
	}
	order := []string{
		"\x1b[96m",
		"\x1b[93m",
		"\x1b[104m",
		"\x1b[32mA\x1b[0m",
		"\x1b[38;2;1;2;3m",
		"",
	}
	for _, in := range order {
		if got := gocuiSGR(in); got != cases[in] {
			t.Errorf("gocuiSGR(%q) = %q, want %q", in, got, cases[in])
		}
	}
}
