# 08 — Custom

A custom diagram type (bullet graph) built from the public API.

## What it shows

A `bullet` package implementing the `Drawable` interface — the same
contract every built-in diagram satisfies. The step-by-step tutorial
for how this package is built lives in
[`docs/extending.md`](../../docs/extending.md).

## How to run

```sh
go run ./examples/08_custom
```

## Caveats

- The sub-package (`bullet/`) lives under `examples/08_custom/` and
  imports the parent module directly.
- Colour degrades gracefully; the bullet zones use distinct named
  colours (`DodgerBlue`, `Lime`, `Orange`) that map cleanly to 16/256
  colour terminals.
