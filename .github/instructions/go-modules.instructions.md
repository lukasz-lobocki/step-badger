---
applyTo: "**/{go.mod,go.sum,go.work,tools.go}"
description: Dependency and module hygiene.
---

# Module hygiene

- Never hand-edit `go.sum`. Use `go mod tidy`.
- Justify every new direct dependency in the PR description.
- Keep the `go` directive at the minimum version the code actually requires.
- Pin tool dependencies via the `tool` directive (Go 1.24+) instead of `tools.go`.
- After any dependency change run: `go mod tidy && go mod verify && go build ./...`.
- Prefer `golang.org/x/...` over third-party equivalents.