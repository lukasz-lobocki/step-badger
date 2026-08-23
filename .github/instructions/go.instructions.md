---
applyTo: "**/*.go"
description: Idiomatic Go conventions for all Go source files.
---

# Go source conventions

- Group imports in three blocks: stdlib, third-party, local. `goimports` ordering.
- Prefer early returns over nested `if`/`else`; keep the happy path left-aligned.
- Name receivers consistently and briefly (`func (s *Server)`), never `self`/`this`.
- Use `any` over `interface{}` (Go 1.18+).
- Slices: preallocate with `make([]T, 0, n)` when `n` is known.
- Strings: use `strings.Builder` for concatenation in loops.
- Avoid naked returns except in very short functions.
- Struct literals must use field names.
- Guard `nil` maps before write; reads of `nil` maps are fine.
- `defer` for cleanup; check the error of deferred `Close()` when it matters
  (writers, transactions) via a named return.
- Time: always `time.Duration`, never bare ints. Use `time.Since`.
- Prefer `log/slog` for structured logging; no `fmt.Println` in production paths.