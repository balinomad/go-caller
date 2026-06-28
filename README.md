[![GoDoc](https://pkg.go.dev/badge/github.com/balinomad/go-caller/v2?status.svg)](https://pkg.go.dev/github.com/balinomad/go-caller/v2?tab=doc)
[![GoMod](https://img.shields.io/github/go-mod/go-version/balinomad/go-caller)](https://github.com/balinomad/go-caller)
[![Size](https://img.shields.io/github/languages/code-size/balinomad/go-caller)](https://github.com/balinomad/go-caller)
[![License](https://img.shields.io/github/license/balinomad/go-caller)](./LICENSE)
[![Go](https://github.com/balinomad/go-caller/actions/workflows/go.yml/badge.svg)](https://github.com/balinomad/go-caller/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/balinomad/go-caller/v2)](https://goreportcard.com/report/github.com/balinomad/go-caller/v2)
[![codecov](https://codecov.io/github/balinomad/go-caller/graph/badge.svg?token=L1K68IIN51)](https://codecov.io/github/balinomad/go-caller)

# caller

_A lightweight and idiomatic Go library that captures and formats caller information such as file name, line number, function name, and package path._

Perfect for use in:

- Custom error types
- Logging wrappers
- Debugging utilities
- Tracing and instrumentation

## Why not just use `runtime.Caller`?

`runtime.Caller` and `runtime.FuncForPC` already give you the raw data this package is built on:

```go
pc, file, line, ok := runtime.Caller(0)
if !ok {
    // handle failure
}
fn := ""
if f := runtime.FuncForPC(pc); f != nil {
    fn = f.Name()
}
```

That gets you a file path, a line number, and one raw string such as `main.(*Service).Handle`. Splitting that string into a package path, a package name, and a bare method name — correctly, across closures, methods, vendored import paths, and functions with no package at all — is the part most people rewrite slightly differently in every project and rarely cover with tests:

```go
c := caller.Immediate()
c.Function()    // "(*Service).Handle"
c.Package()     // "main"
c.PackageName() // "main"
```

`go-caller` does that parsing once, with full test coverage, and adds ready-made `json.Marshaler`/`json.Unmarshaler` and `slog.LogValuer` implementations so caller info drops straight into structured logs or JSON-encoded errors without writing that glue yourself.

If you only need a raw `file:line` for one `fmt.Println`, the stdlib alone is enough. Reach for this package once you want that information parsed, structured, serializable, or handled consistently across a codebase.

## Features

- Minimal and dependency-free
- Clean, consistent API with platform-independent file paths
- Get full or short location (`file:line`)
- Extract function name, full function path, and package info
- Implements `fmt.Stringer` interface for easy logging
- Implements `json.Marshaler` and `json.Unmarshaler` interfaces for easy JSON serialization
- Implements `slog.LogValuer` interface for structured logging
- Semantic equality comparison between callers

## Requirements

Go 1.22 or later.

## Installation

### For v2.x (latest)

```bash
go get github.com/balinomad/go-caller/v2@latest
```

## Usage

```go
import "github.com/balinomad/go-caller/v2"

func someFunc() {
    c := caller.Immediate()

    // Check validity
    if c.Valid() {
        fmt.Println("Caller information is valid")
    }

    // Basic information
    fmt.Println("Caller location:", c.Location())
    fmt.Println("Short:", c.ShortLocation())
    fmt.Println("Function:", c.Function())
    fmt.Println("Package:", c.PackageName())
}
```

## API Reference

### Constructor Functions

| Function                       | Description                                             |
| ------------------------------ | ------------------------------------------------------- |
| `Immediate() Caller`           | Returns caller info for the immediate caller            |
| `New(skip int) Caller`         | Returns caller info with custom stack skip depth        |
| `NewFromPC(pc uintptr) Caller` | Creates caller info from a program counter              |
| `NewEmpty() Caller`            | Returns an empty, invalid `Caller` for `json.Unmarshal` |

### Caller Interface Methods

| Method                          | Description                                           | Example Output                   |
| ------------------------------- | ----------------------------------------------------- | -------------------------------- |
| `Valid() bool`                  | Returns true if the caller info is usable             | `true`/`false`                   |
| `File() string`                 | Full file path                                        | `/path/to/file.go`               |
| `Line() int`                    | Line number                                           | `42`                             |
| `Location() string`             | Full location with file:line                          | `/path/to/file.go:42`            |
| `ShortLocation() string`        | Short location with just filename:line                | `file.go:42`                     |
| `Function() string`             | Function/method name without package                  | `MyFunction`                     |
| `FullFunction() string`         | Full function name including package                  | `github.com/user/pkg.MyFunction` |
| `Package() string`              | Full import path of the package                       | `github.com/user/pkg`            |
| `PackageName() string`          | Last element of the package path                      | `pkg`                            |
| `Equal(other Caller) bool`      | Checks if two callers are semantically equal          | `true`/`false`                   |
| `String() string`               | Returns `ShortLocation()` (implements `fmt.Stringer`) | `file.go:42`                     |
| `MarshalJSON() ([]byte, error)` | Marshals caller info to JSON                          | `{"file":"...","line":42,...}`   |
| `UnmarshalJSON([]byte) error`   | Unmarshals JSON to caller info                        | -                                |
| `LogValue() slog.Value`         | Returns structured value for slog                     | `{file:..., line:42, ...}`       |

`Equal` treats a nil `Caller` as never equal to anything, including another nil `Caller` — there is no "two unset callers are the same" case.

## Advanced Usage

### Custom Stack Depth

`New` is meant to be called from inside your own helper function, the same way `log.Logger.Output`'s `calldepth` parameter works: `skip` accounts for any wrapper frames above the line where you actually want the location captured.

```go
func logWithCaller(msg string) {
    c := caller.New(0) // resolves to whoever called logWithCaller, not to this line
    fmt.Println(c.Location(), msg)
}
```

Calling `New(0)` directly from top-level code, with no wrapper function of your own in between, resolves one frame higher than that — use `Immediate()` instead when you want your own call site captured with no wrapper involved.

### Using with Program Counter

```go
pc, _, _, _ := runtime.Caller(0)
c := caller.NewFromPC(pc)
```

`pc` must be a call-site program counter such as `runtime.Caller`'s first return value. Program counters captured via `runtime.Callers` are return addresses, not call-site addresses — passing one of those directly to `NewFromPC` can resolve to the wrong line, or even an unrelated function. Subtract 1 from a `runtime.Callers` value before passing it here, or resolve frames with `runtime.CallersFrames` instead.

### JSON Serialization

```go
// Marshal
c := caller.Immediate()
jsonData, err := json.Marshal(c)
// Output: {"file":"main.go","line":10,"function":"main","package":"main"}

// Unmarshal
c2 := caller.NewEmpty()
err = json.Unmarshal(jsonData, &c2)
```

`Caller` is an interface with no exported implementation, so `json.Unmarshal` has no concrete type to construct on its own — `NewEmpty()` is what gives you one to unmarshal into.

### Structured Logging with slog

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
c := caller.Immediate()
logger.Info("user action",
    slog.String("action", "login"),
    slog.String("caller", c.Location()),
    slog.String("func", c.Function()))
// Output includes structured caller information
```

### Comparing Callers

```go
c1 := caller.Immediate()
// ... some other code ...
c2 := caller.Immediate()

if c1.Equal(c2) {
    fmt.Println("Called from the same location")
} else {
    fmt.Println("Called from different locations")
}
```

## Concurrency

A `Caller` is safe for concurrent reads once constructed — multiple goroutines may call `Location()`, `Function()`, `MarshalJSON()`, and the other accessors on the same instance at the same time. The one exception is `UnmarshalJSON`: it mutates the receiver in place with no internal locking, so it must not be called on a `Caller` that another goroutine might be reading or unmarshaling into concurrently. Populate a `Caller` fully (via `NewEmpty()` + `json.Unmarshal`, or one of the constructors) before sharing it across goroutines.

## Migration from v1 to v2

### Breaking Changes

1. **`Valid()` behavior**: Now only requires a non-empty file path (previously required file, line, and function)
2. **`New()` edge cases**: Returns `nil` for a negative skip value, or when the requested stack depth exceeds the call stack

### New Features

- Added `Equal()` method for comparing callers
- Added `NewEmpty()` constructor to support unmarshaling JSON into a fresh `Caller`
- Improved performance and error handling
- Comprehensive test coverage

### Migration Example

```go
// v1 code that might need review
c := caller.New(0)
if c.Valid() {
    // In v1: Valid only if file, line, AND function are present
    // In v2: Valid if file is present (more permissive)
}

// New v2 feature
if c1.Equal(c2) {
    // Semantic comparison of callers
}
```

## Performance

The library is designed to be lightweight with minimal allocations:

- Zero dependencies beyond Go standard library
- A `Caller` is a single small struct, one heap allocation per capture
- Comprehensive benchmarks included in tests

## Testing

Run tests with:

```bash
go test -race -v ./...
```

Run benchmarks with:

```bash
go test -bench=. -benchmem
```

## License

MIT License — see `LICENSE` file for details.
