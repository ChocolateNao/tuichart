# Styling & degradation

tuichart renders to **cells** — every cell holds a character, a foreground
color, a background color, and a bold flag. The `Style` type bundles these
properties, and every paint method on `Canvas` takes a style as its last
argument.

At the chart level, a *color profile* controls how styles are emitted as
ANSI escape codes — or whether they are emitted at all.

## The color pipeline

```
Style{Fg, Bg, Bold}
  └─► palette lookup (Level) → ANSI SGR sequence in the output string
```

Profiles, in order of fidelity:

| Level          | Name         | When chosen                                                         |
| -------------- | ------------ | ------------------------------------------------------------------- |
| `LevelNone`    | No color     | Explicit `WithNoColor()`, `NO_COLOR` non-empty (no-color.org spec), or terminal explicitly reports no support. |
| `Level16`      | 16 colors    | Minimal terminals, `TERM=dumb`-like.                                 |
| `Level256`     | 256 colors   | Most modern terminals without truecolor support.                     |
| `LevelTrue`    | Truecolor    | Default for terminals advertising 24-bit color.                      |

You can force a profile:

```go
g := tuichart.New(tuichart.WithTrueColor())  // force truecolor
g := tuichart.New(tuichart.WithNoColor())    // force mono (ANSI-free)
```

Detection is cached globally; tests that touch detection should call
`ResetDetection()` afterward.

## Colors

Colors are values of `tuichart.Color`. Three constructors exist:

```go
Indexed(5)   // ANSI 256 palette index
RGB(255, 100, 0) // truecolor triplet
```

There are also named constants for the most common hues:

```go
DodgerBlue, DeepPink, Orange, PaleGreen, MediumPurple, Salmon, SkyBlue,
LightGreen, Khaki, CornflowerBlue, Blue, Cyan, Lime, BrightRed, DimGray, ...
```

## Styles

`tuichart.NewStyle(fg)` creates a `Style` with a foreground color:

```go
s := tuichart.NewStyle(tuichart.Lime)           // just foreground
s = s.On(tuichart.Gray)                  // add background
s = s.Bolder()                           // add bold
```

Every `Canvas` paint method takes a style as its last argument:

```go
cv.Set(x, y, '█', tuichart.NewStyle(tuichart.Lime).Bolder())
cv.Text(5, 0, "hello", tuichart.NewStyle(tuichart.DodgerBlue))
```

When the level is `LevelNone` the SGR emission is empty — the character is
written, but no color or bold attribute follows it. This means a single code
path renders everywhere; you never need to branch on the level yourself.

## Palettes

Diagrams that show multiple data series draw from a shared *palette* — a
`[]Color` embedded in the rendering context. Each new series advances a
palette cursor (`rc.Next()`) that cycles when it runs out.

Default palette (in order):

```
DodgerBlue → DeepPink → Orange → PaleGreen → MediumPurple → Salmon →
SkyBlue → LightGreen → Khaki → CornflowerBlue
```

You can replace it:

```go
g := tuichart.New(tuichart.WithPalette(tuichart.Cyan, tuichart.BrightRed, tuichart.Lime))
```

Or set per-diagram colors explicitly with the `.Color(c)` fluent setter
available on most diagram types.

## Degradation tables

When the terminal lacks truecolor, tuichart approximates the nearest ANSI
256-color index; when it lacks ANSI entirely, it falls back to character
glyphs that are visually distinct even in monochrome.

### Box-drawing

| Unicode  | ASCII  | Use            |
| -------- | ------ | -------------- |
| `┌─…─┐`  | `+-..+` | Frame corners  |
| `│`       | `\|`    | Vertical edges |
| `─`       | `-`     | Horizontal edges |

### Braille (line plots)

Braille dots are absent in pure ASCII. tuichart substitutes slope
characters built from `|`, `/`, `\`, `-`, `.`, `*` so that lines remain
readable.

### Gauge fill

| Style         | Unicode chars             | ASCII fallback       |
| ------------- | ------------------------- | -------------------- |
| `GaugeBlocks` | `████░░░░`                | `####----`           |
| `GaugeSegments`| `▰▰▱▱▱`                 | `==:::`              |
| `GaugeArrow`  | `━━━━╸───`                | `----->---`          |

### Pie slices

| Unicode  | ASCII  |
| -------- | ------ |
| Palette colors | `#`, `@`, `*`, `o` cycling |
| `SliceColor` | same glyph set |

### Heatmap cells

The default gradient maps to a text-density ramp:

| Level      | Characters                          |
| ---------- | ----------------------------------- |
| Truecolor  | Full RGB blend from blue to red.    |
| 256/16     | Approximated ANSI block+fg.         |
| Mono       | `.`, `:`, `-`, `=`, `+`, `*`, `#`, `@` |

## Tips for writing custom diagrams

- Always use `rc.Palette[i]` or `rc.Next()` for series colors; never hard-code
  a color index. That way your diagram adapts to the user's palette.
- Always check `rc.Info.Unicode` when choosing between Unicode glyphs and
  ASCII fallback characters. Example from the tutorial's `bullet` package:

```go
zoneCh, valCh, markCh := '░', '█', '┃'
if !uni {
	zoneCh, valCh, markCh = '.', '#', '|'
}
```

- Use `FormatValue(v)` for tick labels; it rounds and abbreviates
  automatically and respects the available width.

## `NO_COLOR` and the environment

`NO_COLOR` is honored per [no-color.org](https://no-color.org): it must be
a non-empty string (`os.Getenv` returns `""` if unset, so it is ignored when
unset). Detection is global and cached; see `ResetDetection()` to invalidate
the cache between tests.