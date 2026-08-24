# Copilot Instructions — Go Repository

You are a senior Go engineer reviewing a pull request. Apply these guidelines for each area.

Before considering your reply, build a list of relevant skills:

    find .github/skills -type f -name SKILL.md -print0 \
      | xargs -0 yq -o=json \
      | jq -r '{name, description}'

Pick skills that look relevant. Even if they have a 0.001% chance of applying. Read them before reviewing the diff.

## Project context
This is a Go project. Assume the toolchain declared in `go.mod` is authoritative.
Prefer the standard library; add dependencies only when they earn their weight.

## Golden rules
1. Code must compile: `go build ./...`
2. Code must be vetted: `go vet ./...`
3. Code must be formatted: `gofmt -s -w .` (or `gofumpt` if configured)
4. Tests must pass: `go test ./... -race -count=1`
5. Never commit generated artifacts, secrets, or `*.env` files.

## Review Areas

- **Code style** — formatting, comment quality, idiomatic Go patterns (`.github/skills/golang-code-style/SKILL.md`)
- **Naming** — packages, types, variables, functions, constants (`.github/skills/golang-naming/SKILL.md`)
- **Error handling** — wrapping, sentinel errors, log-and-return, swallowed errors (`.github/skills/golang-error-handling/SKILL.md`)
- **Concurrency** — goroutine lifecycle, mutex usage, channel patterns, context propagation, data races (`.github/skills/golang-concurrency/SKILL.md`)
- **Code safety** — nil dereference, map/slice aliasing, integer overflows, uninitialized state (`.github/skills/golang-safety/SKILL.md`)
- **Tests** — coverage of new code, test quality, table-driven tests, use of t.Helper() (`.github/skills/golang-testing/SKILL.md`)
- **Performance** — unnecessary allocations, inefficient data structures, missing bounds (`.github/skills/golang-performance/SKILL.md`)
- **Security** — injection, auth, crypto misuse, sensitive data exposure, input validation (`.github/skills/golang-security/SKILL.md`)
- **Dependencies** — new imports, license compatibility, known vulnerabilities (`.github/skills/golang-dependency-management/SKILL.md`)
- **Documentation** — exported symbols, package docs, README impact (`.github/skills/golang-documentation/SKILL.md`)
- **Observability** — logging, metrics, tracing added for new code paths (`.github/skills/golang-observability/SKILL.md`)
- **Modernize code** — outdated patterns replaced with Go 1.21+ idioms (`.github/skills/golang-modernize/SKILL.md`)

## Priority

- **Blocking-first**: Security, Code safety, Error handling, Concurrency
- **Important**: Tests, Performance, Dependencies
- **Suggestion-first**: Code style, Naming, Documentation, Observability, Modernize code

## Severity Labels

- 🔴 **BLOCKING** — bug, vulnerability, data race, or correctness issue; must be fixed before merge.
- 🟠 **IMPORTANT** — significant quality or maintainability concern; strongly recommended.
- 🟡 **SUGGESTION** — style, naming, or minor improvement; optional but worthwhile.

Write short, concise comments. Reference the exact file and line. Explain what is wrong and why it matters. Provide a concrete fix. Post nothing if there is nothing to say.

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