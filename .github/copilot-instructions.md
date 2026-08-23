# Copilot Instructions — Go Repository

## Project context
This is a Go project. Assume the toolchain declared in `go.mod` is authoritative.
Prefer the standard library; add dependencies only when they earn their weight.

## Golden rules
1. Code must compile: `go build ./...`
2. Code must be vetted: `go vet ./...`
3. Code must be formatted: `gofmt -s -w .` (or `gofumpt` if configured)
4. Tests must pass: `go test ./... -race -count=1`
5. Never commit generated artifacts, secrets, or `*.env` files.

## Style
- Follow [Effective Go](https://go.dev/doc/effective_go) and the
  [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).
- Package names: short, lowercase, no underscores, no `util`/`common`/`helpers`.
- Exported identifiers require doc comments starting with the identifier name.
- Accept interfaces, return structs. Define interfaces in the *consumer* package.
- Keep interfaces small (1–3 methods).
- Zero value should be useful where practical.

## Errors
- Wrap with context: `fmt.Errorf("read config %q: %w", path, err)`.
- Error strings: lowercase, no trailing punctuation, no capitalisation.
- Use `errors.Is` / `errors.As`; never compare error strings.
- Sentinel errors: `var ErrNotFound = errors.New("not found")`.
- Do not `panic` in library code. `log.Fatal` only in `main`.

## Concurrency
- Every goroutine must have a defined exit path.
- Pass `context.Context` as the first parameter, named `ctx`; never store it in a struct.
- Protect shared state with mutexes or channels; always test with `-race`.
- Use `errgroup` for fan-out with error propagation.

## Testing
- Table-driven tests with subtests (`t.Run`).
- Use `t.Cleanup`, `t.Helper`, `t.TempDir`, `t.Context`.
- Prefer real implementations over mocks; fake at package boundaries.
- Name tests `TestXxx_scenario`. Add `Example` functions for public API docs.

## Layout
- `cmd/<binary>/main.go` — thin entrypoints, wiring only.
- `internal/...` — everything not intended for external import.
- `pkg/...` — only if the code is genuinely reusable externally.

## What NOT to do
- Do not introduce a framework or DI container.
- Do not add `init()` functions with side effects.
- Do not use `interface{}`/`any` where a concrete type or generic works.
- Do not rename or move files unless the task explicitly asks for it.
- Do not reformat unrelated code — keep diffs minimal and reviewable.