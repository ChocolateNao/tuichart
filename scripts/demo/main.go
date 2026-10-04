package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/ChocolateNao/tuichart"
	"github.com/ChocolateNao/tuichart/examples/08_custom/bullet"
)

func main() {
	demo := flag.String(
		"demo",
		"all",
		"demo to run: all, quickstart, plot, function, bar, stacked, horizontal, pie, donut, heatmap, gauge, sparkline, timeline, gantt, candlestick, timeseries, histogram, funnel, radar, treemap, ascii, composition, live, bullet, readme-index",
	)
	width := flag.Int("width", 80, "output width")
	noColor := flag.Bool("no-color", false, "disable color")
	noUnicode := flag.Bool("no-unicode", false, "disable unicode")

	flag.Parse()

	opts := []tuichart.Option{tuichart.WithWidth(*width)}
	if *noColor {
		opts = append(opts, tuichart.WithNoColor())
	}

	if *noUnicode {
		opts = append(opts, tuichart.WithUnicode(false))
	}

	switch *demo {
	case "quickstart":
		runQuickstart(opts)
	case "plot":
		runPlot(opts)
	case "function":
		runFunction(opts)
	case "bar":
		runBar(opts)
	case "stacked":
		runStacked(opts)
	case "horizontal":
		runHorizontal(opts)
	case "pie":
		runPie(opts)
	case "donut":
		runDonut(opts)
	case "heatmap":
		runHeatmap(opts)
	case "gauge":
		runGauge(opts)
	case "sparkline":
		runSparkline(opts)
	case "timeline":
		runTimeline(opts)
	case "gantt":
		runGantt(opts)
	case "candlestick":
		runCandlestick(opts)
	case "timeseries":
		runTimeSeries(opts)
	case "histogram":
		runHistogram(opts)
	case "funnel":
		runFunnel(opts)
	case "radar":
		runRadar(opts)
	case "treemap":
		runTreemap(opts)
	case "ascii":
		runASCII(opts)
	case "composition":
		runComposition(opts)
	case "live":
		runLive(opts)
	case "bullet":
		runBullet(opts)
	case "all":
		runAll(opts)
	case "readme-index":
		runReadmeIndex(opts)
	default:
		fmt.Fprintf(os.Stderr, "Unknown demo: %s\n", *demo)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func runQuickstart(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	p := tuichart.NewPlot()

	v := make([]float64, 60)
	for i := range v {
		v[i] = math.Sin(float64(i) / 9)
	}

	p.Add(tuichart.NewLineVals("sin", v))
	g.Add(p)
	fmt.Print(g.Render())
}

func runPlot(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	p := tuichart.NewPlot()

	v := make([]float64, 60)
	for i := range v {
		v[i] = math.Sin(float64(i) / 9)
	}

	p.Add(tuichart.NewLineVals("sin", v))
	g.Add(p)
	fmt.Print(g.Render())
}

func runFunction(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	f := tuichart.NewFunction(math.Cos).Domain(-math.Pi, math.Pi).Samples(60)
	f.SetYRange(-1.2, 1.2)
	g.Add(f.Title("y = cos(x)"))
	fmt.Print(g.Render())
}

func runBar(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	g.Add(tuichart.NewBarValues([]string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"},
		[]float64{120, 200, 150, 80, 70, 110, 130}).
		Title("orders").ShowValues(true))
	fmt.Print(g.Render())
}

func runStacked(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	b := tuichart.NewBarValues([]string{"q1", "q2", "q3", "q4"},
		[]float64{12, 15, 9, 14}).Title("north")
	b.Add(tuichart.BarSeries{Name: "west", Values: []float64{5, 7, 6, 8}})
	b.Stacked(true)
	g.Add(b)
	fmt.Print(g.Render())
}

func runHorizontal(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	b := tuichart.NewBarValues([]string{"orders", "returns", "refunds"},
		[]float64{90, 40, 15}).Title("volume")
	b.Horizontal(true)
	g.Add(b)
	fmt.Print(g.Render())
}

func runPie(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	g.Add(tuichart.NewPie().
		Slice("web", 400).
		Slice("ios", 300).
		Slice("android", 200).
		Slice("cli", 100))
	fmt.Print(g.Render())
}

func runDonut(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	g.Add(tuichart.NewPie().
		Slice("web", 400).
		Slice("ios", 300).
		Slice("android", 200).
		Slice("cli", 100).
		Donut(true))
	fmt.Print(g.Render())
}

func runHeatmap(opts []tuichart.Option) {
	g := tuichart.New(opts...)

	grid := make([][]float64, 5)
	for y := range grid {
		grid[y] = make([]float64, 6)
		for x := range grid[y] {
			grid[y][x] = float64(x*10 + y*4)
		}
	}

	g.Add(tuichart.NewHeat(grid).Colors(tuichart.Blue, tuichart.BrightRed).Title("latency"))
	fmt.Print(g.Render())
}

func runGauge(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	g.Add(tuichart.NewGauge(72, 100).Title("cpu").Label("4 core"))
	g.Add(
		tuichart.NewGauge(
			50,
			100,
		).Style(
			tuichart.GaugeSegments,
		).ShowPercent(
			false,
		).Color(
			tuichart.Lime,
		),
	)
	g.Add(tuichart.NewGauge(38, 100).Style(tuichart.GaugeArrow).Color(tuichart.Cyan))
	fmt.Print(g.Render())
}

func runSparkline(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	g.Add(tuichart.NewSpark(5, 4, 9, 8, 6, 7, 5, 3, 4, 6, 8, 7, 5, 2, 1).Title("packet loss"))
	g.Add(tuichart.NewSpark(0, 1, 2, 3, 4, 5, 6, 7, 8, 7, 6, 5, 4, 3, 2, 1))
	fmt.Print(g.Render())
}

func runTimeline(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	now := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	tl := tuichart.NewTimeline().Format("Jun 02")
	tl.Event(now, "deploy")
	tl.Event(now.Add(6*time.Hour), "canary")
	tl.Events(tuichart.TimelineEvent{
		At:     now.Add(20 * time.Hour),
		Label:  "spike",
		Side:   tuichart.SideBelow,
		Detail: "cpu 98%",
	})
	g.Add(tl)
	fmt.Print(g.Render())
}

func runGantt(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	day := 24 * time.Hour
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	g.Add(tuichart.NewGantt().Format("Jun 02").
		Bar("design", start, start.Add(day*3)).
		Bar("impl", start.Add(day*2), start.Add(day*6)).
		Bar("qa", start.Add(day*5), start.Add(day*8)))
	fmt.Print(g.Render())
}

func runCandlestick(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	base := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	g.Add(tuichart.NewCandlestick().Title("ACME").Format("Jun 02").
		Candle(base, 142.10, 145.80, 141.55, 144.90).
		Candle(base.AddDate(0, 0, 1), 145.00, 146.20, 142.30, 142.60).
		Candle(base.AddDate(0, 0, 2), 142.50, 143.90, 141.20, 143.40).
		Candle(base.AddDate(0, 0, 3), 143.30, 150.10, 143.10, 149.80))
	fmt.Print(g.Render())
}

func runTimeSeries(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	ts := tuichart.NewTimeSeries().Format("15:04").Title("requests")
	base := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)

	line := ts.Line("rps")
	for i, v := range []float64{120, 90, 200, 260, 150, 180} {
		line.Add(base.Add(time.Duration(i)*time.Hour), v)
	}

	g.Add(ts)
	fmt.Print(g.Render())
}

func runHistogram(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	data := []float64{1, 2, 2, 3, 3, 3, 4, 4, 5, 5, 5, 5, 6, 6, 7, 7, 8, 8, 8, 8, 8, 9, 9, 10}
	g.Add(tuichart.NewHistogram(data).Bins(5).Title("distribution"))
	fmt.Print(g.Render())
}

func runFunnel(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	g.Add(tuichart.NewFunnel().
		Title("conversion funnel").
		Step("visitors", 100).
		Step("signups", 55).
		Step("paid", 22).
		ShowValues(true))
	fmt.Print(g.Render())
}

func runRadar(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	g.Add(tuichart.NewRadar().Title("skill profile").
		Axes("frontend", "backend", "data", "testing", "devops", "ux").
		Series("you", 4, 6, 8, 5, 7, 3).
		Series("team", 5, 6, 6, 6, 5, 5))
	fmt.Print(g.Render())
}

func runTreemap(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	g.Add(tuichart.NewTreemap().Title("disk use").
		Add(&tuichart.TreemapNode{Name: "web", Children: []*tuichart.TreemapNode{
			{Name: "static", Value: 40},
			{Name: "logs", Value: 30},
		}}).
		Add(&tuichart.TreemapNode{Name: "mobile", Children: []*tuichart.TreemapNode{
			{Name: "ios", Value: 20},
			{Name: "android", Value: 15},
		}}).
		Item("docs", 10))
	fmt.Print(g.Render())
}

func runASCII(opts []tuichart.Option) {
	asciiOpts := append(opts, tuichart.WithNoColor(), tuichart.WithUnicode(false))
	g := tuichart.New(asciiOpts...)
	p := tuichart.NewPlot().Title("sin")

	v := make([]float64, 60)
	for i := range v {
		v[i] = math.Sin(float64(i) / 9)
	}

	p.Add(tuichart.NewLineVals("sin", v))
	g.Add(p)
	fmt.Print(g.Render())
}

func runComposition(opts []tuichart.Option) {
	g := tuichart.New(opts...)
	p := tuichart.NewPlot().Title("latency")
	p.Add(tuichart.NewLineVals("p50", []float64{30, 28, 31, 29, 33, 30, 32}))
	p.Add(tuichart.NewLineVals("p99", []float64{90, 84, 120, 96, 110, 100, 105}))
	g.Add(p)
	g.Row(
		tuichart.NewGauge(70, 100).Title("cpu"),
		tuichart.NewGauge(45, 100).Title("mem"),
	)
	g.Add(tuichart.NewSpark(3, 5, 2, 6, 4, 7, 5, 2, 1).Title("errors/min"))
	fmt.Print(g.Render())
}

func runLive(opts []tuichart.Option) {
	fmt.Println("This demo requires a real TTY. Run manually with:")
	fmt.Println("  go run ./examples/06_live")
	fmt.Println("Or use the Frame method for static preview:")

	g := tuichart.New(opts...)
	p := tuichart.NewPlot().Title("requests/s")
	g.Add(p)
	l := tuichart.NewLive(g, tuichart.WithInterval(100*time.Millisecond))
	frame := l.Frame(80)
	fmt.Print(frame)
}

func runBullet(opts []tuichart.Option) {
	latency := bullet.NewBullet("latency", 180, 250, 500).
		Title("p95 latency (ms)").
		Zone(350).
		Zone(450).
		Color(tuichart.DodgerBlue)

	cpu := bullet.NewBullet("cpu", 62, 50, 100).
		Zone(80).
		Color(tuichart.Lime)

	orders := bullet.NewBullet("orders", 3800, 4500, 6000).
		Zone(5000).
		Color(tuichart.Orange)

	g := tuichart.New(opts...)
	g.Title("service health")
	g.Add(latency)
	g.Row(cpu, orders)
	fmt.Print(g.Render())
}

func runAll(opts []tuichart.Option) {
	runQuickstart(opts)
	fmt.Println()
	runPlot(opts)
	fmt.Println()
	runFunction(opts)
	fmt.Println()
	runBar(opts)
	fmt.Println()
	runStacked(opts)
	fmt.Println()
	runHorizontal(opts)
	fmt.Println()
	runPie(opts)
	fmt.Println()
	runDonut(opts)
	fmt.Println()
	runHeatmap(opts)
	fmt.Println()
	runGauge(opts)
	fmt.Println()
	runSparkline(opts)
	fmt.Println()
	runTimeline(opts)
	fmt.Println()
	runGantt(opts)
	fmt.Println()
	runCandlestick(opts)
	fmt.Println()
	runTimeSeries(opts)
	fmt.Println()
	runHistogram(opts)
	fmt.Println()
	runFunnel(opts)
	fmt.Println()
	runRadar(opts)
	fmt.Println()
	runTreemap(opts)
	fmt.Println()
	runASCII(opts)
	fmt.Println()
	runComposition(opts)
}

func runReadmeIndex(opts []tuichart.Option) {
	diagrams := []struct {
		fn   func([]tuichart.Option)
		name string
	}{
		{runFunction, "Function"},
		{runBar, "Bar"},
		{runStacked, "Stacked Bar"},
		{runHorizontal, "Horizontal Bar"},
		{runPie, "Pie"},
		{runHeatmap, "Heatmap"},
		{runGauge, "Gauge"},
		{runSparkline, "Sparkline"},
		{runTimeline, "Timeline"},
		{runGantt, "Gantt"},
		{runCandlestick, "Candlestick"},
		{runTimeSeries, "TimeSeries"},
		{runFunnel, "Funnel"},
		{runRadar, "Radar"},
		{runTreemap, "Treemap"},
	}

	for i, d := range diagrams {
		fmt.Printf("\033[H\033[2J")
		fmt.Printf("--- %s (%d/%d) ---\n\n", d.name, i+1, len(diagrams))
		d.fn(opts)

		if i < len(diagrams)-1 {
			time.Sleep(1 * time.Second)
		}
	}
}
