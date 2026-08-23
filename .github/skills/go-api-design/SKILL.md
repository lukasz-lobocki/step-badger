---
name: go-api-design
description: Designing and evolving Go package APIs — naming, options, compatibility, deprecation. Use when adding or changing exported identifiers.
---

# Go API design

## Naming
- Package name is part of the identifier: `http.Server`, not `http.HTTPServer`.
- Avoid stutter: `store.New`, not `store.NewStore`.
- Getters have no `Get` prefix: `u.Name()`. Setters use `SetName`.
- Single-method interfaces end in `-er`: `Reader`, `Validator`.

## Shape
- Return concrete types; accept interfaces.
- Constructor: `New(required...) (*T, error)`; add options for the optional parts.
- Make the zero value useful when you can (`bytes.Buffer`, `sync.Mutex`).
- Prefer a slice of results over a callback, unless streaming is required.
- Put `ctx` first, `opts ...Option` last, `error` last in returns.

## Functional options
```go
type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func New(addr string, opts ...Option) (*Server, error) {
    s := &Server{addr: addr, timeout: 30 * time.Second}
    for _, o := range opts { o(s) }
    return s, s.validate()
}
```

## Compatibility
Breaking, for a released package:
- removing/renaming an exported identifier
- changing a signature, adding a method to an exported interface
- changing struct field types, or adding fields to a struct used as a positional literal

Non-breaking: adding functions, adding methods to a concrete type, adding
options, adding fields to a struct whose literals are always keyed
(add an unexported zero-width field to enforce this).

## Deprecation
```go
// Deprecated: use NewWithContext instead. Removed in v3.
func New(addr string) *Server { ... }
```
Keep for at least one minor release. Update all internal callers immediately.

## Documentation
Every exported identifier gets a comment starting with its name and forming a
complete sentence. Document error conditions, concurrency safety, and ownership
of passed-in slices/maps/readers.