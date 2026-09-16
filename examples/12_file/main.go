// Command 12_file renders a board to a file through board.RenderTo, following
// the same io.Writer path as the other integration examples but with a
// plain-text sink. The board is written with WithNoColor so the file carries
// no ANSI escapes and renders cleanly in any pager or editor.
//
// Run it from this directory (it has its own module so the writer stays out
// of the library module):
//
//	go run .
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	plot := tuichart.NewPlot().Title("requests/s")
	v := make([]float64, 48)
	for i := range v {
		v[i] = math.Sin(float64(i)/5)*40 + 50
	}
	plot.Add(tuichart.NewLineVals("rps", v))

	board := tuichart.New(tuichart.WithWidth(72), tuichart.WithNoColor(), tuichart.WithUnicode(true))
	board.Add(plot)
	board.Add(tuichart.NewSpark(v...).Title("spark"))

	f, err := os.Create("report.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	if err := board.RenderTo(f, 72); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote report.txt (raw chart, no frame)")
}
