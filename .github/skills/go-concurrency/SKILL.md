---
name: go-concurrency
description: Safe goroutine, channel, context, and synchronisation patterns for Go. Use when writing or auditing concurrent code.
---

# Go concurrency

## Rules
1. Whoever starts a goroutine is responsible for stopping it.
2. Never start a goroutine without knowing how it exits.
3. Only the sender closes a channel.
4. `context.Context` is the first parameter, never a struct field.
5. Run everything under `-race`.

## Bounded fan-out
```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(runtime.GOMAXPROCS(0))
for _, item := range items {
    g.Go(func() error { return process(ctx, item) })
}
if err := g.Wait(); err != nil { return err }
```

## Worker pool with clean shutdown
```go
jobs := make(chan Job)
var wg sync.WaitGroup
for range n {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for j := range jobs { handle(ctx, j) }
    }()
}
// producer
for _, j := range all { 
    select {
    case jobs <- j:
    case <-ctx.Done():
    }
}
close(jobs)
wg.Wait()
```

## Cancellation
```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()          // always, even on the success path
```

## Lazy init
```go
var load = sync.OnceValue(func() *Config { return mustLoad() })
```

## Audit checklist
- [ ] Every `go func` has a bounded lifetime.
- [ ] Every `context.With*` result has a `defer cancel()`.
- [ ] No lock held across a channel op, I/O call, or another lock.
- [ ] Consistent lock ordering (document it if more than one lock exists).
- [ ] `select` on sends into potentially-blocked channels includes `<-ctx.Done()`.
- [ ] No `time.Sleep` used as synchronisation.
- [ ] Shared maps/slices are copied or mutex-guarded, not shared by reference.

## Testing
- `go test -race -count=10` on concurrency tests.
- Use `testing/synctest` (Go 1.24+) for deterministic time-based tests.
- Detect leaks with `go.uber.org/goleak` in `TestMain`.