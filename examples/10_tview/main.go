// Command 10_tview renders a tuichart chart inside a tview application by
// painting each canvas cell directly onto the tcell screen, framed by the
// same ─── title bar and status footer used by the other integration
// examples.
//
// Run it from this directory (it has its own module so the framework deps
// stay out of the library module):
//
//	go run .
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ChocolateNao/tuichart"
	"github.com/gdamore/tcell/v2"
	tv "github.com/rivo/tview"
)

// chartView is a tview Box that repaints a tuichart canvas on every Draw.
type chartView struct {
	*tv.Box
	board *tuichart.Board
}

func (c *chartView) SetBoard(b *tuichart.Board) *chartView {
	c.board = b
	return c
}

func (c *chartView) Draw(screen tcell.Screen) {
	c.Box.Draw(screen)
	if c.board == nil {
		return
	}
	x, y, w, h := c.GetInnerRect()
	cv, _ := c.board.RenderCanvas(w)

	pal := tcell.StyleDefault
	for i := 0; i < cv.Height() && y+i < h; i++ {
		for j := 0; j < cv.Width() && x+j < w; j++ {
			cell := cv.CellAt(j, i)
			if cell.Ch == 0 {
				continue
			}
			st := pal
			if !cell.Fg.IsZero() {
				r, g, b, _ := cell.Fg.RGB()
				st = st.Foreground(tcell.NewRGBColor(int32(r), int32(g), int32(b)))
			}
			if cell.Bold {
				st = st.Bold(true)
			}
			screen.SetContent(x+j, y+i, cell.Ch, nil, st)
		}
	}
}

// bar mirrors the header/footer filler from the bubbletea and HTTP examples
// so all three share one visual language.
func bar(s string, w int) string {
	if n := utf8.RuneCountInString(s); n < w {
		return s + strings.Repeat("─", w-n)
	}
	runes := []rune(s)
	if len(runes) > w {
		runes = runes[:w]
	}
	return string(runes)
}

func main() {
	plot := tuichart.NewPlot().Title("requests/s")
	line := tuichart.NewLineVals("rps", []float64{})
	plot.Add(line)
	spark := tuichart.NewSpark(0).Title("spark")

	board := tuichart.New(
		tuichart.WithWidth(70),
		tuichart.WithNoColor(),
		tuichart.WithUnicode(true),
	)
	board.Add(plot)
	board.Add(spark)

	app := tv.NewApplication()

	view := tv.NewBox()
	cvw := &chartView{Box: view}
	cvw.SetBoard(board)

	header := tv.NewTextView().
		SetDynamicColors(true).
		SetText("[#808080]" + bar("─── requests/s · tview", 70))
	footer := tv.NewTextView().
		SetDynamicColors(true)

	flex := tv.NewFlex().
		SetDirection(tv.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(cvw, 0, 1, true).
		AddItem(footer, 1, 0, false)
	flex.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'q' {
			app.Stop()
		}
		return ev
	})

	app.SetRoot(flex, true)

	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		var vals []float64
		for range t.C {
			v := math.Sin(float64(time.Now().UnixNano())/4e9)*20 + 30
			vals = append(vals, v)
			if len(vals) > 48 {
				vals = vals[len(vals)-48:]
			}
			line.SetValues(vals)
			spark.SetValues(vals)
			footer.SetText("[#808080]" + bar(
				fmt.Sprintf("─── last %.1f rps · %d samples · q to quit", v, len(vals)), 70))
			app.Draw()
		}
	}()

	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
