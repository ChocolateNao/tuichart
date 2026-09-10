// Command 08_bullet demonstrates a custom diagram type — a "bullet graph"
// implemented from the public Drawable contract. See docs/extending.md for
// the step-by-step walk-through of how this package is built.
//
// Run it from the repo root:
//
//	go run ./_example/08_bullet
package main

import (
	"fmt"

	"github.com/ChocolateNao/tuichart"
	"github.com/ChocolateNao/tuichart/_example/08_bullet/bullet"
)

func main() {
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

	g := tuichart.New(tuichart.WithWidth(62), tuichart.WithUnicode(true))
	g.Title("service health")
	g.Add(latency)
	g.Row(cpu, orders)
	fmt.Print(g.Render())
}
