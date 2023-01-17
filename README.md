# SCATR

Static Code Analysis Testing Framework which just works!

SCATR is a simple framework to test static analyzers and Autofixers.

[![DeepSource](https://deepsource.io/gh/deepsourcelabs/SCATR.svg/?label=active+issues&show_trend=true&token=l9MSP_TWMIT2_-Lr5cW4YdBl)](https://deepsource.io/gh/deepsourcelabs/SCATR/?ref=repository-badge)
[![DeepSource](https://deepsource.io/gh/deepsourcelabs/SCATR.svg/?label=resolved+issues&show_trend=true&token=l9MSP_TWMIT2_-Lr5cW4YdBl)](https://deepsource.io/gh/deepsourcelabs/SCATR/?ref=repository-badge)

## Getting Started

`scatr` accepts a `.scatr.toml` as its configuration file.

```toml
files = "*.go"
comment_prefix = ["//"]
code_path = ""
excluded_dirs = []

[checks]
script = """
go run ./cmd/runner --output-file=./analysis_result.json
"""
interpreter = "sh"
output_file = "analysis_result.json"

[processor]
skip_processing = false
script = "process-results $INPUT_FILE"
interpreter = "sh"

[autofix]
script = """
go run ./cmd/autofix
"""
interpreter = "sh"
```

### `code_path`

SCATR optionally accepts a configuration item `code_path` which is absolute,
or a path relative to the `cwd`. When the `code_path` is specified, the runner
expects the file paths in the `analysis_results.json` (output from the
processor) to either be absolute, or to be relative to the `code_path`. Same
goes for the `files` flag. Only the files in the `code_path` are tested.

### `excluded_dirs`

SCATR optionally also accepts a list of directories (absolute or relative to the
`cwd`) from which results are excluded. Say for example, an issue was raised
in one of the excluded directory. SCATR will ignore matching any files inside
the `excluded_dirs`. Same applies for Autofix.

## Testing Checks

SCATR has two stages,

1. The `run` stage which runs the provided script using the provided interpreter
2. The `processor` stage which takes the `run` output and converts it into the
   format compatible with SCATR

### The `processor`

The processor is expected to convert the `run` `output_file` to a JSON-based
format `scatr` expects. It is expected to print the JSON to `stdout`. Just like
the runner, SCATR accepts an arbitrary script as a processor. The `output_file`
from the `run` stage is passed as the `INPUT_FILE` environment variable.

The format that `scatr` expects is as follows:

```json
{
  "issues": [
    {
      "code": "ISSUE-CODE",
      "title": "issue occurrence title",
      "position": {
        "file": "file.go",
        "start": {
          "line": 9,
          "column": 10
        }
      }
    }
  ]
}
```

The column numbers are optional.

### Expected result pragma

The runner uses pragmas in comments to get a set of issues which are
expected to be raised. The pragmas are of the following format:

```go
// [ISSUE-CODE]: col-num "title"
```

Here the `col-num` (column number) and the `title` is optional. Pragmas
can be on the same line, or the previous line. Here is an example:

```go
package main

func main() {
	a := 10
	// [VET-V0002]: "Useless assignment"
	a = a

	a = a // [VET-V0002]: "Useless assignment"
}
```

You can chain multiple occurrences of the same issue by using a `,`.
For example,

```go
// [VET-V0002]: 9 "Useless assignment (occurrence 1)", "occurrence 2"
```

You can also chain multiple issues in the same line using `;`. For
example,

```go
package main

// [VET-V0002]: "Useless assignment"; [SCC-U1000]: "func foo is unused"
func foo() {}
```

Pragma comments can optionally be split into multiple lines assuming that there
are no other comments between the lines. For example, the following pragmas have
the same meaning:

- ```javascript
  const foo = true;
  
  // [JS-W0126]: "Variables should not be initialized to undefined"; [JS-0345]
  const bar = foo === false ? undefined : "baz";
  ```

- ```javascript
  const foo = true;
  
  // [JS-W0126]: "Variables should not be initialized to undefined"
  // [JS-0345]
  const bar = foo === false ? undefined : "baz";
  ```

- ```javascript
  const foo = true;
  
  // [JS-W0126]: "Variables should not be initialized to undefined"
  const bar = foo === false ? undefined : "baz"; // [JS-0345]
  ```

The `comment_prefix` in the configuration file is used by the runner
to determine the comments. It accepts a list of prefixes to use for pragma
extraction. For example, it can be `//` for Go files, or `#` for Python files.

The `files` field is used by the runner to get a list of files to
extract the pragmas from.

## Testing Autofix

SCATR uses "golden files" to test for Autofix. It is similar to how testing
checks work, although there is no `processor` stage involved as this Autofix'ed
files are directly compared with "golden files", which basically are files that
contain the expected output after performing Autofix.

Golden files use the same name as the test file, with the `.golden` suffix
appended. For example, the golden file for `main.go` will be `main.go.golden`,
and for `main.py`, it will be `main.py.golden`.

SCATR uses the `files` glob pattern defined in the config along with the
`.gitignore` in the directory root to create a snapshot of the current state
in the `autofix-dir` (optionally provided as a flag). In case no `autofix-dir`
flag is provided, SCATR expects the script to modify the files **in-place** and
restores the snapshot after the results have been calculated. After the snapshot
has been created, it runs the provided Autofix `script`. After the script is
completed, SCATR calculates a `diff` from their `.golden` counterparts.

In case no `autofix-dir` has been provided, the snapshot and is only done
against the `files` glob pattern, so the Autofix tool should be sure to not
modify something else, or it might lead to incorrect results and the modified
files not being restored.

## Running

After creating a `.scatr.toml` file, you can simply run `scatr run`
for the test runner to run.

SCATR sets the `CODE_PATH` environment variable to the provided `cwd` to SCATR
(defaults to the OS current working directory) before running the `check` and
`autofix` script. This is always an absolute path.

### Flags

- `-c`, or `--cwd`: used to set the current working directory of the runner.
  All paths in other flags are relative to this. If not set, "." is used.
- `-p`, or `--pretty`: enables or disable pretty printing. It defaults to
  `false` for non-interactive environments.
- `-v` or `--verbose`: enables verbose logging on `stderr`
- `-f` or `--files`: an array of files to run the tests on. All other files
  in the glob pattern specified in `.scatr.toml` are ignored. This should
  be a subset of the glob pattern.
- `-a` or `--autofix-dir`: specify the directory for Autofix tests. This is
  where the Autofix tool is expected to produce its output. An absolute path
  to the Autofix directory is exposed to the run script in the `OUTPUT_PATH`
  environment variable. In the case where the Autofix directory is not
  specified, it uses the current working directory. In this case, SCATR
  creates a snapshot of the current working directory using the `files` glob
  pattern and the root `.gitignore` and restores it after the results have
  been calculated.

## Development

SCATR is built using [Go](https://go.dev). To hack on SCATR, you need a working
installation of Go.

### Directory Structure

- `cmd` - The entrypoint for the CLI application
- `pragma` - The pragma parser and the file reader
- `runner` - Actual test runner responsible for running the `checks`,
  `processor` and `autofix`, and for the result calculation
  - `runner/testdata` - Data for testing the runner's capabilities
    - `runner/testdata/checks` - Used for testing the `checks` result
      calculation
    - `runner/testdata/autofix` - Used for testing the `autofix` result
      calculation
    - `runner/testdata/backup` - Used for testing the backing up of Autofix'able
      files
    - `runner/testdata/backup_autofixdir` - Used for testing the backing up of
      Autofix'able files when the `--autofix-dir` flag is specified
    - `runner/testdata/config` - Used for testing the configuration handling and
      the configuration defaults

## License

SCATR is licensed under the [MIT license](./LICENSE).
