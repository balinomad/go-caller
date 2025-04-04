# go-caller

*A lightweight and idiomatic Go library that captures and formats caller information such as file name, line number, function name, and package path.*

Perfect for use in:
- Custom error types
- Logging wrappers
- Debugging utilities
- Tracing and instrumentation

## ✨ Features

- Get full or short location (`file:line`)
- Extract function name, full function path, and package info
- Clean, consistent API with platform-independent internals
- Implements `fmt.Stringer` for easy logging
- Minimal and dependency-free

## 🚀 Usage

```go
import "github.com/balinomad/go-caller"

func someFunc() {
    c := caller.Immediate()
    fmt.Println("Caller location:", c.Location())
    fmt.Println("Short:", c.ShortLocation())
    fmt.Println("Function:", c.Function())
    fmt.Println("Package:", c.PackageName())
}
```

## 📌 Installation

```bash
go get github.com/balinomad/go-caller@latest
```

## 📘 API Highlights

| Method           | Description                                  |
|------------------|----------------------------------------------|
| `File()`         | Full file path                               |
| `Line()`         | Line number                                  |
| `Location()`     | Full location (`path/to/file.go:123`)        |
| `ShortLocation()`| Short location (`file.go:123`)               |
| `Function()`     | Method/function name only                    |
| `FullFunction()` | Full path to method including package        |
| `Package()`      | Full import path of the package              |
| `PackageName()`  | Last element of the package path             |
| `String()`       | Returns `ShortLocation()` for easy logging   |

## 🔧 Advanced

For custom skipping depth or use with program counters:

```go
c := caller.New(skip)
c := caller.NewFromPC(pc)
```

## ⚖️ License

MIT License — see `LICENSE` file for details.
