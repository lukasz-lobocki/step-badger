---
applyTo: "**/*_test.go"
description: Conventions for Go tests.
---

# Go test conventions

- Table-driven pattern:

```go
tests := map[string]struct {
    in      string
    want    int
    wantErr error
}{
    "empty":   {in: "", want: 0},
    "invalid": {in: "x", wantErr: ErrInvalid},
}
for name, tc := range tests {
    t.Run(name, func(t *testing.T) {
        t.Parallel()
        got, err := Parse(tc.in)
        if !errors.Is(err, tc.wantErr) {
            t.Fatalf("err = %v, want %v", err, tc.wantErr)
        }
        if got != tc.want {
            t.Errorf("got %d, want %d", got, tc.want)
        }
    })
}
```

- Use `t.Parallel()` unless the test mutates global state.
- Prefer `testing` + `google/go-cmp` over assertion libraries.
- Compare structs with `cmp.Diff(want, got)` and report `(-want +got)`.
- Use `httptest` for HTTP, `testing/fstest` for filesystems.
- Add fuzz targets (`FuzzXxx`) for parsers and decoders.
- Benchmarks: `b.ReportAllocs()`, `b.Loop()` (Go 1.24+) or `b.ResetTimer()`.
- No `time.Sleep` for synchronisation — use channels, `sync.WaitGroup`, or `synctest`.