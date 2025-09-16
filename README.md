[![GoDoc](https://pkg.go.dev/badge/github.com/balinomad/go-caller/v2?status.svg)](https://pkg.go.dev/github.com/balinomad/go-caller/v2?tab=doc)
[![GoMod](https://img.shields.io/github/go-mod/go-version/balinomad/go-caller)](https://github.com/balinomad/go-caller)
[![Size](https://img.shields.io/github/languages/code-size/balinomad/go-caller)](https://github.com/balinomad/go-caller)
[![License](https://img.shields.io/github/license/balinomad/go-caller)](./LICENSE)
[![Go](https://github.com/balinomad/go-caller/actions/workflows/go.yml/badge.svg)](https://github.com/balinomad/go-caller/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/balinomad/go-caller/v2)](https://goreportcard.com/report/github.com/balinomad/go-caller/v2)
[![codecov](https://codecov.io/github/balinomad/go-caller/graph/badge.svg?token=L1K68IIN51)](https://codecov.io/github/balinomad/go-caller)

# caller

*A lightweight and idiomatic Go library that captures and formats caller information such as file name, line number, function name, and package path.*

Perfect for use in:
- Custom error types
- Logging wrappers
- Debugging utilities
- Tracing and instrumentation

## ✨ Features

- Minimal and dependency-free
- Clean, consistent API with platform-independent file paths
- Get full or short location (`file:line`)
- Extract function name, full function path, and package info
- Implements `fmt.Stringer` interface for easy logging
- Implements `json.Marshaler` and `json.Unmarshaler` interfaces for easy JSON serialization
- Implements `slog.LogValuer` interface for structured logging
- Semantic equality comparison between callers

## 📌 Installation

### For v2.x (latest)

```bash
go get github.com/balinomad/go-caller/v2@latest
```

## 🚀 Usage

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

## 📘 API Reference

### Constructor Functions

| Function | Description |
|----------|-------------|
| `Immediate() Caller` | Returns caller info for the immediate caller |
| `New(skip int) Caller` | Returns caller info with custom stack skip depth |
| `NewFromPC(pc uintptr) Caller` | Creates caller info from a program counter |

### Caller Interface Methods

| Method | Description | Example Output |
|--------|-------------|----------------|
| `Valid() bool` | Returns true if the caller info is usable | `true`/`false` |
| `File() string` | Full file path | `/path/to/file.go` |
| `Line() int` | Line number | `42` |
| `Location() string` | Full location with file:line | `/path/to/file.go:42` |
| `ShortLocation() string` | Short location with just filename:line | `file.go:42` |
| `Function() string` | Function/method name without package | `MyFunction` |
| `FullFunction() string` | Full function name including package | `github.com/user/pkg.MyFunction` |
| `Package() string` | Full import path of the package | `github.com/user/pkg` |
| `PackageName() string` | Last element of the package path | `pkg` |
| `Equal(other Caller) bool` | Checks if two callers are semantically equal | `true`/`false` |
| `String() string` | Returns `ShortLocation()` (implements `fmt.Stringer`) | `file.go:42` |
| `MarshalJSON() ([]byte, error)` | Marshals caller info to JSON | `{"file":"...","line":42,...}` |
| `UnmarshalJSON([]byte) error` | Unmarshals JSON to caller info | - |
| `LogValue() slog.Value` | Returns structured value for slog | `{file:..., line:42, ...}` |

## 🔧 Advanced Usage

### Custom Stack Depth

```go
// Skip 0 = immediate caller (same as Immediate())
// Skip 1 = caller of the immediate caller
// Skip 2 = caller of the caller, etc.
c := caller.New(2)
```

### Using with Program Counter

```go
pc, _, _, _ := runtime.Caller(0)
c := caller.NewFromPC(pc)
```

### JSON Serialization

```go
// Marshal
c := caller.Immediate()
jsonData, err := json.Marshal(c)
// Output: {"file":"main.go","line":10,"function":"main","package":"main"}

// Unmarshal
var c2 caller.Caller
err = json.Unmarshal(jsonData, &c2)
```

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

## 🔄 Migration from v1 to v2

### Breaking Changes

1. **`Valid()` behavior**: Now only requires a non-empty file path (previously required file, line, and function)
2. **`New()` edge cases**: Only returns `nil` for negative skip values

### New Features

- Added `Equal()` method for comparing callers
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

## 📊 Performance

The library is designed to be lightweight with minimal allocations:
- Zero dependencies beyond Go standard library
- Optimized string operations
- Efficient memory usage with `uint16` for line numbers
- Comprehensive benchmarks included in tests

## 🧪 Testing

Run tests with:
```bash
go test -v
```

Run benchmarks with:
```bash
go test -bench=. -benchmem
```

## ⚖️ License

MIT License — see `LICENSE` file for details.
