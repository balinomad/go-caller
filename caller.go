/*
Package caller provides utilities to extract source code location
information (file, line, function, and package) for the current
or specified call frame.
It is designed for use in logging, error reporting, and debugging
with a lightweight and idiomatic API. Caller captures runtime metadata
using the Go runtime and formats it in a developer-friendly way.

Example usage:

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
*/
package caller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Caller provides access to source information about the caller.
type Caller interface {
	fmt.Stringer
	json.Marshaler
	json.Unmarshaler
	slog.LogValuer

	// Valid returns true if the caller is usable.
	Valid() bool

	// File returns the file name.
	File() string

	// Line returns the line number.
	Line() int

	// Location returns a formatted string with file:line.
	Location() string

	// ShortLocation returns a formatted string with just filename:line.
	ShortLocation() string

	// Function returns just the function or method name
	// without package prefix.
	Function() string

	// FullFunction returns the full function name including package.
	FullFunction() string

	// Package returns the full import path of the function.
	Package() string

	// PackageName returns the name of the package without the directory.
	PackageName() string

	// Equal reports whether this caller is semantically equal to another.
	Equal(other Caller) bool
}

// callerInfo represents source information about the caller.
// It implements the Caller interface.
//
// Using uint16 instead of int to save space
// and to force limitations.
type callerInfo struct {
	file   string // File name
	line   uint16 // Line number
	fn     string // Function name
	dotIdx int    // Index of the function name dot separator within the full name
}

// caller implements the Caller interface.
var _ Caller = (*callerInfo)(nil)

// skipAdjust is the number of stack frames to skip
// to get to the caller of the function that creates Caller.
const skipAdjust = 2

// New returns a new Caller with source information populated.
// The skip parameter specifies the number of stack frames to skip
// in addition to the default offset. Use 0 to get the immediate caller.
// It returns nil if the skip is invalid or the caller cannot be determined.
func New(skip int) Caller {
	// A negative skip is invalid as it would look up the stack
	if skip < 0 {
		return nil
	}

	// Get caller information with the effective depth to skip
	pc, file, line, ok := runtime.Caller(skip + skipAdjust)
	if !ok {
		return nil
	}

	// Get the full function name
	var fullFunc string
	if f := runtime.FuncForPC(pc); f != nil {
		fullFunc = f.Name()
	}

	// Validate the line
	lineUint, ok := safeUint16(line)
	if !ok {
		lineUint = 0
	}

	return &callerInfo{
		file:   file,
		line:   lineUint,
		fn:     fullFunc,
		dotIdx: functionNameIndex(fullFunc),
	}
}

// Immediate returns a Caller for the immediate caller of the function
// that calls Immediate().
// It returns nil if the caller cannot be determined.
func Immediate() Caller {
	return New(0)
}

// NewFromPC returns a new Caller with source information populated
// based on the provided program counter.
// It returns nil if the caller cannot be determined.
func NewFromPC(pc uintptr) Caller {
	var (
		fullFunc string
		file     string
		line     int
	)

	// Get the full function name
	f := runtime.FuncForPC(pc)
	if f == nil {
		return nil
	}

	// Get the full function name, file, and line
	fullFunc = f.Name()
	file, line = f.FileLine(pc)

	// Validate the line
	lineUint, ok := safeUint16(line)
	if !ok {
		lineUint = 0
	}

	return &callerInfo{
		file:   file,
		line:   lineUint,
		fn:     fullFunc,
		dotIdx: functionNameIndex(fullFunc),
	}
}

// Valid returns true if the caller is usable.
func (c *callerInfo) Valid() bool {
	return c != nil && c.file != ""
}

// File returns the file name.
func (c *callerInfo) File() string {
	if c == nil {
		return ""
	}
	return c.file
}

// Line returns the line number.
func (c *callerInfo) Line() int {
	if c == nil {
		return 0
	}
	return int(c.line)
}

// Location returns a formatted string with file:line.
func (c *callerInfo) Location() string {
	if c == nil || c.file == "" {
		return ""
	}
	if c.line <= 0 {
		return c.file
	}

	var sb strings.Builder
	sb.WriteString(c.file)
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(c.Line()))
	return sb.String()
}

// ShortLocation returns a formatted string with just filename:line.
func (c *callerInfo) ShortLocation() string {
	if c == nil || c.file == "" {
		return ""
	}
	shortFile := filepath.Base(c.file)
	if c.line <= 0 {
		return shortFile
	}

	var sb strings.Builder
	sb.WriteString(shortFile)
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(c.Line()))
	return sb.String()
}

// Function returns just the function or method name
// without package prefix.
func (c *callerInfo) Function() string {
	if c == nil || c.fn == "" || c.dotIdx < 0 || c.dotIdx >= len(c.fn)-1 {
		return ""
	}
	return c.fn[c.dotIdx+1:]
}

// FullFunction returns the full function name including package.
func (c *callerInfo) FullFunction() string {
	if c == nil {
		return ""
	}
	return c.fn
}

// Package returns the full import path of the package.
func (c *callerInfo) Package() string {
	if c == nil || c.fn == "" || c.dotIdx <= 0 {
		return ""
	}
	return c.fn[:c.dotIdx]
}

// PackageName returns the name of the package without the directory.
// It returns an empty string if the package name cannot be determined.
func (c *callerInfo) PackageName() string {
	pkg := c.Package()
	if pkg == "" {
		return ""
	}
	return filepath.Base(pkg)
}

// String returns a formatted string as returned by ShortLocation().
// It is provided for compatibility with the fmt.Stringer interface.
func (c *callerInfo) String() string {
	return c.ShortLocation()
}

// Equal reports whether this caller is semantically equal to another.
// It ignores cached/internal fields like dotIdx.
// A nil caller is not considered equal to any other caller, including another nil.
func (c *callerInfo) Equal(other Caller) bool {
	// A nil receiver or an untyped nil interface parameter are never equal
	if c == nil || other == nil {
		return false
	}

	// Fast path comparison for callerInfo
	if oc, ok := other.(*callerInfo); ok {
		if oc == nil {
			return false // other is a typed nil
		}
		if c == oc {
			return true // same pointer
		}
		return c.file == oc.file &&
			c.line == oc.line &&
			c.fn == oc.fn
	}

	// Fallback for other implementations of the Caller interface
	return c.file == other.File() &&
		int(c.line) == other.Line() &&
		c.fn == other.FullFunction()
}

// MarshalJSON implements the json.Marshaler interface.
func (c *callerInfo) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}
	return json.Marshal(struct {
		File     string `json:"file,omitempty"`
		Line     int    `json:"line,omitempty"`
		Function string `json:"function,omitempty"`
		Package  string `json:"package,omitempty"`
	}{
		File:     c.file,
		Line:     int(c.line),
		Function: c.Function(),
		Package:  c.Package(),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *callerInfo) UnmarshalJSON(data []byte) error {
	var aux struct {
		File     string `json:"file"`
		Line     int    `json:"line"`
		Function string `json:"function"`
		Package  string `json:"package"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c.file = aux.File

	// Validate and set line
	line, ok := safeUint16(aux.Line)
	if !ok {
		return fmt.Errorf("invalid line number: %d", aux.Line)
	}
	c.line = line

	// Early return if Function is empty
	if aux.Function == "" {
		c.fn = ""
		c.dotIdx = -1
		return nil
	}

	// If package is empty, use only function name
	if aux.Package == "" {
		c.fn = aux.Function
		c.dotIdx = -1
		return nil
	}

	c.fn = aux.Package + "." + aux.Function
	c.dotIdx = functionNameIndex(c.fn)
	return nil
}

// LogValue constructs and returns a slog.Value representing the caller information.
// It includes attributes such as the file name, line number, function name,
// and package if they are available.
// For an invalid or nil caller, it returns an empty slog.Value.
func (c *callerInfo) LogValue() slog.Value {
	if !c.Valid() {
		return slog.Value{}
	}

	attrs := make([]slog.Attr, 0, 4)
	if file := c.File(); file != "" {
		attrs = append(attrs, slog.String("file", file))
		if line := c.Line(); line > 0 {
			attrs = append(attrs, slog.Int("line", line))
		}
	}
	if fn := c.Function(); fn != "" {
		attrs = append(attrs, slog.String("function", fn))
	}
	if pkg := c.Package(); pkg != "" {
		attrs = append(attrs, slog.String("package", pkg))
	}

	return slog.GroupValue(attrs...)
}

// functionNameIndex returns the index of the dot separator
// between the package path and the function name
// in the given full function name.
// For example, if the function name is
// "path/to/package.function", the result is
// the index of the dot (e.g. 17 in this case).
func functionNameIndex(name string) int {
	if name == "" {
		return -1
	}

	// Extract the base name (part after the last slash)
	base := name
	lastSlash := strings.LastIndexByte(name, '/') + 1
	if lastSlash > 0 {
		base = name[lastSlash:]
	}

	// Find the first dot in the base name
	if firstDot := strings.IndexByte(base, '.'); firstDot != -1 {
		return lastSlash + firstDot
	}

	return -1
}

// safeUint16 converts an int to a uint16.
// Returns false if the value is out of range.
func safeUint16(value int) (uint16, bool) {
	if value < 0 || value > int(^uint16(0)) {
		return 0, false
	}
	return uint16(value), true
}
