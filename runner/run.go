package runner

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func Run(printer IssuePrinter, files []string, autofixDir string) (bool, error) {
	config, err := ReadConfig(".scatr.toml")
	if err != nil {
		return false, err
	}

	if !config.TestAutofix && !config.TestChecks {
		return false, errors.New("nothing to do")
	}

	includedFiles, err := normalizeFileList(files, config.CodePath)
	if err != nil {
		return false, err
	}

	passed := true

	if config.TestChecks {
		printer.PrintHeader("Testing checks")
		res, testPassed, err := testChecks(config, includedFiles)
		if err != nil {
			return false, err
		}

		if !testPassed {
			printChecksDiff(res, printer)
			passed = false
		}
	}

	if config.TestAutofix {
		printer.PrintHeader("Testing Autofix")
		res, testPassed, err := testAutofix(config, includedFiles, autofixDir)
		if err != nil {
			return false, err
		}

		if !testPassed {
			printAutofixDiff(res, printer)
			passed = false
		}
	}

	printer.PrintStatus(passed)
	return passed, nil
}

func testChecks(config *Config, includedFiles map[string]bool) (checksDiff, bool, error) {
	log.Printf("Running the checks test script with the interpreter %q\n", config.Checks.Interpreter)
	log.Println("--- Checks run log ---")

	startTime := time.Now()
	err := runScript(config.Checks, config.CodePath, map[string]string{})
	if err != nil {
		return nil, false, err
	}

	log.Println("Checks test script completed in", time.Since(startTime))

	result, err := runProcessor(config.Processor, config.Checks.OutputFile, config.CodePath)
	if err != nil {
		return nil, false, err
	}

	files, err := readFiles(config, includedFiles)
	if err != nil {
		return nil, false, err
	}

	res, passed := diffChecksResult(files, includedFiles, result)
	return res, passed, err
}

func testAutofix(
	config *Config,
	includedFiles map[string]bool,
	autofixDir string,
) (autofixDiff, bool, error) {
	log.Println("Backing up the potentially Autofix'able files")
	backup, err := NewAutofixBackup(config, includedFiles, autofixDir)
	if err != nil {
		return nil, false, err
	}

	log.Printf("Running the Autofix test script with the interpreter %q\n", config.Checks.Interpreter)
	log.Println("--- Autofix run log ---")

	startTime := time.Now()

	var outputDir string
	if autofixDir == "" {
		outputDir, err = os.Getwd()
		if err != nil {
			return nil, false, err
		}
	} else {
		outputDir, err = filepath.Abs(autofixDir)
		if err != nil {
			return nil, false, err
		}
	}

	err = runScript(
		config.Autofix,
		config.CodePath,
		map[string]string{"OUTPUT_DIR": outputDir},
	)
	if err != nil {
		return nil, false, err
	}

	err = os.Unsetenv("OUTPUT_DIR")
	if err != nil {
		return nil, false, err
	}

	log.Println("Autofix test script completed in", time.Since(startTime))

	return diffAutofixResult(config.CodePath, backup)
}

// runScript runs a test runner script with the provided interpreter and pipes
// the command's stdout and stderr on the host's stderr
func runScript(cfg TestRunnerConfig, codePath string, env map[string]string) error {
	if env == nil {
		env = make(map[string]string)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	env["CODE_PATH"] = filepath.Join(cwd, codePath)
	err = setEnv(env)
	if err != nil {
		return err
	}

	scriptFile, err := os.CreateTemp("", "scatr-script")
	if err != nil {
		return err
	}

	scriptFilePath := scriptFile.Name()
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Println("Cleanup error", err)
		}
	}(scriptFilePath)

	_, err = scriptFile.WriteString(cfg.Script)
	if err != nil {
		return err
	}

	err = scriptFile.Sync()
	if err != nil {
		return err
	}

	err = scriptFile.Close()
	if err != nil {
		return err
	}

	cmd := exec.Command(cfg.Interpreter, append(cfg.Args, scriptFilePath)...)

	if cfg.Interactive {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Stdin = nil
	}

	err = cmd.Run()
	if err != nil {
		return err
	}

	return unsetEnv(env)
}

func normalizeFileList(files []string, codePath string) (map[string]bool, error) {
	m := make(map[string]bool)
	for _, f := range files {
		normalized, err := normalizeFilePath(filepath.Join(codePath, f))
		if err != nil {
			log.Println("Error normalizing the file path", f, "err:", err)
			continue
		}

		m[normalized] = true
	}

	return m, nil
}
