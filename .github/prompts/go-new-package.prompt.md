---
mode: agent
description: Scaffold a new internal Go package with tests and docs.
---

Create the package `internal/${input:name}` with:

- `doc.go` containing a package comment explaining purpose and usage.
- The primary type with a `New...` constructor returning `(*T, error)`.
- Options via functional options if there are more than two settings.
- Sentinel errors declared at the top of the file.
- `${input:name}_test.go` with table-driven tests and an `Example`.

No global state, no `init()`, no external dependencies unless I approve them.
Finish with the full verification suite.