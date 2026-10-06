# Integrating tuichart with TUI frameworks

tuichart renders to strings and cell buffers, which is exactly what every major
Go TUI framework consumes. No adapters or dependencies are required — tuichart
stays pure stdlib and your app pulls in whichever framework it likes. This guide
shows the idiomatic wiring for each.

## The three integration surfaces

| Surface      | API                                                        | Use with                                   |
| ------------ | ---------------------------------------------------------- | ------------------------------------------ |
| ANSI string  | `Board.Render(w)` / `Board.RenderLines(w)`                 | Bubble Tea, any `io.Writer`-based view     |
| Cell buffer  | `Canvas.CellAt` / `Canvas.EachCell`, `Cell{Ch,Fg,Bg,Bold}` | tcell screens, tview primitives            |
| Diagram rects| `Board.RenderLayout(w)` → `([]LayoutEntry, *Canvas, Info)` | chrome/frames around each diagram, hit-testing (see `examples/10_tview`, `11_server`) |
| Concrete RGB | `Color.RGB() (r,g,b int, ok bool)`                         | converting styles to framework color types |

Rules that apply everywhere:

- Let the framework own the screen. Do **not** run `tuichart.Live` inside a TUI
  app's main loop — drive repaints through the framework instead (`Live` is for
  raw-terminal apps).
- Re-render at the current width every frame; frameworks report resizes (Bubble
  Tea sends `tea.WindowSizeMsg`, tcell has `Screen.Size`, gocui has `gui.Size`).
  tuichart never caches geometry between renders.
- If a producer goroutine mutates series data, guard the mutation/render pair
  yourself (or reuse `Ring` + swap values inside the framework's update hook, as
  shown below).

`RenderLayout(w)` returns everything `RenderCanvas(w)` does plus a
`[]LayoutEntry` (one `{Row, X, Y, W, H}` per diagram, absolute in the canvas).
Use the rects to frame diagrams, route clicks, or place chrome without guessing
at board internals. Every example in `examples/09`–`13` exercises it: 09 counts
diagrams in its footer, 10 paints a per-diagram geometry chip in the gap row
above each diagram, 11 exposes the rects as JSON on `/layout`, 12 appends a
manifest to its report file, and 13 shows the live chart geometry in its status
strip.

---

## Bubble Tea (charmbracelet/bubbletea)

Bubble Tea diffs and repaints whatever `View()` returns, so charts are just a
string. Refresh frequency comes from `tea.Tick`.

```go
type model struct {
    g     *tuichart.Board
    line  *tuichart.Line
    ring  *tuichart.Ring
    width int
    tick  bool
}

func initialModel() model {
    g := tuichart.New(tuichart.WithColor256())
    p := tuichart.NewPlot().Title("requests/s")
    line := tuichart.NewLine("reqs")
    p.Add(line)
    g.Add(p)
    return model{g: g, line: line, ring: tuichart.NewRing(80)}
}

func (m model) Init() tea.Cmd { return tickCmd() }

func tickCmd() tea.Cmd {
    return tea.Tick(time.Second, func(time.Time) tea.Msg { return tickMsg{} })
}

type tickMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
    case tickMsg:
        m.ring.Push(sample())            // data pump lives here
        m.line.SetValues(m.ring.Values())
        return m, tickCmd()              // re-arm the ticker
    }
    return m, nil
}

func (m model) View() string {
    w := m.width
    if w == 0 {
        w = 80
    }
    return m.g.Render(w) // multi-line, ANSI-styled — passed through as-is
}
```

Notes:

- Render on every `View()` call; it is cheap (microseconds for typical sizes)
  and keeps the frame in sync with data and window size.
- `RenderLines(w)` returns `[]string` if you assemble views from line slices
  (e.g. `lipgloss.JoinVertical(lipgloss.Left, lines...)`).
- Charts compose with lipgloss borders/padding normally since output is ordinary
  styled text.
- Do not emit alternate-screen sequences from tuichart here; if you want
  full-screen mode, run Bubble Tea with `tea.WithAltScreen()`.

---

## tview (rivo/tview)

Two solid options.

### Option A — TextView + ANSI translation

`tview.TranslateANSI` converts SGR sequences into tview color tags:

```go
g := tuichart.New(tuichart.WithColor256())
tv := tview.NewTextView().SetDynamicColors(true)

go func() {
    for range time.Tick(500 * time.Millisecond) {
        s := g.Render(tvWidth(tv))          // see note below
        app.QueueUpdateDraw(func() {
            tv.SetText(tview.TranslateANSI(s))
        })
    }
}()
```

Track the width you gave the view in your layout (e.g. store it when building
flex/grid items, or fix it with `SetSize`). `QueueUpdateDraw` is required
because tview is single-threaded.

### Option B — custom Primitive (full fidelity, no markup parsing)

Embed a `*tview.Box`, render the diagram into a tuichart canvas, then copy cells
onto the tcell screen. This preserves exact per-cell colors including truecolor
and avoids any tag-escaping pitfalls:

```go
type chartPrimitive struct {
    *tview.Box
    plot *tuichart.Plot
}

func NewChartPrimitive(p *tuichart.Plot) *chartPrimitive {
    return &chartPrimitive{Box: tview.NewBox(), plot: p}
}

func (c *chartPrimitive) Draw(screen tcell.Screen) {
    x, y, w, h := c.GetInnerRect()
    if w < 10 || h < 3 {
        return
    }
    rc := tuichart.NewRenderCtx(tuichart.Info{
        Level:   tuichart.LevelTrueColor,
        Unicode: true,
        W:       w, H: h,
    })
    // Set Unicode: false when the host screen may not be UTF-8 (e.g.
    // tcell with a non-UTF-8 encoding): every diagram then renders pure
    // printable ASCII. User-provided label text still passes through
    // verbatim — transliterate it yourself if needed.
    cv := tuichart.NewCanvas(w, h)
    c.plot.Draw(rc, cv)

    cv.EachCell(func(cx, cy int, cl tuichart.Cell) {
        st := tcell.StyleDefault
        if r, g_, b, ok := cl.Fg.RGB(); ok {
            st = st.Foreground(tcell.NewRGBColor(int32(r), int32(g_), int32(b)))
        }
        if r, g_, b, ok := cl.Bg.RGB(); ok {
            st = st.Background(tcell.NewRGBColor(int32(r), int32(g_), int32(b)))
        }
        if cl.Bold {
            st = st.Bold(true)
        }
        screen.SetContent(x+cx, y+cy, cl.Ch, nil, st)
    })
}
```

`Color.RGB()` resolves both `IndexedColor(n)` (via the xterm palette) and
`RGB(r,g,b)` colors to concrete components, so this works for every style the
library emits. Add the primitive to any tview layout like a built-in widget.

---

## tcell (direct)

The tview Option B loop is the whole story for raw tcell too:

```go
for {
    ev := screen.PollEvent()
    switch ev.(type) {
    case *tcell.EventResize:
        screen.Sync()
        repaint(screen) // re-render + EachCell -> SetContent
    case *tcell.EventKey:
        return
    }
}
```

Drive periodic repaints with a goroutine that calls `screen.Show()` after
copying cells, guarded by `screen.Lock()`-style discipline tcell requires (all
draws on one goroutine, or use `tcell.EventUser` posted to PollEvent).

---

## gocui

gocui views consume plain bytes; they do not support per-cell styling from
written text. Two choices:

1. **Monochrome embed** — render with the no-color profile so shapes carry all
   the information:

```go
g := tuichart.New(tuichart.WithNoColor())

func layout(guiG *gocui.Gui) error {
    maxX, maxY := guiG.Size()
    v, err := guiG.SetView("chart", 0, 0, maxX-1, maxY-1)
    if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
        return err
    }
    v.Clear()
    v.Title = "metrics"
    fmt.Fprint(v, g.Render(maxX))
    return nil
}
```

1. **Colored embed via tcell escape passthrough** — some terminals honor SGR
   sequences written into a gocui view, but this is fragile across redraws;
prefer option 1, or wrap gocui on tcell and use the cell-copy approach
from the tcell section. A complete colored implementation is the
`examples/13_gocui` integration — the board fills a framed, titled gocui
`View` by writing its ANSI output into the view's content buffer, which
gocui's `OutputTrue` escape interpreter parses per cell (`cd examples/13_gocui && go run .`).
Frames, focus, mouse and per-diagram layout (`RenderLayout`) all live in one
example.

Refresh by re-running `layout`'s body from a ticker goroutine through
`gui.Update(func(*gocui.Gui) error {...})`.

---

## termui / other string-based frameworks

Any framework widget that accepts pre-formatted text (termui paragraphs, custom
widgets, log panes) takes `Board.Render(w)` directly. Match the color depth to
the environment: `WithProfile(Level256)` is the safe default for libraries that
don't tell you the terminal capability, `WithTrueColor()` when the host reports
it, `WithNoColor()` when the widget strips escapes.

## Which profile should hosts request?

- Framework knows nothing → `Level256` (safe everywhere modern).
- tcell/tview report truecolor capability → `LevelTrue`.
- Widget strips ANSI → `WithNoColor()`.
