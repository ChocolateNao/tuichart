package tuichart

import (
	"testing"
)

// BenchmarkCanvasRender benchmarks rendering a canvas per frame.
func BenchmarkCanvasRender(b *testing.B) {
	info := Detect()
	_ = info

	cv := NewCanvas(80, 24)
	for b.Loop() {
		cv.Render(info.Level)
	}
}

// BenchmarkDiffPaint benchmarks the frame diffing per frame.
func BenchmarkDiffPaint(b *testing.B) {
	cv2 := NewCanvas(80, 24)
	for b.Loop() {
		paintFrame(nil, cv2, LevelNone, 0)
	}
}

// BenchmarkBlit benchmarks sub-canvas copy per frame.
func BenchmarkBlit(b *testing.B) {
	src := NewCanvas(40, 12)

	dst := NewCanvas(80, 24)
	for b.Loop() {
		dst.Blit(src, 0, 0)
	}
}

// BenchmarkBoardRender benchmarks full board rendering per frame.
func BenchmarkBoardRender(b *testing.B) {
	board := New()
	board.Add(NewSpark(1, 2, 3, 4, 5))

	for b.Loop() {
		board.Render()
	}
}
