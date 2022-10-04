package runner

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func init() {
	// Do not show the runner logs during tests.
	log.SetOutput(io.Discard)
}

// TestTestChecks is an integration test for testing the checks
func TestTestChecks(t *testing.T) {
	type testResult struct {
		Passed bool       `json:"passed"`
		Result checksDiff `json:"result"`
	}

	tests := []string{
		"cpp", "cpp_failing",
		"go", "go_failing", "go_failing_misc",
		"go_multiple_pragmas", "go_failing_multiple_files", "go_included_files",
		"go_code_path", "go_code_path_included_files",
		"py", "py_failing",
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testDir := filepath.Join(cwd, "testdata", "checks")

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			dir := filepath.Join(testDir, test)
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}

			b, err := os.ReadFile("test_result.json")
			if err != nil {
				t.Fatal(err)
			}

			var expected testResult
			err = json.Unmarshal(b, &expected)
			if err != nil {
				t.Fatal(err)
			}

			config, err := ReadConfig(".scatr.toml")
			if err != nil {
				t.Fatal(err)
			}

			filesBytes, err := os.ReadFile("files.json")
			if err != nil {
				t.Fatal(err)
			}

			var files []string
			err = json.Unmarshal(filesBytes, &files)
			if err != nil {
				t.Fatal(err)
			}

			normalized, err := normalizeFileList(files, config.CodePath)
			if err != nil {
				t.Fatal(err)
			}

			got, passed, err := testChecks(config, normalized)
			if err != nil {
				t.Fatal(err)
			}

			if passed != expected.Passed {
				t.Fatalf("expected passed: %v, got %v", expected.Passed, passed)
			}

			for file, issues := range got {
				filePath, err := filepath.Rel(dir, file)
				if err != nil {
					t.Fatal(err)
				}

				if len(issues.Unexpected) == 0 &&
					len(issues.NotRaised) == 0 {
					continue
				}

				opts := []cmp.Option{
					cmpopts.IgnoreFields(IssuePosition{}, "fileNormalized"),
					cmpopts.IgnoreFields(IssuePosition{}, "File"),
					cmpopts.SortSlices(func(a, b any) bool {
						if a, ok := a.(*Issue); ok {
							b := b.(*Issue)
							return a.Code < b.Code
						}

						return false
					}),
				}

				if !cmp.Equal(issues, expected.Result[filePath], opts...) {
					t.Fatalf("unexpected result for file %s, diff: %s",
						filePath,
						cmp.Diff(expected.Result[filePath], issues, opts...))
				}
			}
		})
	}

	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}
}

// TestTestAutofix is an integration test for testing Autofix
func TestTestAutofix(t *testing.T) {
	tests := []string{
		"go", "go_included_files", "go_failing",
		"go_failing_code_path", "go_no_golden_file",
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testDir := filepath.Join(cwd, "testdata", "autofix")

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			dir := filepath.Join(testDir, test)
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}

			filesFailingBytes, err := os.ReadFile("files_failing.json")
			if err != nil {
				t.Fatal(err)
			}

			var expectedFilesFailing []string
			err = json.Unmarshal(filesFailingBytes, &expectedFilesFailing)
			if err != nil {
				t.Fatal(err)
			}

			filesBytes, err := os.ReadFile("files.json")
			if err != nil {
				t.Fatal(err)
			}

			var files []string
			err = json.Unmarshal(filesBytes, &files)
			if err != nil {
				t.Fatal(err)
			}

			config, err := ReadConfig(".scatr.toml")
			if err != nil {
				t.Fatal(err)
			}

			normalized, err := normalizeFileList(files, config.CodePath)
			if err != nil {
				t.Fatal(err)
			}

			got, passed, err := testAutofix(config, normalized, "")
			if err != nil {
				t.Fatal(err)
			}

			expectedPassed := len(expectedFilesFailing) == 0
			if passed != expectedPassed {
				t.Fatalf("expected passed: %v, got: %v", expectedPassed, passed)
			}

			filesFailing := make([]string, 0, len(got))
			for fileFailing := range got {
				filesFailing = append(filesFailing, fileFailing)
			}

			opts := []cmp.Option{
				cmpopts.SortSlices(func(a, b any) bool {
					if a, ok := a.(string); ok {
						return a < b.(string)
					}
					return false
				}),
			}

			if !cmp.Equal(filesFailing, expectedFilesFailing, opts...) {
				t.Fatalf("unexpected files failing, diff: %s",
					cmp.Diff(filesFailing, expectedPassed, opts...))
			}
		})
	}

	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}
}

// TestTestAutofix_AutofixDir is an integration test for testing Autofix with
// AutofixDir set
func TestTestAutofix_AutofixDir(t *testing.T) {
	tests := []string{
		"go", "go_included_files", "go_failing",
		"go_failing_code_path", "go_no_golden_file",
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testDir := filepath.Join(cwd, "testdata", "autofix")

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			dir := filepath.Join(testDir, test)
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}

			autofixDir, err := os.MkdirTemp("", "scatr-test")
			if err != nil {
				t.Fatal(err)
			}

			filesFailingBytes, err := os.ReadFile("files_failing.json")
			if err != nil {
				t.Fatal(err)
			}

			var expectedFilesFailing []string
			err = json.Unmarshal(filesFailingBytes, &expectedFilesFailing)
			if err != nil {
				t.Fatal(err)
			}

			filesBytes, err := os.ReadFile("files.json")
			if err != nil {
				t.Fatal(err)
			}

			var files []string
			err = json.Unmarshal(filesBytes, &files)
			if err != nil {
				t.Fatal(err)
			}

			config, err := ReadConfig(".scatr.toml")
			if err != nil {
				t.Fatal(err)
			}

			normalized, err := normalizeFileList(files, config.CodePath)
			if err != nil {
				t.Fatal(err)
			}

			got, passed, err := testAutofix(config, normalized, autofixDir)
			if err != nil {
				t.Fatal(err)
			}

			expectedPassed := len(expectedFilesFailing) == 0
			if passed != expectedPassed {
				t.Fatalf("expected passed: %v, got: %v", expectedPassed, passed)
			}

			filesFailing := make([]string, 0, len(got))
			for fileFailing := range got {
				filesFailing = append(filesFailing, fileFailing)
			}

			opts := []cmp.Option{
				cmpopts.SortSlices(func(a, b any) bool {
					if a, ok := a.(string); ok {
						return a < b.(string)
					}
					return false
				}),
			}

			if !cmp.Equal(filesFailing, expectedFilesFailing, opts...) {
				t.Fatalf("unexpected files failing, diff: %s",
					cmp.Diff(filesFailing, expectedPassed, opts...))
			}

			err = os.RemoveAll(autofixDir)
			if err != nil {
				t.Fatal(err)
			}
		})
	}

	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}
}
