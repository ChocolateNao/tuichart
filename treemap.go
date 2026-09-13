package tuichart

import "sort"

// TreemapNode is one cell of a treemap. Leaf values are sized by Value;
// interior nodes are sized by the sum of their children. An interior node
// may also carry an explicit Color that wins over the palette.
type TreemapNode struct {
	Name     string
	Children []*TreemapNode
	Value    float64
	Color    Color
}

func (n *TreemapNode) value() float64 {
	if len(n.Children) > 0 {
		s := 0.0
		for _, c := range n.Children {
			s += c.value()
		}
		return s
	}
	if n.Value > 0 {
		return n.Value
	}
	return 0
}

// TreemapChart renders a hierarchy of values as nested rectangles whose
// areas are proportional to their summed values.
type TreemapChart struct {
	roots []*TreemapNode
	chartBase
}

// NewTreemap creates an empty treemap.
func NewTreemap() *TreemapChart {
	return &TreemapChart{chartBase: newChartBase()}
}

// Item appends a leaf root node with the given value.
func (t *TreemapChart) Item(name string, val float64) *TreemapChart {
	return t.Add(&TreemapNode{Name: name, Value: val})
}

// Add attaches a root node (which may carry its own children).
func (t *TreemapChart) Add(n *TreemapNode) *TreemapChart {
	t.roots = append(t.roots, n)
	return t
}

// Title sets the chart title.
func (t *TreemapChart) Title(ti string) *TreemapChart { t.SetTitle(ti); return t }

// ShowValues includes each node's value in its label when it fits.
func (t *TreemapChart) ShowValues(v bool) *TreemapChart { t.SetShowValues(v); return t }

// HeightHint returns the suggested height in rows for the given width.
func (t *TreemapChart) HeightHint(width int) int {
	if h := t.chartBase.HeightHint(width); h > 0 {
		return h
	}
	h := width * 3 / 5
	if h > 24 {
		h = 24
	}
	if h < 8 {
		h = 8
	}
	return h
}

type treemapRect struct {
	node *TreemapNode
	x0   int
	y0   int
	x1   int
	y1   int
}

// Draw lays out the tree and fills each leaf rectangle.
func (t *TreemapChart) Draw(rc *Ctx, cv *Canvas) {
	inner := t.frameTitle(cv, rc.Info.Unicode)
	uni := rc.Info.Unicode
	mono := rc.Info.Level == LevelNone

	total := 0.0
	for _, n := range t.roots {
		total += n.value()
	}
	if total <= 0 || len(t.roots) == 0 || inner.W < 6 || inner.H < 4 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", NewStyle(Gray))
		return
	}

	sorted := append([]*TreemapNode(nil), t.roots...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].value() > sorted[j].value()
	})

	var rects []treemapRect
	layoutTreemap(sorted, inner.X, inner.Y, inner.W, inner.H, total, true, &rects)

	fill := '█'
	if mono || !uni {
		fill = pieASCIIChars[0]
	}
	ci := 0
	for i := range rects {
		rt := &rects[i]
		c := rt.node.Color
		if c.IsZero() {
			c = rc.Palette[ci%len(rc.Palette)]
			ci++
			rt.node.Color = c
		}
		if len(rt.node.Children) > 0 {
			continue
		}
		for y := rt.y0; y <= rt.y1; y++ {
			for x := rt.x0; x <= rt.x1; x++ {
				cv.Set(x, y, fill, NewStyle(c))
			}
		}
	}

	// Interior labels first so leaves paint over them, then leaf labels.
	for i := range rects {
		rt := &rects[i]
		n := rt.node
		if len(n.Children) == 0 {
			continue
		}
		if rt.width() >= 3 {
			cv.Text(rt.x0, rt.y0, ellipTrunc(n.Name, rt.width()-1, uni), NewStyle(DimGray))
		}
	}
	for i := range rects {
		rt := &rects[i]
		n := rt.node
		if len(n.Children) > 0 {
			continue
		}
		lbl := n.Name
		if t.showVals && n.value() > 0 {
			lbl = n.Name + " " + FormatValue(n.value())
		}
		w := rt.width()
		if w >= 3 {
			cv.Text(rt.x0, rt.y0, ellipTrunc(lbl, w-1, uni), NewStyle(n.Color))
		}
	}
}

// layoutTreemap splits rect across its children, alternating orientation
// per level, leaving the last child as the remainder. Non-leaf nodes recurse
// so their rects are subdivided by their own children.
func layoutTreemap(
	nodes []*TreemapNode,
	x0, y0, w, h int,
	total float64,
	vertical bool,
	out *[]treemapRect,
) {
	if len(nodes) == 0 || w < 2 || h < 2 {
		return
	}
	curX, curY := x0, y0
	for i, n := range nodes {
		v := n.value()
		if i == len(nodes)-1 {
			rc := treemapRect{node: n, x0: curX, y0: curY, x1: x0 + w - 1, y1: y0 + h - 1}
			*out = append(*out, rc)
			recurseTreemap(n, rc, total, vertical, out)
			break
		}
		var rc treemapRect
		if vertical {
			w2 := int(float64(w) * v / total)
			rc = treemapRect{node: n, x0: curX, y0: curY, x1: curX + w2 - 1, y1: y0 + h - 1}
			curX += w2
		} else {
			h2 := int(float64(h) * v / total)
			rc = treemapRect{node: n, x0: curX, y0: curY, x1: x0 + w - 1, y1: curY + h2 - 1}
			curY += h2
		}
		*out = append(*out, rc)
		recurseTreemap(n, rc, total, vertical, out)
	}
}

func recurseTreemap(
	n *TreemapNode,
	rc treemapRect,
	total float64,
	vertical bool,
	out *[]treemapRect,
) {
	if len(n.Children) == 0 || rc.width() < 2 || rc.height() < 2 {
		return
	}
	kids := append([]*TreemapNode(nil), n.Children...)
	sort.SliceStable(kids, func(i, j int) bool {
		return kids[i].value() > kids[j].value()
	})
	// Reserve the top row for this node's caption when there is room.
	y0, h := rc.y0, rc.height()
	if h >= 3 {
		y0++
		h--
	}
	layoutTreemap(kids, rc.x0, y0, rc.width(), h, n.value(), !vertical, out)
}

func (r *treemapRect) width() int  { return r.x1 - r.x0 + 1 }
func (r *treemapRect) height() int { return r.y1 - r.y0 + 1 }
