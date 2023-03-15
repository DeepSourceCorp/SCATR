package pragma

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewFile(t *testing.T) {
	type args struct {
		content       string
		commentPrefix []string
	}
	tests := []struct {
		name string
		args args
		want map[int]*Pragma
	}{
		{
			name: "no pragmas - go",
			args: args{
				content: `
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello World") // something just like this
}
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{},
		},
		{
			name: "no message or column - go",
			args: args{
				content: `
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello World") // [GO-W1000]
}
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				9: {
					Issues: map[string][]*Issue{"GO-W1000": {}},
					Hit:    map[string]bool{"GO-W1000": false},
				},
			},
		},
		{
			name: "mixed messages and columns - go",
			args: args{
				content: `
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello World") // [GO-W1000]: 10 "Hello", 20, "World"
}
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				9: {
					Issues: map[string][]*Issue{
						"GO-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
							{Message: "World", Column: 0},
						},
					},
					Hit: map[string]bool{"GO-W1000": false},
				},
			},
		},
		{
			name: "mixed messages and columns on previous line - go",
			args: args{
				content: `
package main

import (
	"fmt"
)

func main() {
	// [GO-W1000]: 10 "Hello", 20, "World"
	fmt.Println("Hello World")
}
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				10: {
					Issues: map[string][]*Issue{
						"GO-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
							{Message: "World", Column: 0},
						},
					},
					Hit: map[string]bool{"GO-W1000": false},
				},
			},
		},
		{
			name: "no pragmas - python",
			args: args{
				content: `
print("Hello World") # something just like this
`,
				commentPrefix: []string{"#"},
			},
			want: map[int]*Pragma{},
		},
		{
			name: "no message or column - python",
			args: args{
				content: `
print("Hello World") # [PY-W1000]
`,
				commentPrefix: []string{"#"},
			},
			want: map[int]*Pragma{
				2: {
					Issues: map[string][]*Issue{"PY-W1000": {}},
					Hit:    map[string]bool{"PY-W1000": false},
				},
			},
		},
		{
			name: "mixed messages and columns - python",
			args: args{
				content: `
print("Hello World") # [PY-W1000]: 10 "Hello", 20, "World"
`,
				commentPrefix: []string{"#"},
			},
			want: map[int]*Pragma{
				2: {
					Issues: map[string][]*Issue{
						"PY-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
							{Message: "World", Column: 0},
						},
					},
					Hit: map[string]bool{"PY-W1000": false},
				},
			},
		},
		{
			name: "mixed messages and columns on previous line - python",
			args: args{
				content: `
# [PY-W1000]: 10 "Hello", 20, "World"
print("Hello World")
`,
				commentPrefix: []string{"#"},
			},
			want: map[int]*Pragma{
				3: {
					Issues: map[string][]*Issue{
						"PY-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
							{Message: "World", Column: 0},
						},
					},
					Hit: map[string]bool{"PY-W1000": false},
				},
			},
		},
		{
			name: "multi line pragma",
			args: args{
				content: `
class C2 : public B {
public:
  // [CXX-W2008]: 34 "Grand-parent method"
  // [CXX-W2008]: 49 "Grand-parent method"
  int virt_1() override { return A1::virt_1() + A2::virt_1() + A3::virt_1(); }
  // CHECK-FIXES:  int virt_1() override { return B::virt_1() + B::virt_1() + B::virt_1(); }
};
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				6: {
					Issues: map[string][]*Issue{
						"CXX-W2008": {
							{Column: 49, Message: "Grand-parent method"},
							{Column: 34, Message: "Grand-parent method"},
						},
					},
					Hit: map[string]bool{"CXX-W2008": false},
				},
			},
		},
		{
			name: "pragmas split on multiple lines",
			args: args{
				content: `
// [GO-W1000]: 10 "Hello", 20; [GO-W1001]
// [GO-W1002]: 30
fmt.Println("Hello")
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				4: {
					Issues: map[string][]*Issue{
						"GO-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
						},
						"GO-W1001": {},
						"GO-W1002": {
							{Message: "", Column: 30},
						},
					},
					Hit: map[string]bool{
						"GO-W1000": false,
						"GO-W1001": false,
						"GO-W1002": false,
					},
				},
			},
		},
		{
			name: "pragmas split on multiple lines - same line pragma",
			args: args{
				content: `
// [GO-W1000]: 10 "Hello", 20; [GO-W1001]
fmt.Println("Hello") // [GO-W1002]: 30
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				3: {
					Issues: map[string][]*Issue{
						"GO-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
						},
						"GO-W1001": {},
						"GO-W1002": {
							{Message: "", Column: 30},
						},
					},
					Hit: map[string]bool{
						"GO-W1000": false,
						"GO-W1001": false,
						"GO-W1002": false,
					},
				},
			},
		},
		{
			name: "pragmas split on multiple lines - edge case",
			args: args{
				content: `
// [GO-W1000]: 10 "Hello", 20; [GO-W1001]
fmt.Println("Hello") // [GO-W1002]: 30
fmt.Println("Hello") // [GO-W1003]: 30
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				3: {
					Issues: map[string][]*Issue{
						"GO-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
						},
						"GO-W1001": {},
						"GO-W1002": {
							{Message: "", Column: 30},
						},
					},
					Hit: map[string]bool{
						"GO-W1000": false,
						"GO-W1001": false,
						"GO-W1002": false,
					},
				},
				4: {
					Issues: map[string][]*Issue{
						"GO-W1003": {
							{Message: "", Column: 30},
						},
					},
					Hit: map[string]bool{
						"GO-W1003": false,
					},
				},
			},
		},
		{
			name: "multiple prefixes with mixed messages and columns with multiple messages - vue",
			args: args{
				content: `
<template>
	<h1>Hello Vue</h1>
	<!-- [VUE-W1000]: 10 "Hello", 20, "World"; [VUE-W1002]: 20 "Vue" -->
</template>
<script>
	// [JS-W1000]: 10 "Hello", 20; [JS-W1002]: "Something"; [JS-W1003]
</script>
`,
				commentPrefix: []string{"<!--", "//"},
			},
			want: map[int]*Pragma{
				5: {
					Issues: map[string][]*Issue{
						"VUE-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
							{Message: "World", Column: 0},
						},
						"VUE-W1002": {
							{Message: "Vue", Column: 20},
						},
					},
					Hit: map[string]bool{
						"VUE-W1000": false,
						"VUE-W1002": false,
					},
				},
				8: {
					Issues: map[string][]*Issue{
						"JS-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
						},
						"JS-W1002": {
							{Message: "Something", Column: 0},
						},
						"JS-W1003": {},
					},
					Hit: map[string]bool{
						"JS-W1000": false,
						"JS-W1002": false,
						"JS-W1003": false,
					},
				},
			},
		},
		{
			name: "issue collapse bug - cxx",
			args: args{
				content: `issueCaseNotRaised(00); // [CXX-S111]
issueCaseRaised(010); // [CXX-S112]`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				1: {
					Issues: map[string][]*Issue{"CXX-S111": {}},
					Hit:    map[string]bool{"CXX-S111": false},
				},
				2: {
					Issues: map[string][]*Issue{"CXX-S112": {}},
					Hit:    map[string]bool{"CXX-S112": false},
				},
			},
		},
		{
			name: "issue collapse - blank line",
			args: args{
				content: `code();

// [CXX-S112]
issueCaseRaised(010);`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				4: {
					Issues: map[string][]*Issue{"CXX-S112": {}},
					Hit:    map[string]bool{"CXX-S112": false},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFile("", tt.args.content, tt.args.commentPrefix); !reflect.DeepEqual(got.Pragmas, tt.want) {
				t.Errorf("NewFile() = %v, want %v, diff %v", got.Pragmas, tt.want,
					cmp.Diff(tt.want, got.Pragmas))
			}
		})
	}
}

func TestNewFile__CheckMode(t *testing.T) {
	type args struct {
		name          string
		content       string
		commentPrefix []string
	}
	type want struct {
		checkMode  CheckMode
		issueCodes []string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "check all issues - go",
			args: args{
				name: "file.go",
				content: `package main

func main() {}`,
				commentPrefix: []string{"//"},
			},
			want: want{
				checkMode:  CheckAll,
				issueCodes: nil,
			},
		},
		{
			name: "check all issues - python",
			args: args{
				name:          "file.py",
				content:       `print("Hello World")`,
				commentPrefix: []string{"#"},
			},
			want: want{
				checkMode: CheckAll,
			},
		},
		{
			name: "include issues - pragma",
			args: args{
				name:          "file.py",
				content:       `# scatr-check: PY-W1000, PY-S1024,PY-1234, issue-code`,
				commentPrefix: []string{"#"},
			},
			want: want{
				checkMode:  CheckInclude,
				issueCodes: []string{"PY-W1000", "PY-S1024", "PY-1234", "issue-code"},
			},
		},
		{
			name: "exclude issues",
			args: args{
				name:          "file.py",
				content:       `# scatr-ignore: PY-W1000, PY-S1024,PY-1234`,
				commentPrefix: []string{"#"},
			},
			want: want{
				checkMode:  CheckExclude,
				issueCodes: []string{"PY-W1000", "PY-S1024", "PY-1234"},
			},
		},
		{
			name: "invalid pragma",
			args: args{
				name: "file.py",
				content: `# pragma needs to be on the first line
# scatr-ignore: PY-W1000, PY-S1024,PY-1234`,
				commentPrefix: []string{"#"},
			},
			want: want{
				checkMode: CheckAll,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFile(tt.args.name, tt.args.content, tt.args.commentPrefix)

			if !reflect.DeepEqual(got.CheckMode, tt.want.checkMode) {
				t.Errorf("NewFile().CheckMode = %v, want %v", got.CheckMode, tt.want.checkMode)
			}
			if !reflect.DeepEqual(got.IssueCodes, tt.want.issueCodes) {
				t.Errorf("NewFile().IssueCodes = %v, want %v", got.IssueCodes, tt.want.issueCodes)
			}
		})
	}
}
