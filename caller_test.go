package caller

import (
	"testing"
)

func Test_callerInfo_Function(t *testing.T) {
	type fields struct {
		fn string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty name",
			fields: fields{
				fn: "",
			},
			want: "",
		},
		{
			name: "function without directory",
			fields: fields{
				fn: "package.FunctionName",
			},
			want: "FunctionName",
		},
		{
			name: "method without directory",
			fields: fields{
				fn: "package.(*Type).MethodName",
			},
			want: "(*Type).MethodName",
		},
		{
			name: "package function with Unix-style path",
			fields: fields{
				fn: "github.com/user/package.FunctionName",
			},
			want: "FunctionName",
		},
		{
			name: "package method with Unix-style path",
			fields: fields{
				fn: "github.com/user/package.(*Type).MethodName",
			},
			want: "(*Type).MethodName",
		},
		{
			name: "package function with Windows-style path",
			fields: fields{
				fn: "C:\\user\\package.FunctionName",
			},
			want: "FunctionName",
		},
		{
			name: "package method with Windows-style path",
			fields: fields{
				fn: "C:\\user\\package.(*Type).MethodName",
			},
			want: "(*Type).MethodName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &callerInfo{
				fn:     tt.fields.fn,
				dotIdx: functionNameIndex(tt.fields.fn),
			}
			if got := c.Function(); got != tt.want {
				t.Errorf("caller.Function() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_callerInfo_Package(t *testing.T) {
	type fields struct {
		fn string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty name",
			fields: fields{
				fn: "",
			},
			want: "",
		},
		{
			name: "function without directory",
			fields: fields{
				fn: "package.FunctionName",
			},
			want: "package",
		},
		{
			name: "method without directory",
			fields: fields{
				fn: "package.(*Type).MethodName",
			},
			want: "package",
		},
		{
			name: "package function with Unix-style path",
			fields: fields{
				fn: "github.com/user/package.FunctionName",
			},
			want: "github.com/user/package",
		},
		{
			name: "package method with Unix-style path",
			fields: fields{
				fn: "github.com/user/package.(*Type).MethodName",
			},
			want: "github.com/user/package",
		},
		{
			name: "package function with Windows-style path",
			fields: fields{
				fn: "c:\\user\\package.FunctionName",
			},
			want: "c:\\user\\package",
		},
		{
			name: "package method with Windows-style path",
			fields: fields{
				fn: "c:\\user\\package.(*Type).MethodName",
			},
			want: "c:\\user\\package",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &callerInfo{
				fn:     tt.fields.fn,
				dotIdx: functionNameIndex(tt.fields.fn),
			}
			if got := c.Package(); got != tt.want {
				t.Errorf("caller.Package() = %v, want %v", got, tt.want)
			}
		})
	}
}
