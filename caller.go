/*
Package caller provides utilities to extract source code location
information (file, line, function, and package) for the current
or specified call frame.
It is designed for use in logging, error reporting, and debugging
with a lightweight and idiomatic API. Caller captures runtime metadata
using the Go runtime and formats it in a developer-friendly way.
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
// Returns nil if the skip is invalid or the caller cannot be determined.
func New(skip int) Caller {
	// Check depth and return nil if invalid
	effectiveDepth := skip + skipAdjust
	if effectiveDepth <= 0 {
		return nil
	}

	// Get caller information with the effective depth to skip
	pc, file, line, ok := runtime.Caller(effectiveDepth)
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
// Returns nil if the caller cannot be determined.
func Immediate() Caller {
	return New(0)
}

// NewFromPC returns a new Caller with source information populated
// based on the provided program counter.
// Returns nil if the caller cannot be determined.
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
	return c.file != "" && c.line > 0 && c.fn != ""
}

// File returns the file name.
func (c *callerInfo) File() string {
	return c.file
}

// Line returns the line number.
func (c *callerInfo) Line() int {
	return int(c.line)
}

// Location returns a formatted string with file:line.
func (c *callerInfo) Location() string {
	if c.file == "" {
		return ""
	}
	if c.line == 0 {
		return c.file
	}

	var sb strings.Builder
	sb.WriteString(c.file)
	sb.WriteByte(':')
	sb.Write(strconv.AppendInt(nil, int64(c.line), 10))
	return sb.String()
}

// ShortLocation returns a formatted string with just filename:line.
func (c *callerInfo) ShortLocation() string {
	if c.file == "" {
		return ""
	}

	shortFile := filepath.Base(c.file)
	if c.line == 0 {
		return shortFile
	}

	var sb strings.Builder
	sb.WriteString(shortFile)
	sb.WriteByte(':')
	sb.Write(strconv.AppendInt(nil, int64(c.line), 10))
	return sb.String()
}

// Function returns just the function or method name
// without package prefix.
func (c *callerInfo) Function() string {
	if c.fn == "" || c.dotIdx == -1 {
		return ""
	}
	if c.dotIdx == 0 {
		return c.fn
	}

	// Return a copy of the function name
	return string([]byte(c.fn[c.dotIdx+1:]))
}

// FullFunction returns the full function name including package.
func (c *callerInfo) FullFunction() string {
	return c.fn
}

// Package returns the full import path of the package.
func (c *callerInfo) Package() string {
	if c.fn == "" || c.dotIdx == 0 {
		return ""
	}
	if c.dotIdx == -1 {
		return c.fn
	}

	// Return a copy of the package path
	return string([]byte(c.fn[:c.dotIdx]))
}

// PackageName returns the name of the package without the directory.
func (c *callerInfo) PackageName() string {
	if c.fn == "" || c.dotIdx == 0 {
		return ""
	}

	pkg := c.fn
	if c.dotIdx != -1 {
		pkg = c.fn[:c.dotIdx]
	}

	return filepath.Base(pkg)
}

// String returns a formatted string as returned by ShortLocation().
// It is provided for compatibility with the fmt.Stringer interface.
func (c *callerInfo) String() string {
	return c.ShortLocation()
}

// MarshalJSON implements the json.Marshaler interface.
func (c *callerInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		File     string `json:"file"`
		Line     int    `json:"line"`
		Function string `json:"function"`
		Package  string `json:"package"`
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
// and package if they are available. If no attributes are available, it returns
// an empty slog.Value.
func (c *callerInfo) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, 4)

	if c.file != "" {
		attrs = append(attrs, slog.String("file", c.file))
	}
	if c.line > 0 {
		attrs = append(attrs, slog.Int("line", int(c.line)))
	}
	if fn := c.Function(); fn != "" {
		attrs = append(attrs, slog.String("function", fn))
	}
	if pkg := c.Package(); pkg != "" {
		attrs = append(attrs, slog.String("package", pkg))
	}

	if len(attrs) == 0 {
		return slog.Value{}
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

	// Extract the base name
	base := name
	lastSlash := strings.LastIndexByte(name, '/') + 1
	if lastSlash != 0 {
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
