---
mode: agent
description: Diagnose a failing Go test or panic.
---

Diagnose this failure:

```
${input:output:paste the test output, panic, or race report}
```

1. Identify the exact failing line and the invariant that was violated.
2. Reproduce deterministically (`go test -run ... -race -count=10`).
3. Explain the root cause — not the symptom.
4. Propose the minimal fix plus a regression test that fails without it.
5. Note any other places in the codebase with the same defect pattern.