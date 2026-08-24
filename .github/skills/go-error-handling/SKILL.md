---
name: go-error-handling
description: Patterns for designing, wrapping, inspecting, and testing errors in Go. Use when writing or reviewing error paths.
---

# Go error handling

## Declaring
```go
var (
    ErrNotFound = errors.New("not found")
    ErrConflict = errors.New("conflict")
)

type ValidationError struct {
    Field  string
    Reason string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}
```

## Wrapping
```go
if err := db.Get(ctx, id); err != nil {
    return fmt.Errorf("get user %s: %w", id, err)
}
```
- Add the *operation and its inputs*, not the type name.
- Wrap once per layer. Do not wrap in the same function that created the error.
- Use `%w` only when callers should be able to unwrap; otherwise `%v`.

## Inspecting
```go
if errors.Is(err, ErrNotFound) { ... }

var verr *ValidationError
if errors.As(err, &verr) { ... }
```

## Joining
```go
err = errors.Join(err1, err2)   // for independent, parallel failures
```

## Boundaries
- **Library**: return errors, never log them.
- **Handler/main**: log once, at the outermost layer, with the full chain.
- **HTTP/gRPC**: map sentinels to status codes in one dedicated function.
- Never expose internal error text to end users.

## Anti-patterns
- `if err != nil { return err }` at every layer with no added context.
- `strings.Contains(err.Error(), "...")`.
- Returning `nil` error alongside a zero value that the caller cannot distinguish.
- `panic` for control flow; recovering broadly and continuing.

## Testing
```go
if !errors.Is(err, tc.wantErr) {
    t.Fatalf("err = %v, want %v", err, tc.wantErr)
}
```
Test the *behaviour* (`errors.Is`), never the message string.