# 12 — File

Write a rendered chart to a plain-text file.

## What it shows

A line plot rendered with `WithNoColor()` and written to `report.txt`
through `board.RenderTo(f)`. The file carries no ANSI escapes and
renders cleanly in any pager or editor.

## How to run

```sh
cd examples/12_file
go run .
cat report.txt
```

## Caveats

- **Separate module**: run from inside `examples/12_file/`.
- `report.txt` is overwritten on every run.
