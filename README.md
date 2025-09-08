[![Go](https://github.com/balinomad/go-caller/actions/workflows/go.yml/badge.svg)](https://github.com/balinomad/go-caller/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/balinomad/go-caller)](https://goreportcard.com/report/github.com/balinomad/go-caller)

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

## 🚀 Usage

```go
import "github.com/balinomad/go-caller"

func someFunc() {
    c := caller.Immediate()
    fmt.Println("Caller location:", c.Location())
    fmt.Println("Short:", c.ShortLocation())
    fmt.Println("Function:", c.Function())
    fmt.Println("Package:", c.PackageName())
    data, err := json.Marshal(c)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("JSON:", string(data))
}
```

## 📌 Installation

```bash
go get github.com/balinomad/go-caller@latest
```

## 📘 API Highlights

| Method            | Description                                     |
|-------------------|-------------------------------------------------|
| `File()`          | Full file path                                  |
| `Line()`          | Line number                                     |
| `Location()`      | Full location (`path/to/file.go:123`)           |
| `ShortLocation()` | Short location (`file.go:123`)                  |
| `Function()`      | Method/function name only                       |
| `FullFunction()`  | Full path to method including package           |
| `Package()`       | Full import path of the package                 |
| `PackageName()`   | Last element of the package path                |
| `String()`        | Returns `ShortLocation()` for easy logging      |
| `MarshalJSON()`   | Marshal caller info to JSON                     |
| `UnmarshalJSON()` | Unmarshal JSON to caller info                   |
| `LogValue()`      | Construct a `slog.Value` for structured logging |

## 🔧 Advanced

For custom skipping depth or use with program counters:

```go
c := caller.New(skip)
c := caller.NewFromPC(pc)
```

## ⚖️ License

MIT License — see `LICENSE` file for details.
