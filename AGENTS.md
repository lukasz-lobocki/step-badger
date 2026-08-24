# AGENTS.md — step-badger

Go CLI that exports certificate and table data from a **step-ca** Badger database.
Single package `cmd`, imported by a thin `main.go` (`cmd.Execute()`).

> Generic Go style, error, concurrency, and testing rules live in
> [.github/copilot-instructions.md](.github/copilot-instructions.md) — follow them.
> This file covers only what is specific to this repo. See [README.md](README.md) for
> usage and the Badger single-user copy workaround, and [BUILD.md](BUILD.md) for releases.

## Build / test / lint
- Build: `make snapshot` (goreleaser) or `go build ./...`. Run: `make run`.
- Lint/format is enforced by [.golangci.yml](.golangci.yml): `gofumpt`, `goimports`, and
  linters incl. `errorlint`, `gosec`, `revive` (exported/package-comments/error-strings),
  `staticcheck`, `unused`. Run `golangci-lint run` before committing.
- **There are no Go test files yet.** If you add tests, follow the conventions in
  copilot-instructions.md (table-driven, `-race`).

## Architecture — file roles (all in `package cmd`)
Files use a **role prefix** and, per feature, a **suffix** (`_sshcerts`, `_x509certs`):
- `command_<feature>.go` — the `cobra.Command`, its `init()` registration + flags, the
  `export<Feature>Main([]string)` handler, and DB lookup/parse helpers.
- `defs_<feature>.go` / `defs_root.go` — record types (`json:` tagged), shared consts,
  and cross-feature helpers (loggers, choices, markdown escaping).
- `columns_<feature>.go` — declarative column model + `get<Feature>Columns()`.
- `output_<feature>.go` — one `emit<Feature><Format>` function per output format.
- `aux_choice_flag.go` — the reusable `tChoice` pflag value type.

Features: `sshcerts`, `x509certs`, `dbtable`, and hidden `markdowndocs`.

## Cobra conventions
- Each subcommand file's `init()` does `rootCmd.AddCommand(...)` — **there is no central
  registry**; registration is per-file. Always also hide the help cmd
  (`SetHelpCommand(&cobra.Command{Hidden: true})`) and set `Flags().SortFlags = false`.
- Use `DisableFlagsInUseLine: true` and a multi-line `Use:` with an `Arguments:` block.
- Validate arity with `cobra.ExactArgs(n)`. Provide `Aliases`.
- **Only one persistent flag** exists: root's `--logging`. All other flags are local,
  declared in the subcommand's `init()`, and bound directly into fields of the single
  global `config tConfig` (e.g. `--valid` → `config.showValid`). Choice flags bind to
  `*tChoice` via `Flags().Var(...)`.

## Database access (smallstep/nosql)
- Open: `db, err := nosql.New("badgerv2", path, database.WithValueDir(path))` — returns
  `database.DB`. Badger is an *indirect* dep; never import it directly.
- List a bucket: `db.List([]byte(bucket))` → records with `.Bucket/.Key/.Value`.
  Buckets: `ssh_certs`, `x509_certs`; revocation lookups on `revoked_ssh_certs` /
  `revoked_x509_certs`; x509 provisioner data on `x509_certs_data`.
- Read a key: `db.Get([]byte(bucket), []byte(key))`; treat missing as
  `errors.Is(err, database.ErrNotFound)` (`github.com/pkg/errors`).
- Always `db.Close()` and check the error.

## Output / emit system
- Format is chosen by a `switch` on the bound choice value: `config.emitSshFormat.Value`
  or `config.emitX509Format.Value`, against `FORMAT_*` consts in `defs_root.go`.
  ssh supports table/json/markdown/plain; x509 adds `openssl`.
- Tables render with **`github.com/lukasz-lobocki/tabby`** (`new(tabby.Table)`,
  `SetHeader`, `AppendRow`, `Print(nil)`). Colors via `fatih/color`.
- Columns are data-driven: each column struct exposes closures `isShown(tConfig) bool`,
  `title()`, `contentSource(record, tConfig) string`, `contentColor(record)`, plus
  markdown align/escape fields. Emitters iterate the slice from `get<Feature>Columns()`.

## Error handling & logging (project-specific — differs from generic rules)
- Errors surface via the three colored loggers (`logInfo`/`logWarning`/`logError`,
  created in `initLoggers()`), **not** by returning up a call stack. Existing code mixes:
  - `logError.Fatalln(err)` — DB open/list/close, "no records found".
  - `logError.Panic(err)` / `logError.Panicf(...)` — parse & emit helpers.
  Match the surrounding file's pattern; do not introduce new error-wrapping style here.
- Verbosity is gated by global `loggingLevel` (max `MAX_LOGGING_LEVEL = 3`):
  `>=1` config/open/close info, `>=2` per-row progress, `>=3` per-record dumps + table
  spacing debug. Guard verbose logs with these level checks.

## Naming conventions
- Types: unexported, `t`-prefixed (`tConfig`, `tChoice`, `tSshColumn`,
  `tCertificateRevocation`, ...). Functions: `get*` accessors/columns,
  `parse*ValueTo<Type>` deserializers, `emit<Feature><Format>` output,
  `export<Feature>Main([]string)` cobra handler. Consts are SCREAMING_SNAKE.

## Adding a new subcommand/feature
1. `cmd/command_<feature>.go` — `<feature>Cmd *cobra.Command` + `init()` (register to
   `rootCmd`, hide help, `SortFlags=false`, flags bound into `config`).
2. `cmd/defs_<feature>.go` — record types with `json:` tags.
3. `cmd/columns_<feature>.go` — `<Feature>Column` struct + `get<Feature>Columns()`.
4. `cmd/output_<feature>.go` — an `emit<Feature><Format>` per format, dispatched by a
   `switch` on the emit choice inside the main handler.
5. `cmd/defs_root.go` — add new `tConfig` fields and, if needed, wire any new shared
   choice in `initChoices()` (+ its `FORMAT_*` const).
6. Nothing else to register — `init()` auto-registers; leave `main.go` untouched.
