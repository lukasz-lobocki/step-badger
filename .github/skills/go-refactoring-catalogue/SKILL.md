---
name: go-refactoring-catalogue
description: Catalogue of safe, mechanical Go refactorings with preconditions and verification steps. Use when restructuring Go code.
---

# Go refactoring catalogue

Each entry: **precondition → steps → verification**.

## Extract package
- *Pre*: a cohesive symbol cluster with no back-references to the origin package.
- *Steps*: move files verbatim → fix package clause → add explicit imports →
  narrow the exported surface to what callers actually use.
- *Verify*: `go build ./...`; import graph has no new cycle (`go mod graph`).

## Invert dependency (move interface to consumer)
- *Pre*: producer package defines an interface only its consumer uses.
- *Steps*: declare a minimal interface in the consumer → delete the producer's
  interface → producer returns a concrete type.
- *Verify*: compile; the producer no longer imports the consumer.

## Replace bool parameter
- *Pre*: `f(x, true)` appears at call sites and the bool selects behaviour.
- *Steps*: introduce a named type or split into `fEager`/`fLazy`.
- *Verify*: every call site reads unambiguously.

## Introduce functional options
- *Pre*: constructor has >3 parameters or a growing config struct.
- *Steps*: `type Option func(*config)` → `New(required, opts ...Option)` →
  keep old constructor as a deprecated wrapper for one release.
- *Verify*: existing call sites compile unchanged.

## Replace interface{} with generics
- *Pre*: type switches over a closed, homogeneous set.
- *Steps*: define the constraint → parameterise → remove runtime assertions.
- *Verify*: no `.(type)` remains; benchmarks show no regression.

## Thread context
- *Pre*: blocking call (I/O, lock, sleep) with no cancellation.
- *Steps*: add `ctx context.Context` as first param up the call chain →
  `context.TODO()` at the boundary → replace with real ctx → honour `ctx.Done()`.
- *Verify*: `go vet` `lostcancel` clean; cancellation test passes.

## Split god function
- *Pre*: >50 lines, or cyclomatic complexity >10, or >3 levels of nesting.
- *Steps*: identify the seams (comment-delimited blocks) → extract each into a
  named unexported function → invert conditions for early return.
- *Verify*: behaviour identical; each extracted function independently testable.

## Collapse error handling
- *Pre*: repeated `if err != nil { return fmt.Errorf(...) }` with identical context.
- *Steps*: hoist into a helper, or use a struct with a sticky `err` field for
  sequential I/O (the `bufio.Scanner` pattern).
- *Verify*: no error is silently swallowed; `errcheck` clean.

## Remove single-implementation interface
- *Pre*: interface with exactly one implementation and no test double need.
- *Steps*: replace all uses with the concrete type → delete the interface.
- *Verify*: compile; note that this *reduces* indirection deliberately.

## Universal rule
One refactoring per commit. Tests green before and after. If a refactoring
requires changing a test's assertions, it is not behaviour-preserving — stop.