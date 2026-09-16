# 11 — Server

Serve a tuichart chart over HTTP as plain text.

## What it shows

A live chart rendered through `board.RenderTo(w)` into the
`http.ResponseWriter` on every request, framed by the same title/footer
chrome as the other integration examples.

## How to run

```sh
cd examples/11_server
go run .
# open http://localhost:8080 in a terminal, or curl the URL
```

## Caveats

- **Separate module**: run from inside `examples/11_server/`.
- The chart repaints on each request, not on a timer so there is no
  live diff painting. Each visit returns the current state.
- ANSI escapes are included; use a terminal-aware viewer (curl in a
  terminal) or a renderer that understands them. Pipe through `cat -v`
  to see the raw escapes, or set `NO_COLOR=1` if you prefer plain text.
