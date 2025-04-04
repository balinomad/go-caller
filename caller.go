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
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Caller provides access to source information about the caller.
type Caller interface {
	fmt.Stringer

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

	return &callerInfo{
		file:   file,
		line:   uint16(line),
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

	fullFunc = f.Name()
	file, line = f.FileLine(pc)

	return &callerInfo{
		file:   file,
		line:   uint16(line),
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
	sb.WriteString(strconv.Itoa(int(c.line)))
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
	sb.WriteString(strconv.Itoa(int(c.line)))
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
	return c.fn[c.dotIdx+1:]
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
	return c.fn[:c.dotIdx]
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

	// Extract the package path and base name
	dir, base := filepath.Split(name)
	if base == "" {
		return -1
	}

	// Find the first dot in the base name
	firstDot := strings.Index(base, ".")
	if firstDot != -1 {
		return len(dir) + firstDot
	}
	return -1
}
