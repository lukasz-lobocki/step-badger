---
mode: agent
description: Prepare the Go solution for release
agent: Go Reviewer
---

# Context

**Goal:** bring the codebase to a shippable, reproducible, secure, observable and well-documented
release state, then produce everything needed to cut the tag.

**Prime directive:** do not change runtime behaviour unless it fixes a defect you identify and document.
Every behavioural change must be called out explicitly in the PR description with rationale and blast radius.

---

# Phase 0 — Reconnaissance (do this first, report before changing anything)

Produce a short written assessment covering:
1. Repository layout: packages, entrypoints (`cmd/`), internal vs public API surface (`internal/`, `pkg/`).
2. Build system in use: plain `go build`, `Makefile`, `Taskfile`, `mage`, GoReleaser, Bazel.
3. Existing CI: which workflows exist, what they run, which are currently failing.
4. Existing release process: tags, changelog, artifacts published previously.
5. Test posture: number of test files, presence of integration/e2e tests, fixtures, golden files.
6. Delta since last release: `git log [vA.B.C]..HEAD --oneline` and merged PRs — categorise them.
7. A prioritised list of blockers vs. nice-to-haves for this release.

Stop and summarise. Then proceed through the phases below.

---

# Phase 1 — Correctness & code health

## Module hygiene
- `go mod tidy` — `go.mod`/`go.sum` must be clean and committed.
- `go mod verify` passes.
- Confirm the `go` directive matches the true minimum supported version; do not bump it casually
  (it is a breaking change for library consumers).
- Review `replace`, `exclude` and `retract` directives — remove local/dev `replace` lines.
- Audit direct dependencies: for each, note current vs latest version, whether the bump is patch/minor/major,
  whether it is still maintained, and whether it can be dropped in favour of stdlib.
- Apply safe upgrades (`go get -u=patch ./...` then targeted minors). Defer major bumps and list them explicitly.
- Check for duplicate/overlapping dependencies (two YAML libs, two logging libs, etc.).
- Check the dependency tree depth and any unexpectedly heavy transitive deps (`go mod graph`, `go mod why`).

## Static analysis
- `go build ./...` — zero errors, including with `-tags '[buildtags]'` if build tags are used.
- `go vet ./...` — zero findings.
- `gofmt -l .` (or `gofumpt -l .`) — empty output.
- `goimports` grouping consistent (stdlib / external / internal).
- `golangci-lint run` with a committed `.golangci.yml`. Recommended linter set:
  `errcheck, govet, staticcheck, unused, gosimple, ineffassign, revive, gocritic, gosec, bodyclose,
  rowserrcheck, sqlclosecheck, noctx, errorlint, wrapcheck (scoped), copyloopvar, nilerr, misspell,
  unconvert, dupl (high threshold), prealloc, exhaustive (for enums), contextcheck, containedctx`.
- Every remaining `//nolint:...` must have an explanatory comment.
- `go test ./... -run XXX -vet=all` sanity pass.

## Code review sweep
- Error handling: all errors wrapped with `%w` where the caller may need `errors.Is/As`; sentinel errors
  exported where appropriate; no `err != nil { return err }` that loses context; no swallowed errors.
- `context.Context` propagated through all I/O paths; no `context.TODO()` in production paths;
  no context stored in structs; timeouts/deadlines set on every outbound call.
- Concurrency: every goroutine has a clear owner and termination path; no leaked goroutines
  (consider `go.uber.org/goleak` in tests); channels closed by the sender; `sync.WaitGroup`/`errgroup` used correctly;
  no data races; mutexes not copied; `atomic` types used instead of raw ints where relevant.
- Resource management: every `Close()` deferred and its error handled; HTTP response bodies drained and closed;
  file handles, DB rows, tickers (`defer ticker.Stop()`) cleaned up.
- Nil safety: no dereference of possibly-nil pointers/maps/interfaces; typed-nil-in-interface traps checked.
- Slices/maps: no unintended aliasing of caller-owned slices; defensive copies where the API contract implies it.
- Panics: none in library code paths; `recover()` only at well-defined boundaries (HTTP middleware, worker loops)
  and always logged.
- Time: use `time.Time`/`time.Duration` correctly; no wall-clock arithmetic where monotonic is needed;
  clock injectable for tests.
- Determinism: no map-iteration-order dependence in output.
- Remove dead code, unused exports, debug `fmt.Println`, commented-out blocks, stale `TODO`/`FIXME`
  (either fix, or convert to a tracked issue and reference it).
- Ensure `internal/` is used to keep non-public helpers out of the public API.

---

# Phase 2 — Testing

- `go test ./... -race -count=1` passes cleanly.
- `go test ./... -count=5` to catch flakiness; also run with `-shuffle=on`.
- Coverage: `go test ./... -coverprofile=coverage.out -covermode=atomic`; report total and per-package coverage.
  Add tests for uncovered exported behaviour that is release-critical. State a coverage floor and enforce it in CI.
- Table-driven tests for public API edge cases: empty input, nil, zero values, max sizes,
  unicode/multibyte strings, negative numbers, timezone boundaries, cancelled contexts.
- Golden-file tests for any rendered output (CLI help text, templates, serialised formats) — with an `-update` flag.
- Fuzz targets (`FuzzXxx`) for parsers, decoders, and anything handling untrusted input; run
  `go test -fuzz=Fuzz -fuzztime=[60s]` and commit any crashers found in `testdata/fuzz`.
- Benchmarks (`BenchmarkXxx`) for hot paths; compare against the previous release with `benchstat`
  and report regressions >[5]%.
- Integration tests gated behind a build tag or `testing.Short()`; document how to run them.
- Tests must be hermetic: no real network, no reliance on host timezone/locale, no `time.Sleep` for synchronisation,
  temp dirs via `t.TempDir()`, parallel-safe (`t.Parallel()` where sound).
- `go test ./... -race` under the oldest and newest supported Go version.
- Example tests (`ExampleXxx`) for the main public API — these double as documentation on pkg.go.dev.

---

# Phase 3 — Build, versioning & reproducibility

- Version information embedded at build time:
  ```
  -ldflags "-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}"
  ```
  or derived from `runtime/debug.ReadBuildInfo()` (preferred — works with `go install`).
- A `--version` / `version` subcommand printing version, commit, build date, Go version, OS/arch.
- Reproducible builds: `CGO_ENABLED=0` where possible, `-trimpath`, pinned toolchain, `SOURCE_DATE_EPOCH` honoured.
- Cross-compilation verified for every target platform in the matrix.
- Static vs dynamic linking decision documented (especially for CGO / `net` / `os/user` resolver behaviour).
- Binary size checked and reported; strip symbols in release builds.
- `.goreleaser.yaml`:
  - builds matrix, `ignore` entries for unsupported combos
  - archives (`tar.gz` for unix, `zip` for windows) with README/LICENSE included
  - `checksums.txt` with SHA-256
  - changelog generation from conventional commits, with `groups` and `filters.exclude`
  - `release.draft: true` for review before publish, `prerelease: auto`
  - optional: Homebrew tap, Scoop bucket, nfpm (deb/rpm/apk), Docker manifests, SBOM, signing
  - validate with `goreleaser check` and `goreleaser release --snapshot --clean`, then inspect `dist/`.
- Container image (if applicable): multi-stage build, `distroless`/`alpine`/`scratch` base, non-root `USER`,
  `HEALTHCHECK` where meaningful, OCI labels (`org.opencontainers.image.*`), multi-arch manifest,
  `.dockerignore` present, image scanned with Trivy/Grype.
- `Makefile`/`Taskfile` targets: `build`, `test`, `lint`, `fmt`, `cover`, `snapshot`, `release`, `clean`, `tools`.
- Tool dependencies pinned via `tools.go` + `go.mod`, or a `tools/go.mod`, so contributors get identical versions.

---

# Phase 4 — CI/CD

- **PR workflow** (`.github/workflows/ci.yml`): matrix over `[os] × [go-version]`; steps: checkout,
  setup-go with module+build cache, `go mod download`, build, vet, `golangci-lint`, `go test -race -coverprofile`,
  upload coverage. Concurrency group cancels superseded runs. Least-privilege `permissions:` block.
- **Release workflow** (`.github/workflows/release.yml`): triggered on `push: tags: ['v*']`;
  `permissions: contents: write, packages: write, id-token: write`; full `fetch-depth: 0` for changelog;
  runs GoReleaser; publishes artifacts, checksums, SBOM, signatures and container images.
- **Security workflows**: CodeQL (Go), `govulncheck`, dependency review on PRs, secret scanning enabled.
- **Dependabot** (`.github/dependabot.yml`) for `gomod`, `github-actions`, and `docker`, grouped updates.
- All GitHub Actions pinned to a full commit SHA, not a floating tag.
- Branch protection: required status checks, required review, linear history, no force-push to `main`.
- Optional: release-please or changesets if you want automated version bumps + changelog PRs.
- Verify the workflows actually pass — do not hand back untested YAML.

---

# Phase 5 — Security & supply chain

- `govulncheck ./...` — every finding either fixed or documented with justification and tracking issue.
- `gosec ./...` — review findings, especially file permissions, command execution, TLS config, weak crypto, path traversal.
- SBOM generated (CycloneDX or SPDX via Syft/GoReleaser) and attached to the release.
- Artifact signing: `cosign sign-blob` / keyless OIDC signing for binaries and container images;
  document the verification command for users.
- SLSA provenance attestation if targeting a higher assurance level.
- Secrets audit: `gitleaks detect` over full history; no tokens, keys, internal hostnames, or customer data committed.
- Input validation on all external boundaries; size limits on request bodies and file reads; no unbounded allocations.
- TLS: minimum version 1.2+, no `InsecureSkipVerify` outside explicitly opt-in debug flags.
- Dependency licence audit (`go-licenses report`) — confirm all licences are compatible with the project's licence;
  produce a `THIRD_PARTY_NOTICES` file if required.
- `LICENSE` present and correct; SPDX headers consistent if the project uses them.
- `SECURITY.md` with a vulnerability disclosure policy and contact.

---

# Phase 6 — Operability (for services/daemons)

- Structured logging (`log/slog`), configurable level, no secrets logged, consistent field names.
- Metrics (Prometheus) and/or OpenTelemetry traces on key paths; documented metric names and labels.
- Health/readiness endpoints.
- Graceful shutdown on `SIGINT`/`SIGTERM` with a bounded drain timeout; in-flight work completed or checkpointed.
- Configuration precedence documented (flags > env > file > defaults); config validated at startup with clear errors.
- Sensible resource limits and backpressure; retries with exponential backoff and jitter; circuit breaking where relevant.
- Exit codes documented and meaningful.

---

# Phase 7 — Documentation

- `README.md`: badges, one-paragraph description, feature list, installation for every distribution channel,
  quick start with copy-pasteable commands, full flags/env/config reference table, worked examples,
  troubleshooting/FAQ, compatibility & support policy, links to CHANGELOG/CONTRIBUTING/SECURITY.
- Doc comments on **every** exported identifier, starting with the identifier name; package-level `doc.go`
  for each non-trivial package. Verify the rendering with `go doc -all ./...`.
- Runnable `Example` functions so pkg.go.dev shows usage.
- `CHANGELOG.md` for `[vX.Y.Z]` in Keep a Changelog format:
  `Added / Changed / Deprecated / Removed / Fixed / Security`, with PR and issue links, and an
  explicit statement of whether the release is breaking under SemVer.
- Migration guide for any breaking change: before/after code, deprecation timeline, automated migration if feasible.
- Deprecation notices use the `// Deprecated: ...` convention so tooling picks them up.
- `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, issue/PR templates present and current.
- If the CLI supports it, regenerate shell completions and man pages as part of the release.
- Architecture notes / ADRs updated if the design changed.

---

# Phase 8 — Release readiness checklist

Produce this as a filled-in checklist in the PR description:

- [ ] Version number chosen per SemVer and justified
- [ ] No breaking changes, or breaking changes documented + major bump
- [ ] `go build`, `go vet`, lint, `go test -race` all green on the full matrix
- [ ] Coverage ≥ `[N]%`, no new uncovered critical paths
- [ ] `govulncheck` clean or all findings documented
- [ ] `goreleaser release --snapshot --clean` produces correct artifacts for all platforms
- [ ] CHANGELOG updated, README accurate, docs regenerated
- [ ] CI workflows pass on the release branch
- [ ] Upgrade path from `[vA.B.C]` verified (config compatibility, data/schema migration, API compatibility)
- [ ] Rollback plan documented
- [ ] Post-release verification steps listed (install from each channel and smoke test)

---

# Deliverables

1. A pull request with all changes, split into **logical, reviewable commits**
   (e.g. `chore(deps):`, `fix(lint):`, `test:`, `ci:`, `docs:`, `build:`), following Conventional Commits.
2. A PR description containing:
   - Summary table: area → what changed → risk level
   - Full list of commands executed, with their output/results
   - Behavioural changes, if any, flagged prominently
   - Findings deliberately **not** fixed, each with reasoning and a suggested follow-up issue
   - Before/after metrics: binary size, test count, coverage, benchmark deltas, dependency count
   - The exact release commands:
     ```bash
     git checkout main && git pull
     git tag -a [vX.Y.Z] -m "[vX.Y.Z]"
     git push origin [vX.Y.Z]
     # then verify the release workflow and publish the draft release
     ```
3. A draft `CHANGELOG.md` entry ready to publish as the GitHub Release notes.
4. A list of suggested follow-up issues for anything deferred.

---

# Constraints & rules of engagement

- Do **not** modify the public API surface without flagging it prominently in the PR description.
- Do **not** commit generated artifacts (`dist/`, binaries, `coverage.out`, `*.test`); ensure `.gitignore` covers them.
- Do **not** blanket-disable linters to make CI pass; fix or narrowly scope the exclusion with a comment.
- Do **not** perform major dependency upgrades as part of a patch release.
- Do **not** silently skip a step — if something cannot be completed (missing credentials, no network,
  platform unavailable), state it explicitly and explain what would be needed.
- Prefer the standard library over new dependencies; justify any dependency you add.
- Keep changes minimal and surgical; a release-prep PR is not the place for refactors — split those out.
