package pragma

import (
	"reflect"
	"testing"
)

func TestParsePragma(t *testing.T) {
	type args struct{ comment string }

	tests := []struct {
		name string
		args args
		want *Pragma
	}{
		{
			name: "invalid pragma",
			args: args{comment: "foobar"},
			want: nil,
		},
		{
			name: "pragma without column / message",
			args: args{comment: `[GO-W1000]`},
			want: &Pragma{
				Issues: map[string][]*Issue{"GO-W1000": {}},
				Hit:    map[string]bool{"GO-W1000": false},
			},
		},
		{
			name: "pragma with only columns",
			args: args{comment: `[GO-W1000]: 1, 2`},
			want: &Pragma{
				Issues: map[string][]*Issue{
					"GO-W1000": {
						{Column: 1},
						{Column: 2},
					},
				},
				Hit: map[string]bool{"GO-W1000": false},
			},
		},
		{
			name: "pragma with only message",
			args: args{comment: `[GO-W1000]: "hello"`},
			want: &Pragma{
				Issues: map[string][]*Issue{
					"GO-W1000": {{Column: 0, Message: "hello"}},
				},
				Hit: map[string]bool{"GO-W1000": false},
			},
		},
		{
			name: "pragma with both messages and columns",
			args: args{comment: `[GO-W1000]: 1 "hello", 2 "world"`},
			want: &Pragma{
				Issues: map[string][]*Issue{
					"GO-W1000": {
						{Column: 1, Message: "hello"},
						{Column: 2, Message: "world"},
					},
				},
				Hit: map[string]bool{"GO-W1000": false},
			},
		},
		{
			name: "pragmas with mixed messages and columns",
			args: args{comment: `[GO-W1000]: 1 "Hello", 2, "World"`},
			want: &Pragma{
				Issues: map[string][]*Issue{
					"GO-W1000": {
						{Column: 1, Message: "Hello"},
						{Column: 2, Message: ""},
						{Message: "World"},
					},
				},
				Hit: map[string]bool{"GO-W1000": false},
			},
		},
		{
			name: "multiple pragmas",
			args: args{comment: `[GO-W1000]: 1 "Hello", 2, "World"; [GO-W1001]: "Hello"`},
			want: &Pragma{
				Issues: map[string][]*Issue{
					"GO-W1000": {
						{Column: 1, Message: "Hello"},
						{Column: 2, Message: ""},
						{Message: "World"},
					},
					"GO-W1001": {
						{Message: "Hello"},
					},
				},
				Hit: map[string]bool{"GO-W1000": false, "GO-W1001": false},
			},
		},
		{
			name: "multiple pragmas with trailing semi",
			args: args{comment: `[GO-W1000]: 1 "Hello", 2, "World"; [GO-W1001]: "Hello";`},
			want: &Pragma{
				Issues: map[string][]*Issue{
					"GO-W1000": {
						{Column: 1, Message: "Hello"},
						{Column: 2, Message: ""},
						{Message: "World"},
					},
					"GO-W1001": {
						{Message: "Hello"},
					},
				},
				Hit: map[string]bool{"GO-W1000": false, "GO-W1001": false},
			},
		},
		{
			name: "multiple pragmas with last pragma invalid",
			args: args{comment: `[GO-W1000]: 1 "Hello", 2, "World"; [GO-W1001]: "Hello"; invalid pragma?`},
			want: &Pragma{
				Issues: map[string][]*Issue{
					"GO-W1000": {
						{Column: 1, Message: "Hello"},
						{Column: 2, Message: ""},
						{Message: "World"},
					},
					"GO-W1001": {
						{Message: "Hello"},
					},
				},
				Hit: map[string]bool{"GO-W1000": false, "GO-W1001": false},
			},
		},
		{
			name: "multiple pragmas with first pragma invalid",
			args: args{comment: `invalid pragma; [GO-W1000]: 1 "Hello", 2, "World"; [GO-W1001]: "Hello"`},
			want: &Pragma{
				Issues: map[string][]*Issue{
					"GO-W1000": {
						{Column: 1, Message: "Hello"},
						{Column: 2, Message: ""},
						{Message: "World"},
					},
					"GO-W1001": {
						{Message: "Hello"},
					},
				},
				Hit: map[string]bool{"GO-W1000": false, "GO-W1001": false},
			},
		},
		{
			name: "pragma with quote escaping",
			args: args{comment: `[GO-W1000]: 1 "Hello \"World\""; [GO-W1001]: 1 "Foo"`},
			want: &Pragma{
				Issues: map[string][]*Issue{
					"GO-W1000": {{Column: 1, Message: `Hello "World"`}},
					"GO-W1001": {{Column: 1, Message: "Foo"}},
				},
				Hit: map[string]bool{"GO-W1000": false, "GO-W1001": false},
			},
		},
		{
			name: "rust pragma with #[must_use]",
			args: args{comment: "[RS-E1017]: \"Calling `.hash(_)` on expression with unit-type `#[must_use]`\""},
			want: &Pragma{
				Issues: map[string][]*Issue{
					"RS-E1017": {{Message: "Calling `.hash(_)` on expression with unit-type `#[must_use]`"}},
				},
				Hit: map[string]bool{"RS-E1017": false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParsePragma(tt.args.comment); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePragma() = %v, want %v", got, tt.want)
			}
		})
	}
}
