# Embedding in other UIs

tuichart renders to cell buffers, so it is straightforward to integrate into
any other text UI, a bubbletea model, a tview application, an HTTP server,
or a custom logging pipeline.

## The key APIs

### RenderCanvas — raw cell access

`RenderCanvas(width)` returns a `*Canvas` and an `Info` describing the
terminal's profile and unicode capability. You can iterate every cell:

```go
cv, info := board.RenderCanvas(width)
for y := 0; y < cv.Height(); y++ {
	for x := 0; x < cv.Width(); x++ {
		c := cv.At(x, y)
		// c.Ch is the rune, c.Style the SGR, c.X/c.Y the position
	}
}
```

Each `Cell` carries a position, a rune, and a `Style` (foreground, background,
bold). You can convert the style to your framework's color model via
`Color.RGB()` (truecolor), or use `Color.Indexed()` for 256-color terminals.

### RenderLines — row-by-row string slicing

`RenderLines(width)` returns `[]string` with one element per row, no trailing
newlines. Use this when your renderer expects an array of horizontal spans:

```go
rows := board.RenderLines(60)
for _, line := range rows {
	widget.AppendLine(line)
}
```

### Live.Frame — render one frame without painting

If you have a [Live renderer](live.md) but want to pull the current frame
without entering the live loop or touching the screen, call
`Live.Frame(width)` — it returns the rendered string and does not invoke the
OnUpdate callback.

### Canvas primitives for custom painting

Once you have a `*Canvas` you can paint freely:

| Method              | Effect                                          |
| ------------------- | ----------------------------------------------- |
| `Set(x, y, ch, st)` | Set one cell.                                    |
| `Text(x, y, s, st)` | Write a string left-to-right.                    |
| `TextRight(x, y, s, st)` | Write right-aligned to the given column.   |
| `TextCenter(x, y, s, st)` | Center the string around the column.       |
| `Sub(Rect)`         | Return a clipped view of the canvas.             |
| `Border(st, uni)`   | Draw a box-drawing or ASCII border around the full canvas. |
| `Clear(st)`         | Fill the entire canvas with a space and a style. |

The `Canvas` is allocated for you by `RenderCanvas` or by `Draw` inside a
custom diagram. When you write your own `Drawable`, the board gives you a
canvas of exactly `HeightHint` rows — you never resize it.

## Embedding patterns

### bubbletea

bubbletea's `View()` returns a `string`. You can render the board once per
tick:

```go
func (m Model) View() string {
	if m.live != nil {
		return m.live.Frame(m.width)
	}
	return m.board.Render(m.width)
}
```

The `INTEGRATION.md` file in the repo root has a full bubbletea example,
including how to intercept `Update` messages to repaint.

### tview

tview widgets have a `Draw` method receiving an `*tcell.EventUpdate`. Use
`RenderCanvas` to build a tcell buffer:

```go
cv, _ := board.RenderCanvas(width)
for y := 0; y < cv.Height(); y++ {
	for x := 0; x < cv.Width(); x++ {
		c := cv.At(x, y)
		style := tcell.StyleDefault.
			Foreground(tcellColor(c.Style.Fg)).
			Bold(c.Style.Bold)
		screen.SetContent(x, y, c.Ch, nil, style)
	}
}
```

### Plain HTTP

Render the board to a string with ANSI escapes, then wrap it in
`<pre><code>`:

```go
w.Header().Set("Content-Type", "text/plain; charset=utf-8")
fmt.Fprint(w, "<pre>")
board.RenderTo(w, 80)
fmt.Fprint(w, "</pre>")
```

## Color conversion

`Color.RGB()` returns the `(r, g, b)` triplet for truecolor terminals. For
256-color terminals use `Color.Indexed()` which returns the 0–255 palette
index. When embedding into a terminal library that only supports 16 colors,
downgrade via the ANSI approximation already built into `Style.emit` — the
same logic your terminal uses. You do not need to implement this yourself;
just render to a string via `Render` and let the host terminal handle the
escape codes, or use `RenderCanvas` and convert `Style.Fg` via
`Color.Indexed()`.

See [`INTEGRATION.md`](../INTEGRATION.md) for the full API surface and
additional examples.