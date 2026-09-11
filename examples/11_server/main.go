// Command 11_server serves a tuichart chart over HTTP as a framed plain-text
// page, written through chart.RenderTo into the http.ResponseWriter. The
// framing matches the ─── title bar and status footer of the bubbletea and
// tview integration examples.
//
// Run it, then open http://localhost:8080 in a terminal that renders it
// (or curl the URL):
//
//	go run .
package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ChocolateNao/tuichart"
)

// bar mirrors the header/footer filler from the other integration examples.
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
	v := make([]float64, 60)
	for i := range v {
		v[i] = math.Sin(float64(time.Now().UnixNano()/1e9)/4)*40 + 50
	}
	plot.Add(tuichart.NewLineVals("rps", v))

	chart := tuichart.New(tuichart.WithWidth(72), tuichart.WithNoColor())
	chart.Add(plot)
	chart.Add(tuichart.NewSpark(v...).Title("spark"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		head := bar("─── requests/s · http", 72)
		foot := bar(fmt.Sprintf("─── served %d samples · %s", len(v), time.Now().Format("15:04:05")), 72)
		fmt.Fprintln(w, head)
		if err := chart.RenderTo(w, 72); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, foot)
	})

	http.HandleFunc("/raw", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if err := chart.RenderTo(w, 72); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("listening on :8080 — open http://localhost:8080 (raw chart at /raw)")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
