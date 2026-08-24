---
mode: agent
description: Raise test coverage for a package with meaningful tests.
agent: Go Test Author
---

Add tests for `${input:package}`.

1. Report current coverage per function.
2. List the untested branches worth testing, ranked by risk.
3. Write table-driven subtests covering them, plus fuzz targets for any parser.
4. Report the new coverage and confirm `go test ./... -race -count=1` passes.

Do not chase coverage percentage with trivial assertions.