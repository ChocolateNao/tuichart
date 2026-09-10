package tuichart

import (
	"fmt"
	"strings"
)

// paintFrame serializes a new frame for incremental painting. When prev is
// nil or differs in size, the full frame is painted; otherwise only cells
// that actually changed are emitted, which keeps terminal writes small and
// eliminates screen flicker on frames that are entirely unchanged.
//
// termH is the terminal's visible row count when the output is a real TTY
// (0 otherwise). A positive termH selects the terminal-aware full paint:
// the frame is blanked entirely (\x1b[H\x1b[2J) before rendering and no
// trailing newline is emitted, so a resize reflow or a frame shorter than
// the screen can never leave stale cells behind and an exactly-fitting
// frame is never scrolled off by a final newline. The legacy full paint
// (home, render, clear-below) is kept byte-for-byte for non-terminal
// writers (pipes, test buffers) and for the in-memory diff tests.
func paintFrame(prev, cur *Canvas, lvl Level, termH int) string {
	if prev == nil || prev.w != cur.w || prev.h != cur.h {
		if termH > 0 {
			return seqHome + seqClearAll + cur.Render(lvl)
		}
		return seqHome + cur.Render(lvl) + "\n" + seqClearBelow
	}
	return diffPaint(prev, cur, lvl)
}

// diffPaint returns the minimal escape-sequence string that repaints the
// cells changed between prev and cur on the same-size canvas. Rows that are
// unchanged are skipped entirely; within a changed row the cursor jumps over
// runs of identical cells instead of rewriting them.
func diffPaint(prev, cur *Canvas, lvl Level) string {
	var b strings.Builder
	for y := 0; y < cur.h; y++ {
		if rowEqual(prev, cur, y) {
			continue
		}
		fmt.Fprintf(&b, "\x1b[%d;1H", y+1)
		cx := 0
		curStyle := Style{}
		active := false
		for x := 0; x < cur.w; {
			if prev.At(x, y) == cur.At(x, y) {
				x++
				continue
			}
			if cx != x {
				if active {
					b.WriteString(ansiReset)
					active = false
					curStyle = Style{}
				}
				fmt.Fprintf(&b, "\x1b[%dC", x-cx)
				cx = x
			}
			for x < cur.w {
				cl := cur.At(x, y)
				if prev.At(x, y) == cl {
					break
				}
				st := Style{Fg: cl.fg, Bg: cl.bg, Bold: cl.bold}
				if !st.eq(curStyle) {
					if active {
						b.WriteString(ansiReset)
						active = false
					}
					if seq := lvl.seq(st); seq != "" {
						b.WriteString(seq)
						active = true
					}
					curStyle = st
				}
				b.WriteRune(cl.ch)
				x++
				cx++
			}
		}
		if active {
			b.WriteString(ansiReset)
		}
	}
	return b.String()
}

// rowEqual reports whether every cell in row y is identical in both canvases.
func rowEqual(a, b *Canvas, y int) bool {
	for x := 0; x < a.w && x < b.w; x++ {
		if a.At(x, y) != b.At(x, y) {
			return false
		}
	}
	// Unequal widths are treated as different rows.
	return a.w == b.w
}
