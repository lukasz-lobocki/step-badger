---
applyTo: ".github/workflows/**"
description: CI workflow conventions for Go.
---

# CI conventions

- Pin actions to a commit SHA, with a version comment.
- Set `permissions:` explicitly, least-privilege, at workflow level.
- Use `actions/setup-go` with `go-version-file: go.mod` and caching enabled.
- Required jobs: `lint` (golangci-lint), `test` (`-race -coverprofile`), `build`.
- Add a matrix over the current and previous Go minor versions.
- Fail on `gofmt -l .` producing output.