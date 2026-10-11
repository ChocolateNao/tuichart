package main

import (
	"fmt"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	fmt.Println("=== Sparkline Braille Demo ===")

	// Sample data
	vals := []float64{1, 3, 2, 5, 4, 7, 6, 9, 8, 10, 9, 11, 10, 12, 11, 13, 12, 14, 13, 15}

	// Default blocks style
	fmt.Println("SparkBlocks (default):")

	s1 := tuichart.NewSpark(vals...).Title("SparkBlocks").Style(tuichart.SparkBlocks)
	b := tuichart.New(tuichart.WithWidth(80),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true)).Add(s1)
	fmt.Println(b.Render())

	// Braille style
	fmt.Println("\nSparkBraille (high vertical resolution):")

	s2 := tuichart.NewSpark(vals...).Title("SparkBraille").Style(tuichart.SparkBraille)
	b = tuichart.New(tuichart.WithWidth(80),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true)).Add(s2)
	fmt.Println(b.Render())

	// With color
	fmt.Println("\nSparkBraille with custom color:")

	s3 := tuichart.NewSpark(
		vals...).Title(
		"SparkBraille - Cyan",
	).Style(
		tuichart.SparkBraille,
	).Color(
		tuichart.Cyan,
	)
	b = tuichart.New(
		tuichart.WithWidth(80),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true),
	).Add(
		s3,
	)
	fmt.Println(b.Render())

	// With negative values
	fmt.Println("\nWith negative values (SparkBraille):")

	vals2 := []float64{-2, 1, -1, 3, 0, 2, -3, 4, 1, 3, -2, 5, 2, 4, 0, 3, -1, 2, -2, 1}
	s4 := tuichart.NewSpark(
		vals2...).Title(
		"SparkBraille - with negatives",
	).Style(
		tuichart.SparkBraille,
	).Color(
		tuichart.HotPink,
	)
	b = tuichart.New(
		tuichart.WithWidth(80),
		tuichart.WithTrueColor(),
		tuichart.WithUnicode(true),
	).Add(
		s4,
	)
	fmt.Println(b.Render())

	// Comparison at different widths
	fmt.Println("\n=== Width Comparison ===")

	for _, w := range []int{40, 60, 80, 120} {
		fmt.Printf("Width %d:\n", w)

		s := tuichart.NewSpark(vals...).Style(tuichart.SparkBraille)
		b = tuichart.New(
			tuichart.WithWidth(w),
			tuichart.WithTrueColor(),
			tuichart.WithUnicode(true),
		).Add(
			s,
		)
		fmt.Println(b.Render())
	}

	// ASCII fallback
	fmt.Println("\nASCII fallback (SparkBraille with Unicode disabled):")

	s5 := tuichart.NewSpark(vals...).Title("SparkBraille ASCII").Style(tuichart.SparkBraille)
	b = tuichart.New(
		tuichart.WithWidth(80),
		tuichart.WithNoColor(),
		tuichart.WithUnicode(false),
	).Add(
		s5,
	)
	fmt.Println(b.Render())

	// Using Spark() convenience function
	fmt.Println("\nUsing Spark() convenience function:")
	fmt.Println(tuichart.Spark([]float64{1, 3, 2, 5, 4, 7, 6, 9, 8, 10}))
	fmt.Println(tuichart.Spark([]float64{1, 3, 2, 5, 4, 7, 6, 9, 8, 10, 9, 11, 10, 12}))
}
