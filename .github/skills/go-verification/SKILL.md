---
name: go-verification
description: The canonical verification sequence to run after any Go code change. Use before declaring work complete or opening a PR.
---

# Go verification

Run in this order and stop at the first failure.

```bash
gofmt -s -l .                      # must print nothing
go mod tidy && git diff --exit-code go.mod go.sum
go build ./...
go vet ./...
go test ./... -race -count=1
golangci-lint run                  # if .golangci.yml exists
govulncheck ./...                  # if available
```

## Interpreting failures
- `gofmt -l` output → run `gofmt -s -w .`, commit separately.
- `go mod tidy` diff → the dependency graph was stale; commit the tidied files.
- `vet` `composites`/`copylocks`/`lostcancel` → real bugs, never suppress.
- Race detector report → do not retry; the race is real. Identify the two
  conflicting accesses in the report and fix the synchronisation.
- Flaky test → find the shared state or the timing assumption; do not add sleeps.

## Coverage
```bash
go test ./... -coverprofile=cover.out -covermode=atomic
go tool cover -func=cover.out | tail -1
go tool cover -html=cover.out -o cover.html
```

Never report success without pasting the actual command output.