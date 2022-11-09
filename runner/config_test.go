package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/google/go-cmp/cmp"
)

// TestReadConfig tests the different aspects of the config reader
func TestReadConfig(t *testing.T) {
	tests := []string{
		"no_interpreter",
		"no_test_checks", "test_checks",
		"no_test_autofix", "test_autofix",
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testDir := filepath.Join(cwd, "testdata", "config")

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			dir := filepath.Join(testDir, test)

			got, err := ReadConfig(filepath.Join(dir, ".scatr.toml"))
			if err != nil {
				t.Fatal(err)
			}

			expected, err := readExpectedConfig(filepath.Join(dir, ".scatr.expected.toml"))
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(got, expected) {
				t.Fatalf("got and expected configs don't match, diff: %s",
					cmp.Diff(got, expected))
			}
		})
	}
}

// readExpectedConfig decodes the provided config path while not having the
// "smart defaults" unlike the non-test ReadConfig.
func readExpectedConfig(filePath string) (*Config, error) {
	var config Config
	_, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
