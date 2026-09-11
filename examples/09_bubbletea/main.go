// Command 09_bubbletea renders a live tuichart chart inside a bubbletea
// application, framed by a small fixed UI: a title bar above the chart and
// a status footer with the latest sample below it. The chart is written
// through chart.RenderTo into an in-memory buffer (bubbletea's View must
// return a string), so the same io.Writer path serves files, sockets and
// HTTP handlers.
//
// Run it from this directory (it has its own module so the framework deps
// stay out of the library module):
//
//	go run .
package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ChocolateNao/tuichart"
	tea "github.com/charmbracelet/bubbletea"
)

const width = 60

type tickMsg time.Time

type model struct {
	chart *tuichart.Chart
	plot  *tuichart.Plot
	line  *tuichart.Line
	vals  []float64
	width int
	last  float64
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		v := math.Sin(float64(time.Now().UnixNano())/4e9)*20 + 30
		m.last = v
		m.vals = append(m.vals, v)
		if len(m.vals) > 48 {
			m.vals = m.vals[len(m.vals)-48:]
		}
		m.line.SetValues(m.vals)
		return m, tick()
	case tea.WindowSizeMsg:
		m.width = msg.Width - 2
		if m.width < 30 {
			m.width = 30
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View lays out the fixed UI: one title bar, the chart body, then a status
// footer. All bars use the same ─── glyph as the tview and HTTP examples so
// the three integration demos share one visual language.
func (m model) View() string {
	var buf bytes.Buffer
	if err := m.chart.RenderTo(&buf, m.width); err != nil {
		return "tuichart: " + err.Error()
	}
	body := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	head := "─── requests/s · bubbletea"
	foot := fmt.Sprintf("─── last %.1f rps · %d samples · q to quit", m.last, len(m.vals))
	out := bar(head, m.width) + "\n"
	for _, l := range body {
		out += l + "\n"
	}
	out += bar(foot, m.width) + "\n"
	return out
}

// bar writes s and fills the rest of a width-w line with ─ characters,
// truncating s (by runes) if it alone exceeds the width.
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
	m := model{width: width}
	m.plot = tuichart.NewPlot().Title("requests/s")
	gauge := tuichart.NewGauge(65, 100).Title("gauge")
	m.line = tuichart.NewLineVals("rps", []float64{0})
	m.plot.Add(m.line)
	m.chart = tuichart.New(tuichart.WithWidth(width), tuichart.WithNoColor(), tuichart.WithUnicode(true))
	m.chart.Row(m.plot, gauge)
	m.chart.Add(tuichart.NewSpark(0).Title("spark"))

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
