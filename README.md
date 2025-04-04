# go-caller 📞

A lightweight and idiomatic Go library that captures and formats caller information such as file name, line number, function name, and package path.

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
import "github.com/yourname/caller"

func someFunc() {
    c := caller.Immediate()
    fmt.Println("Caller location:", c.Location())
    fmt.Println("Short:", c.ShortLocation())
    fmt.Println("Function:", c.Function())
    fmt.Println("Package:", c.PackageName())
}
