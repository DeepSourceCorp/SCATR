package runner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestAutofixBackup tests the pattern matching support and the paths of files
// which are backed up.
func TestAutofixBackup(t *testing.T) {
	tests := []string{
		"no_gitignore", "gitignore",
		"no_gitignore_included_files", "gitignore_included_files",
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testDir := filepath.Join(cwd, "testdata", "backup")

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			err := os.Chdir(filepath.Join(testDir, test))
			if err != nil {
				t.Fatal(err)
			}

			config, err := ReadConfig(".scatr.toml")
			if err != nil {
				t.Fatal(err)
			}

			var expectedBackedUpFiles []string
			backupFile, err := os.ReadFile("backup.json")
			if err != nil {
				t.Fatal(err)
			}

			err = json.Unmarshal(backupFile, &expectedBackedUpFiles)
			if err != nil {
				t.Fatal(err)
			}

			var includedFiles []string
			includedFilesFile, err := os.ReadFile("files.json")
			if err != nil {
				t.Fatal(err)
			}

			err = json.Unmarshal(includedFilesFile, &includedFiles)
			if err != nil {
				t.Fatal(err)
			}

			normalized, err := normalizeFileList(includedFiles, config.CodePath)
			if err != nil {
				t.Fatal(err)
			}

			backup, err := NewAutofixBackup(config, normalized, "")
			if err != nil {
				t.Fatal(err)
			}

			opts := []cmp.Option{
				cmpopts.SortSlices(func(a, b any) bool {
					if a, ok := a.(string); ok {
						return a < b.(string)
					}
					return false
				}),
			}

			if !cmp.Equal(backup.CopiedFiles, expectedBackedUpFiles, opts...) {
				t.Fatalf("unexpected files backed up, diff: %s",
					cmp.Diff(backup.CopiedFiles, expectedBackedUpFiles, opts...))
			}

			err = backup.RestoreAndDestroy()
			if err != nil {
				t.Fatal(err)
			}

			err = os.Chdir(cwd)
			if err != nil {
				t.Fatal(err)
			}
		})
	}

	err = os.Chdir(cwd)
	if err != nil {
		t.Fatal(err)
	}
}

// TestAutofixBackup tests the pattern matching support and the paths of files
// which are backed up with AutofixDir.
func TestAutofixBackup_AutofixDir(t *testing.T) {
	tests := []string{"no_gitignore", "gitignore"}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testDir := filepath.Join(cwd, "testdata", "backup_autofixdir")

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			autofixDir, err := os.MkdirTemp("", "scatr-test")
			if err != nil {
				t.Fatal(err)
			}

			err = os.Chdir(filepath.Join(testDir, test))
			if err != nil {
				t.Fatal(err)
			}

			config, err := ReadConfig(".scatr.toml")
			if err != nil {
				t.Fatal(err)
			}

			var expectedBackedUpFiles []string
			backupFile, err := os.ReadFile("backup.json")
			if err != nil {
				t.Fatal(err)
			}

			err = json.Unmarshal(backupFile, &expectedBackedUpFiles)
			if err != nil {
				t.Fatal(err)
			}

			var includedFiles []string
			includedFilesFile, err := os.ReadFile("files.json")
			if err != nil {
				t.Fatal(err)
			}

			err = json.Unmarshal(includedFilesFile, &includedFiles)
			if err != nil {
				t.Fatal(err)
			}

			normalized, err := normalizeFileList(includedFiles, config.CodePath)
			if err != nil {
				t.Fatal(err)
			}

			backup, err := NewAutofixBackup(config, normalized, autofixDir)
			if err != nil {
				t.Fatal(err)
			}

			dir, err := os.ReadDir(autofixDir)
			if err != nil {
				t.Fatal(err)
			}

			dirFiles := []string{}
			for _, f := range dir {
				dirFiles = append(dirFiles, f.Name())
			}

			opts := []cmp.Option{
				cmpopts.SortSlices(func(a, b any) bool {
					if a, ok := a.(string); ok {
						return a < b.(string)
					}
					return false
				}),
			}

			if !cmp.Equal(backup.CopiedFiles, expectedBackedUpFiles, opts...) {
				t.Fatalf("unexpected files backed up, diff: %s",
					cmp.Diff(backup.CopiedFiles, expectedBackedUpFiles, opts...))
			}

			if !cmp.Equal(backup.CopiedFiles, dirFiles, opts...) {
				t.Fatalf("unexpected files backed up (dirFiles), diff: %s",
					cmp.Diff(backup.CopiedFiles, dirFiles, opts...))
			}

			err = os.RemoveAll(autofixDir)
			if err != nil {
				t.Fatal(err)
			}

			err = backup.RestoreAndDestroy()
			if err != nil {
				t.Fatal(err)
			}

			err = os.Chdir(cwd)
			if err != nil {
				t.Fatal(err)
			}
		})
	}

	err = os.Chdir(cwd)
	if err != nil {
		t.Fatal(err)
	}
}
