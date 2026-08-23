---
mode: agent
description: Refactor a Go package for cohesion and clarity, preserving behaviour.
agent: Go Refactorer
---

Refactor the package `${input:package:e.g. internal/store}`.

Do this in order and stop for approval after step 3:

1. Summarise the package: exported surface, dependencies, and every caller.
2. List the top problems (cohesion, naming, complexity, error handling,
   concurrency, testability), each with file:line evidence.
3. Propose an ordered, minimal, behaviour-preserving refactoring plan.
4. On approval, execute step by step, running
   `go build ./... && go vet ./... && go test ./... -race -count=1`
   after each step.

Do not change exported signatures without flagging them as breaking.