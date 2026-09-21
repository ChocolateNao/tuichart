// Command 11_server serves a tuichart chart over HTTP as a framed plain-text
// page, written through board.RenderTo into the http.ResponseWriter. The
// framing matches the ─── title bar and status footer of the bubbletea and
// tview integration examples. The per-diagram rectangles from
// board.RenderLayout are exposed as JSON on /layout so a TUI-style client
// can build its own chrome around each diagram.
//
// Run it, then open http://localhost:8080 in a terminal that renders it
// (or curl the URL):
//
//	go run .
package main

import (
	"encoding/json"
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

	board := tuichart.New(tuichart.WithWidth(72), tuichart.WithNoColor())
	board.Add(plot)
	board.Add(tuichart.NewSpark(v...).Title("spark"))

	// Layout is static (the board never changes), so resolve it once and
	// reuse it in every handler.
	entries, _, _ := board.RenderLayout(72)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		head := bar("─── requests/s · http", 72)
		foot := bar(fmt.Sprintf("─── served %d samples · %s · %d diagrams",
			len(v), time.Now().Format("15:04:05"), len(entries)), 72)
		fmt.Fprintln(w, head)
		if err := board.RenderTo(w, 72); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, foot)
	})

	http.HandleFunc("/raw", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if err := board.RenderTo(w, 72); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/layout", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})

	log.Println("listening on :8080 — open http://localhost:8080 (raw chart at /raw, layout JSON at /layout)")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
