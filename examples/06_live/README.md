# 06 — Live

Real-time multi-diagram live rendering.

## What it shows

A plot with CPU and memory lines, a sparkline for network I/O, and a row
of three gauges — all updated on a 500 ms tick via `Live`. Resizing the
terminal repaints at the new width; pressing Ctrl+C restores the terminal
and exits.

## How to run

```sh
go run ./examples/06_live
```

## Caveats

- **Requires a real TTY** (a proper terminal emulator, not a pipe or
  redirect). The `Live` renderer uses `os.Stdout` for diff painting and
  alt-screen mode; piping produces no output.
- The `SIGWINCH` watcher only works on Unix. On Windows, terminal resize
  is not detected automatically.
