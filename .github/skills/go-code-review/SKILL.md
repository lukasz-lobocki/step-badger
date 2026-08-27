---
name: go-code-review
description: Review Go changes in this repo for correctness, error handling, concurrency, API design, tests, and adherence to project conventions. Use when reviewing a PR or diff, or when asked to "review". For a full multi-concern pass delegate to the `Go Reviewer` agent; this skill is the repo-specific lens it applies.
---

# Go code review (step-badger)

Review the **diff**, not the whole tree. For a broad multi-concern pass, delegate to the
[Go Reviewer](../../agents/go-reviewer.agent.md) agent; this skill supplies the
repo-specific invariants that agent must apply.

## Verify first
Run the [go-verification](../go-verification/SKILL.md) sequence and stop at the first
failure. Never review unbuildable code. Paste actual command output, not "tests pass".

## Project invariants — do NOT flag these as violations
This repo deliberately deviates from generic Go rules (see
[AGENTS.md](../../../AGENTS.md#error-handling--logging-project-specific--differs-from-generic-rules)):
- Errors surface via the colored loggers, **not** by returning up a call stack.
  `logError.Fatalln(err)` (DB open/list/close, "no records found") and
  `logError.Panic`/`Panicf` (parse & emit helpers) are the established pattern. Do not
  suggest error-wrapping refactors here; match the surrounding file.
- Verbose logs are gated by global `loggingLevel` (max `MAX_LOGGING_LEVEL = 3`). Check new
  log lines are guarded: `>=1` info, `>=2` per-row, `>=3` per-record dumps.
- Badger is an *indirect* dep reached only through `smallstep/nosql`. Flag any direct
  badger import.

## Repo-specific checklist (on top of the generic Go Reviewer list)
**Cobra / commands** (`command_<feature>.go`)
- New subcommand registered in its own `init()`; help hidden; `Flags().SortFlags=false`.
- Arity via `cobra.ExactArgs(n)`; `Aliases` present; multi-line `Use:` with an
  `Arguments:` block.
- Flags bound into the single global `config tConfig`; choice flags via `Flags().Var(...)`.
- No new persistent flag — only root's `--logging` exists.

**Emit / column system** (`columns_*`, `output_*`)
- A new output format has an `emit<Feature><Format>` **and** a case in the handler's switch.
- Every column exposes all closures: `isShown(tConfig)`, `title()`,
  `contentSource(record, tConfig)`, `contentColor(record)`.
- No direct table or `os.Stdout` writes — go through tabby + the emit functions.

**Naming** (see [AGENTS.md](../../../AGENTS.md#naming-conventions))
- Types unexported and `t`-prefixed; `get*` accessors/columns, `parse*ValueTo<Type>`
  deserializers, `emit<Feature><Format>` output, `export<Feature>Main([]string)` handlers;
  consts SCREAMING_SNAKE.

**Database access** (`smallstep/nosql`)
- `db.Close()` called and its error checked on every path.
- Missing key treated as `errors.Is(err, database.ErrNotFound)`.

## Output
Group findings as **Blocking**, **Should fix**, **Nit** — each with `file:line`, the
problem, and a concrete snippet fixing it. No praise-padding.
