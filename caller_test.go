package caller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// testFunc is a helper to get a caller at a known stack frame.
func testFunc() Caller {
	// The immediate caller of New(0) is this function's caller
	return New(0)
}

// nestedTestFunc is a helper to test stack skipping.
func nestedTestFunc(skip int) Caller {
	return New(skip)
}

// TestNew tests the New function and verifies that it correctly
// captures the caller information from the specified stack frame.
// It tests both immediate callers and callers at an arbitrary
// distance from the current frame.
func TestNew(t *testing.T) {
	t.Run("immediate caller", func(t *testing.T) {
		c := testFunc() // This line is where the call is made
		if c == nil {
			t.Fatal("New(0) from testFunc returned nil")
		}

		_, file, line, ok := runtime.Caller(0)
		if !ok {
			t.Fatal("runtime.Caller(0) failed")
		}
		expectedLine := line - 5 // The testFunc() call is 5 lines above runtime.Caller(0)

		if got := c.File(); got != file {
			t.Errorf("File() = %q, want %q", got, file)
		}
		if got := c.Line(); got != expectedLine {
			t.Errorf("Line() = %d, want %d", got, expectedLine)
		}
		if got := c.Function(); got != "TestNew.func1" {
			t.Errorf("Function() = %q, want %q", got, "TestNew.func1")
		}
	})

	t.Run("skip caller", func(t *testing.T) {
		// The caller of nestedTestFunc is this function.
		c := nestedTestFunc(0)
		if c == nil {
			t.Fatal("New(0) from nested func returned nil")
		}
		if got := c.Function(); got != "TestNew.func2" {
			t.Errorf("Function() = %q, want %q", got, "TestNew.func2")
		}

		// The caller of the caller of nestedTestFunc is t.Run's tRunner.
		c = nestedTestFunc(1)
		if c == nil {
			t.Fatal("New(1) from nested func returned nil")
		}
		if !strings.Contains(c.Function(), "tRunner") { // tRunner is an internal testing function
			t.Logf("Function() = %q (this might change with Go versions)", c.Function())
		}
	})
}

// TestNewWithInvalidSkip tests the New function with invalid skip values.
// It verifies that New correctly returns nil for invalid skips.
func TestNewWithInvalidSkip(t *testing.T) {
	tests := []struct {
		name string
		skip int
		want *callerInfo
	}{
		{"invalid skip -1", -1, nil},
		{"invalid skip -100", -100, nil},
		{"large skip", 10000, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.skip)
			if tt.want == nil && c != nil {
				t.Errorf("New(%d) = %v, want nil", tt.skip, c)
			}
			if tt.want != nil && c == nil {
				t.Errorf("New(%d) = nil, want %v", tt.skip, tt.want)
			}
		})
	}
}

// TestImmediate tests Immediate() by calling it and validating the resulting
// Caller against the runtime.Caller(0) information. It checks that the file,
// line, and function names match.
func TestImmediate(t *testing.T) {
	c := Immediate() // This line is where the call is made
	if c == nil {
		t.Fatal("Immediate() returned nil")
	}

	_, file, line, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	expectedLine := line - 5 // // The Immediate() call is 5 lines above runtime.Caller(0)

	if got := c.File(); got != file {
		t.Errorf("File() = %q, want %q", got, file)
	}
	if got := c.Line(); got != expectedLine {
		t.Errorf("Line() = %d, want %d", got, expectedLine)
	}
	if got := c.Function(); got != "TestImmediate" {
		t.Errorf("Function() = %q, want %q", got, "TestImmediate")
	}
}

// TestNewFromPC tests the NewFromPC function and verifies that it
// correctly captures the caller information based on the provided
// program counter. It tests both valid and invalid PCs.
func TestNewFromPC(t *testing.T) {
	t.Run("valid pc", func(t *testing.T) {
		pc, file, line, ok := runtime.Caller(0)
		if !ok {
			t.Fatal("runtime.Caller(0) failed")
		}

		c := NewFromPC(pc)
		if c == nil {
			t.Fatal("NewFromPC returned nil for a valid PC")
		}

		if got := c.File(); got != file {
			t.Errorf("File() = %q, want %q", got, file)
		}
		if got := c.Line(); got != line {
			t.Errorf("Line() = %d, want %d", got, line)
		}
		if got := c.Function(); got != "TestNewFromPC.func1" {
			t.Errorf("Function() = %q, want %q", got, "TestNewFromPC.func1")
		}
	})

	t.Run("invalid pc", func(t *testing.T) {
		if c := NewFromPC(0); c != nil {
			t.Errorf("NewFromPC(0) should return nil, but got %v", c)
		}
		// A PC that is highly unlikely to map to a valid function.
		if c := NewFromPC(uintptr(0xDEADBEEF)); c != nil {
			t.Errorf("NewFromPC with arbitrary PC should return nil, but got %v", c)
		}
	})
}

// mockCaller is a mock implementation of the Caller interface for testing Equal.
type mockCaller struct {
	file   string
	line   int
	fn     string
	fullFn string
}

func (m *mockCaller) Valid() bool                  { return m.file != "" }
func (m *mockCaller) File() string                 { return m.file }
func (m *mockCaller) Line() int                    { return m.line }
func (m *mockCaller) Location() string             { return fmt.Sprintf("%s:%d", m.file, m.line) }
func (m *mockCaller) ShortLocation() string        { return m.Location() }
func (m *mockCaller) Function() string             { return m.fn }
func (m *mockCaller) FullFunction() string         { return m.fullFn }
func (m *mockCaller) Package() string              { return "pkg" }
func (m *mockCaller) PackageName() string          { return "pkg" }
func (m *mockCaller) String() string               { return m.ShortLocation() }
func (m *mockCaller) MarshalJSON() ([]byte, error) { return nil, nil }
func (m *mockCaller) UnmarshalJSON(b []byte) error { return nil }
func (m *mockCaller) LogValue() slog.Value         { return slog.Value{} }
func (m *mockCaller) Equal(other Caller) bool {
	if other == nil {
		return false
	}
	return m.File() == other.File() &&
		m.Line() == other.Line() &&
		m.FullFunction() == other.FullFunction()
}

// TestCallerInfo_Equal tests the Equal method of callerInfo.
// It checks the comparison of equal and non-equal values, including
// nil interface values and concrete types, as well as different types.
func TestCallerInfo_Equal(t *testing.T) {
	base := &callerInfo{file: "main.go", line: 10, fn: "main.main", dotIdx: 4}
	baseIdentical := &callerInfo{file: "main.go", line: 10, fn: "main.main", dotIdx: 99}
	baseDiffFile := &callerInfo{file: "other.go", line: 10, fn: "main.main", dotIdx: 4}
	baseDiffLine := &callerInfo{file: "main.go", line: 20, fn: "main.main", dotIdx: 4}
	baseDiffFunc := &callerInfo{file: "main.go", line: 10, fn: "main.other", dotIdx: 4}
	mockNotEqual := &mockCaller{file: "main.go", line: 10, fullFn: "pkg.main"}
	mockEqual := &mockCaller{file: "main.go", line: 10, fullFn: "main.main"}

	tests := []struct {
		name  string
		c     *callerInfo
		other Caller
		equal bool
	}{
		{"c against nil interface", base, nil, false},
		{"nil concrete type against c", (*callerInfo)(nil), base, false},
		{"c against nil concrete type", base, (*callerInfo)(nil), false},
		{"two nil concrete types", (*callerInfo)(nil), (*callerInfo)(nil), false},
		{"two empty concrete types", &callerInfo{}, &callerInfo{}, true},
		{"identical values", base, baseIdentical, true},
		{"same pointer", base, base, true},
		{"different file", base, baseDiffFile, false},
		{"different line", base, baseDiffLine, false},
		{"different function", base, baseDiffFunc, false},
		{"different type, not equal", base, mockNotEqual, false},
		{"different type, equal", base, mockEqual, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c != nil {
				tt.c.dotIdx = functionNameIndex(tt.c.fn)
			}
			if oc, ok := tt.other.(*callerInfo); ok && oc != nil {
				oc.dotIdx = functionNameIndex(oc.fn)
			}
			if got := tt.c.Equal(tt.other); got != tt.equal {
				t.Errorf("a.Equal(b) = %v, want %v", got, tt.equal)
			}
			// Test commutativity, but don't call method on a nil interface value
			if tt.other != nil {
				if got := tt.other.Equal(tt.c); got != tt.equal {
					t.Errorf("b.Equal(a) commutativity failed: got %v, want %v", got, tt.equal)
				}
			}
			if got := tt.c.Equal(tt.other); got != tt.equal {
				t.Errorf("a.Equal(b) = %v, want %v", got, tt.equal)
			}
			// Test commutativity, but don't call method on a nil interface value
			if tt.other != nil {
				if got := tt.other.Equal(tt.c); got != tt.equal {
					t.Errorf("b.Equal(a) commutativity failed: got %v, want %v", got, tt.equal)
				}
			}
		})
	}
}

// TestCallerInfo_Valid tests the Valid method of callerInfo, ensuring it
// correctly identifies valid and invalid callerInfo values.
func TestCallerInfo_Valid(t *testing.T) {
	tests := []struct {
		name string
		c    *callerInfo
		want bool
	}{
		{"nil receiver", nil, false},
		{"zero value caller", &callerInfo{}, false},
		{"valid", &callerInfo{file: "main.go", line: 1, fn: "main.main"}, true},
		{"no file", &callerInfo{line: 1, fn: "main.main"}, false},
		{"no line", &callerInfo{file: "main.go", fn: "main.main"}, true},
		{"no function", &callerInfo{file: "main.go", line: 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c != nil {
				tt.c.dotIdx = functionNameIndex(tt.c.fn)
			}
			if got := tt.c.Valid(); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCallerInfo_File tests the File method of callerInfo, ensuring it
// correctly extracts the file name from a valid callerInfo value, and
// returns an empty string for invalid values.
func TestCallerInfo_File(t *testing.T) {
	tests := []struct {
		name string
		c    *callerInfo
		want string
	}{
		{"nil receiver", nil, ""},
		{"zero value caller", &callerInfo{}, ""},
		{"valid", &callerInfo{file: "main.go"}, "main.go"},
		{"no file", &callerInfo{line: 42}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c != nil {
				tt.c.dotIdx = functionNameIndex(tt.c.fn)
			}
			if got := tt.c.File(); got != tt.want {
				t.Errorf("File() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCallerInfo_Line tests the Line method of callerInfo, ensuring it
// correctly extracts the line number from a valid callerInfo value, and
// returns 0 for invalid values.
func TestCallerInfo_Line(t *testing.T) {
	tests := []struct {
		name string
		c    *callerInfo
		want int
	}{
		{"nil receiver", nil, 0},
		{"zero value caller", &callerInfo{}, 0},
		{"valid", &callerInfo{line: 42}, 42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c != nil {
				tt.c.dotIdx = functionNameIndex(tt.c.fn)
			}
			if got := tt.c.Line(); got != tt.want {
				t.Errorf("Line() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestCallerInfo_Location tests the Location method of callerInfo, ensuring it
// correctly formats strings with file:line.
func TestCallerInfo_Location(t *testing.T) {
	tests := []struct {
		name string
		c    *callerInfo
		want string
	}{
		{"nil receiver", nil, ""},
		{"zero value caller", &callerInfo{}, ""},
		{"valid", &callerInfo{file: "/path/to/main.go", line: 42}, "/path/to/main.go:42"},
		{"no file", &callerInfo{line: 42}, ""},
		{"no line", &callerInfo{file: "/path/to/main.go"}, "/path/to/main.go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c != nil {
				tt.c.dotIdx = functionNameIndex(tt.c.fn)
			}
			if got := tt.c.Location(); got != tt.want {
				t.Errorf("Location() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCallerInfo_ShortLocationAndString tests the ShortLocation and String methods of callerInfo,
// ensuring they correctly format strings with file:line or just filename.
func TestCallerInfo_ShortLocationAndString(t *testing.T) {
	tests := []struct {
		name string
		c    *callerInfo
		want string
	}{
		{"nil receiver", nil, ""},
		{"zero value caller", &callerInfo{}, ""},
		{"valid", &callerInfo{file: "/path/to/main.go", line: 42}, "main.go:42"},
		{"no file", &callerInfo{line: 42}, ""},
		{"no line", &callerInfo{file: "/path/to/main.go"}, "main.go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c != nil {
				tt.c.dotIdx = functionNameIndex(tt.c.fn)
			}
			if got := tt.c.ShortLocation(); got != tt.want {
				t.Errorf("ShortLocation() = %q, want %q", got, tt.want)
			}
			if got := tt.c.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCallerInfo_Function tests the Function method of callerInfo, ensuring it correctly extracts
// the function name from a full function name string, including handling of empty names,
// function names without packages, method names on types, full path function names, full path
// method names, function names with no package, and function names with a dot prefix.
func TestCallerInfo_Function(t *testing.T) {
	tests := []struct {
		name string
		c    *callerInfo
		want string
	}{
		{"nil receiver", nil, ""},
		{"zero value caller", &callerInfo{}, ""},
		{"function without package", &callerInfo{fn: "main"}, ""},
		{"function with package", &callerInfo{fn: "pkg.Func"}, "Func"},
		{"method on type", &callerInfo{fn: "pkg.(*Type).Method"}, "(*Type).Method"},
		{"full path function", &callerInfo{fn: "github.com/user/repo.Func"}, "Func"},
		{"full path method", &callerInfo{fn: "github.com/user/repo.(*Type).Method"}, "(*Type).Method"},
		{"no function name", &callerInfo{fn: "pkg."}, ""},
		{"dot prefix", &callerInfo{fn: ".Func"}, "Func"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c != nil {
				tt.c.dotIdx = functionNameIndex(tt.c.fn)
			}
			if got := tt.c.Function(); got != tt.want {
				t.Errorf("Function() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCallerInfo_FullFunction(t *testing.T) {
	tests := []struct {
		name string
		c    *callerInfo
		want string
	}{
		{"nil receiver", nil, ""},
		{"zero value caller", &callerInfo{}, ""},
		{"function without package", &callerInfo{fn: "main"}, "main"},
		{"function with package", &callerInfo{fn: "pkg.Func"}, "pkg.Func"},
		{"method on type", &callerInfo{fn: "pkg.(*Type).Method"}, "pkg.(*Type).Method"},
		{"full path function", &callerInfo{fn: "github.com/user/repo.Func"}, "github.com/user/repo.Func"},
		{"full path method", &callerInfo{fn: "github.com/user/repo.(*Type).Method"}, "github.com/user/repo.(*Type).Method"},
		{"no function name", &callerInfo{fn: "pkg."}, "pkg."},
		{"dot prefix", &callerInfo{fn: ".Func"}, ".Func"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c != nil {
				tt.c.dotIdx = functionNameIndex(tt.c.fn)
			}
			if got := tt.c.FullFunction(); got != tt.want {
				t.Errorf("FullFunction() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCallerInfo_Package tests the Package method of callerInfo, ensuring it correctly extracts
// the package name from a full function name string, including handling of empty names,
// function names without packages, method names on types, full path function names, full path
// method names, function names with no package, and function names with a dot prefix.
func TestCallerInfo_Package(t *testing.T) {
	tests := []struct {
		name string
		c    *callerInfo
		want string
	}{
		{"empty name", &callerInfo{fn: ""}, ""},
		{"nil receiver", nil, ""},
		{"function without package", &callerInfo{fn: "main"}, ""}, // no dot
		{"function with package", &callerInfo{fn: "pkg.Func"}, "pkg"},
		{"method on type", &callerInfo{fn: "pkg.(*Type).Method"}, "pkg"},
		{"full path function", &callerInfo{fn: "github.com/user/repo.Func"}, "github.com/user/repo"},
		{"full path method", &callerInfo{fn: "github.com/user/repo.(*Type).Method"}, "github.com/user/repo"},
		{"no function name", &callerInfo{fn: "pkg."}, "pkg"},
		{"no package name", &callerInfo{fn: ".Func"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c != nil {
				tt.c.dotIdx = functionNameIndex(tt.c.fn)
			}
			if got := tt.c.Package(); got != tt.want {
				t.Errorf("Package() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCallerInfo_PackageName tests the PackageName method of callerInfo, ensuring it correctly
// extracts the last element of the package path from a full function name string,
// including handling of empty names, function names without packages, method names on types,
// full path function names, full path method names, function names with no package,
// and function names with a dot prefix.
func TestCallerInfo_PackageName(t *testing.T) {
	tests := []struct {
		name string
		c    *callerInfo
		want string
	}{
		{"empty name", &callerInfo{fn: ""}, ""},
		{"nil receiver", nil, ""},
		{"function with package", &callerInfo{fn: "pkg.Func"}, "pkg"},
		{"full path function", &callerInfo{fn: "github.com/user/repo.Func"}, "repo"},
		{"full path method", &callerInfo{fn: "github.com/user/repo.(*Type).Method"}, "repo"},
		{"no function name", &callerInfo{fn: "pkg."}, "pkg"},
		{"no package", &callerInfo{fn: "main"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c != nil {
				tt.c.dotIdx = functionNameIndex(tt.c.fn)
			}
			if got := tt.c.PackageName(); got != tt.want {
				t.Errorf("PackageName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCallerInfo_MarshalJSON tests the MarshalJSON method of callerInfo,
// ensuring it correctly marshals a callerInfo object to JSON,
// including handling of omitempty fields.
func TestCallerInfo_MarshalJSON(t *testing.T) {
	t.Run("valid caller", func(t *testing.T) {
		c := &callerInfo{
			file:   "test.go",
			line:   123,
			fn:     "my/pkg.MyFunc",
			dotIdx: functionNameIndex("my/pkg.MyFunc"),
		}
		b, err := json.Marshal(c)
		if err != nil {
			t.Fatalf("MarshalJSON() error = %v", err)
		}
		want := `{"file":"test.go","line":123,"function":"MyFunc","package":"my/pkg"}`
		if string(b) != want {
			t.Errorf("MarshalJSON() = %s, want %s", b, want)
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var c *callerInfo
		b, err := c.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON(nil) error = %v", err)
		}
		want := `null`
		if string(b) != want {
			t.Errorf("MarshalJSON(nil) = %s, want %s", b, want)
		}
	})
}

// TestCallerInfo_UnmarshalJSON tests the UnmarshalJSON method of callerInfo,
// ensuring it correctly unmarshals JSON into a callerInfo object,
// including handling of omitempty fields and invalid data.
func TestCallerInfo_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		want      *callerInfo
		expectErr bool
	}{
		{
			name:      "valid full",
			jsonData:  `{"file":"test.go","line":123,"function":"MyFunc","package":"my/pkg"}`,
			want:      &callerInfo{file: "test.go", line: 123, fn: "my/pkg.MyFunc"},
			expectErr: false,
		},
		{
			name:      "valid no package",
			jsonData:  `{"file":"test.go","line":123,"function":"MyFunc"}`,
			want:      &callerInfo{file: "test.go", line: 123, fn: "MyFunc"},
			expectErr: false,
		},
		{
			name:      "valid no function",
			jsonData:  `{"file":"test.go","line":123,"package":"my/pkg"}`,
			want:      &callerInfo{file: "test.go", line: 123, fn: ""},
			expectErr: false,
		},
		{
			name:      "valid with empty function string",
			jsonData:  `{"file":"test.go","line":123,"package":"my/pkg","function":""}`,
			want:      &callerInfo{file: "test.go", line: 123, fn: ""},
			expectErr: false,
		},
		{
			name:      "invalid json",
			jsonData:  `{`,
			want:      &callerInfo{},
			expectErr: true,
		},
		{
			name:      "invalid line",
			jsonData:  `{"line":-1}`,
			want:      &callerInfo{},
			expectErr: true,
		},
		{
			name:      "line too large",
			jsonData:  `{"line":65536}`,
			want:      &callerInfo{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got callerInfo
			err := got.UnmarshalJSON([]byte(tc.jsonData))

			if tc.expectErr {
				if err == nil {
					t.Error("expected an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Re-calc dotIdx for comparison
			tc.want.dotIdx = functionNameIndex(tc.want.fn)
			if !got.Equal(tc.want) {
				t.Errorf("UnmarshalJSON() got = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestCallerInfo_LogValue tests the LogValue method of callerInfo, ensuring it
// correctly converts a callerInfo object into a slog.Value representing the
// caller information. It includes attributes such as the file name, line
// number, function name, and package if they are available.
func TestCallerInfo_LogValue(t *testing.T) {
	testCases := []struct {
		name   string
		caller *callerInfo
		want   slog.Value
	}{
		{
			name:   "nil receiver",
			caller: nil,
			want:   slog.Value{},
		},
		{
			name:   "zero value caller",
			caller: &callerInfo{},
			want:   slog.Value{},
		},
		{
			name: "valid caller",
			caller: &callerInfo{
				file:   "/path/to/main.go",
				line:   10,
				fn:     "proj.main",
				dotIdx: 4,
			},
			want: slog.GroupValue(
				slog.String("file", "/path/to/main.go"),
				slog.Int("line", 10),
				slog.String("function", "main"),
				slog.String("package", "proj"),
			),
		},
		{
			name: "partial caller - file and line only",
			caller: &callerInfo{
				file: "main.go",
				line: 10,
			},
			want: slog.GroupValue(
				slog.String("file", "main.go"),
				slog.Int("line", 10),
			),
		},
		{
			name: "partial caller - file only",
			caller: &callerInfo{
				file: "main.go",
			},
			want: slog.GroupValue(
				slog.String("file", "main.go"),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.caller.LogValue()

			if tc.want.Kind() == slog.KindAny && tc.want.Any() == nil {
				if got.Kind() != slog.KindAny || got.Any() != nil {
					t.Errorf("LogValue() got = %v, want empty slog.Value", got)
				}
				return
			}

			if got.Kind() != tc.want.Kind() {
				t.Errorf("LogValue() kind mismatch: got %v, want %v", got.Kind(), tc.want.Kind())
			}

			// Compare attributes for GroupValue
			if got.Kind() == slog.KindGroup {
				gotAttrs := got.Group()
				wantAttrs := tc.want.Group()
				if len(gotAttrs) != len(wantAttrs) {
					t.Errorf("LogValue() attribute count mismatch: got %d, want %d.\nGot: %v\nWant: %v", len(gotAttrs), len(wantAttrs), gotAttrs, wantAttrs)
					return
				}
				for i := range gotAttrs {
					if !gotAttrs[i].Equal(wantAttrs[i]) {
						t.Errorf("LogValue() attribute mismatch at index %d: got %v, want %v", i, gotAttrs[i], wantAttrs[i])
					}
				}
			}
		})
	}
}

func TestFunctionNameIndex(t *testing.T) {
	tests := []struct {
		name string
		fn   string
		want int
	}{
		{"empty", "", -1},
		{"no package", "main", -1},
		{"simple package", "pkg.Func", 3},
		{"path package", "path/to/pkg.Func", 11},
		{"receiver", "path/to/pkg.(*Type).Func", 11},
		{"vendored", "vendor/path/to/pkg.Func", 18},
		{"no function name", "path/to/pkg.", 11}, // dot is last char
		{"dot prefix", ".Func", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := functionNameIndex(tt.fn); got != tt.want {
				t.Errorf("functionNameIndex(%q) = %v, want %v", tt.fn, got, tt.want)
			}
		})
	}
}

func TestSafeUint16(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  uint16
		ok    bool
	}{
		{"zero", 0, 0, true},
		{"max", 65535, 65535, true},
		{"negative", -1, 0, false},
		{"too large", 65536, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := safeUint16(tt.value)
			if got != tt.want || ok != tt.ok {
				t.Errorf("safeUint16(%d) = %v, %v, want %v, %v", tt.value, got, ok, tt.want, tt.ok)
			}
		})
	}
}

// --- Benchmarks ---

var (
	// Prevent compiler optimization by storing result in a global variable.
	globalCaller Caller
	globalFile   string
	globalLine   int
	globalFn     string
	globalString string
)

func BenchmarkCallerOverhead(b *testing.B) {
	var (
		c    Caller
		pc   uintptr
		file string
		line int
		fn   string
		ok   bool
	)

	b.Run("without caller", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			pc, file, line, ok = runtime.Caller(0)
			if ok {
				if f := runtime.FuncForPC(pc); f != nil {
					fn = f.Name()
				}
			}
		}
		// Use the results to avoid optimization
		globalFile, globalLine, globalFn = file, line, fn
	})

	b.Run("with caller", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			c = New(0)
			file, line, fn = c.File(), c.Line(), c.Function()
		}
		// Use the results to avoid optimization
		globalCaller, globalFile, globalLine, globalFn = c, file, line, fn
	})
}

// BenchmarkStringOperations benchmarks different approaches to building a string in the
// Location() method.
func BenchmarkStringOperations(b *testing.B) {
	c := &callerInfo{file: "/some/very/long/path/to/a/file/name.go", line: 12345}

	b.Run("string concat", func(b *testing.B) {
		b.ReportAllocs()
		var s string
		for i := 0; i < b.N; i++ {
			s = c.file + ":" + strconv.Itoa(int(c.line))
		}
		// Use the result to avoid optimization
		globalString = s
	})

	b.Run("builder with append", func(b *testing.B) {
		b.ReportAllocs()
		var s string
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.WriteString(c.file)
			sb.WriteByte(':')
			sb.Write(strconv.AppendInt(nil, int64(c.line), 10))
			s = sb.String()
		}
		// Use the result to avoid optimization
		globalString = s
	})

	b.Run("builder with itoa", func(b *testing.B) {
		b.ReportAllocs()
		var s string
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.WriteString(c.file)
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(int(c.line)))
			s = sb.String()
		}
		// Use the result to avoid optimization
		globalString = s
	})

	b.Run("builder with grow", func(b *testing.B) {
		b.ReportAllocs()
		var s string
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.Grow(len(c.file) + 10)
			sb.WriteString(c.file)
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(c.Line()))
			s = sb.String()
		}
		// Use the result to avoid optimization
		globalString = s
	})

	b.Run("builder proper", func(b *testing.B) {
		b.ReportAllocs()
		var s string
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			line := strconv.Itoa(c.Line())
			sb.Grow(len(c.file) + len(line) + 1)
			sb.WriteString(c.file)
			sb.WriteByte(':')
			sb.WriteString(line)
			s = sb.String()
		}
		// Use the result to avoid optimization
		globalString = s
	})
}
