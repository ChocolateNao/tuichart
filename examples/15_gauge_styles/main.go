package main

import (
	"fmt"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	fmt.Println("=== Gauge Styles Demo ===")

	value := 75.0
	max := 100.0

	styles := []struct {
		name  string
		style tuichart.GaugeStyle
	}{
		{"Blocks (default)", tuichart.GaugeBlocks},
		{"ASCII", tuichart.GaugeASCII},
		{"Brackets", tuichart.GaugeBrackets},
		{"Arrow", tuichart.GaugeArrow},
		{"Segments", tuichart.GaugeSegments},
		{"Braille", tuichart.GaugeBraille},
		{"Circles", tuichart.GaugeCircles},
		{"Dots", tuichart.GaugeDots},
		{"Squares", tuichart.GaugeSquares},
	}

	var b *tuichart.Board

	for _, s := range styles {
		g := tuichart.NewGauge(value, max).
			Title(s.name).
			Style(s.style).
			ShowPercent(true)

		b = tuichart.New(tuichart.WithWidth(70),
			tuichart.WithUnicode(true),
			tuichart.WithNoColor()).Add(g)
		fmt.Println(b.Render())
	}

	// With color
	fmt.Println("\n=== With Colors ===")

	g := tuichart.NewGauge(75, 100).
		Title("GaugeBraille - Cyan").
		Style(tuichart.GaugeBraille).
		Color(tuichart.Cyan)
	b = tuichart.New(tuichart.WithWidth(70),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true)).Add(g)
	fmt.Println(b.Render())

	// Gradient examples
	fmt.Println("\n=== Gradient Examples ===")

	g1 := tuichart.NewGauge(80, 100).
		Title("Green to Red Gradient (Braille)").
		Style(tuichart.GaugeBraille).
		Color(tuichart.Green).
		Gradient(true).
		GradientEnd(tuichart.Red)
	b = tuichart.New(tuichart.WithWidth(70),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true)).Add(g1)
	fmt.Println(b.Render())

	g2 := tuichart.NewGauge(60, 100).
		Title("Blue to Fuchsia (Squares)").
		Style(tuichart.GaugeSquares).
		Color(tuichart.Blue).
		Gradient(true).
		GradientEnd(tuichart.Fuchsia)
	b = tuichart.New(
		tuichart.WithWidth(70),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true),
	).Add(
		g2,
	)
	fmt.Println(b.Render())

	g3 := tuichart.NewGauge(90, 100).
		Title("Cyan to Orange (Circles)").
		Style(tuichart.GaugeCircles).
		Color(tuichart.Cyan).
		Gradient(true).
		GradientEnd(tuichart.Orange)
	b = tuichart.New(
		tuichart.WithWidth(70),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true),
	).Add(
		g3,
	)
	fmt.Println(b.Render())

	g4 := tuichart.NewGauge(45, 100).
		Title("Dots with Gradient").
		Style(tuichart.GaugeDots).
		Color(tuichart.Yellow).
		Gradient(true).
		GradientEnd(tuichart.Red)
	b = tuichart.New(
		tuichart.WithWidth(70),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true),
	).Add(
		g4,
	)
	fmt.Println(b.Render())

	// Reverse gradient
	g5 := tuichart.NewGauge(70, 100).
		Title("Red to Green (reverse)").
		Style(tuichart.GaugeBraille).
		Color(tuichart.Red).
		Gradient(true).
		GradientEnd(tuichart.Green)
	b = tuichart.New(
		tuichart.WithWidth(70),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true),
	).Add(
		g5,
	)
	fmt.Println(b.Render())

	// Gradient with Arrow style
	g6 := tuichart.NewGauge(55, 100).
		Title("Arrow with Gradient").
		Style(tuichart.GaugeArrow).
		Color(tuichart.DodgerBlue).
		Gradient(true).
		GradientEnd(tuichart.HotPink)
	b = tuichart.New(
		tuichart.WithWidth(70),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true),
	).Add(
		g6,
	)
	fmt.Println(b.Render())

	// Gradient with Segments
	g7 := tuichart.NewGauge(85, 100).
		Title("Segments with Gradient").
		Style(tuichart.GaugeSegments).
		Color(tuichart.Lime).
		Gradient(true).
		GradientEnd(tuichart.DeepPink)
	b = tuichart.New(
		tuichart.WithWidth(70),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true),
	).Add(
		g7,
	)
	fmt.Println(b.Render())
}
