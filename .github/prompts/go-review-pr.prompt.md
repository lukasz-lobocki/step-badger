---
mode: agent
description: Review the current Go diff.
agent: Go Reviewer
---

Review the changes in the current branch against `${input:base:main}`.

Focus on: correctness, error wrapping, goroutine/context lifetime, allocation in
hot paths, exported-API compatibility, and whether the tests would catch a
regression.

Output **Blocking / Should fix / Nit**, each item with file:line and a patch snippet.
If you find nothing blocking, say so plainly.