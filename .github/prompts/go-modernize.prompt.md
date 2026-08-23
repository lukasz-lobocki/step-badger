---
mode: agent
description: Modernise Go code to current stdlib and language features.
agent: Go Maintainer
---

Modernise `${input:path:./...}`:

- `interface{}` → `any`; type-switch soup → generics where the constraint is obvious.
- Hand-rolled loops → `slices` / `maps` / `cmp` stdlib functions.
- `ioutil.*` → `os` / `io` equivalents.
- `math/rand` → `math/rand/v2`.
- `log` → `log/slog` with structured attributes.
- `for i := 0; i < n; i++` over ranges → `for range n` (Go 1.22+).
- Remove now-unnecessary `x := x` loop-variable shadowing (Go 1.22+).
- Replace mutex-guarded lazy init with `sync.OnceValue`.

Run `go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix ./...`
if available, then review every hunk manually. One category per commit.