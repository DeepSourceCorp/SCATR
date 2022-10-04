package runner

import "github.com/BurntSushi/toml"

type Config struct {
	FilesGlob     string           `toml:"files"`
	CommentPrefix []string         `toml:"comment_prefix"`
	CodePath      string           `toml:"code_path"`
	Checks        TestRunnerConfig `toml:"checks"`
	Autofix       TestRunnerConfig `toml:"autofix"`
	Processor     ProcessorConfig  `toml:"processor"`
	TestChecks    bool             `toml:"test_checks"`
	TestAutofix   bool             `toml:"test_autofix"`
}

type TestRunnerConfig struct {
	Interpreter string   `toml:"interpreter"`
	Script      string   `toml:"script"`
	OutputFile  string   `toml:"output_file"`
	Interactive bool     `toml:"interactive"`
	Args        []string `toml:"args"`
}

type ProcessorConfig struct {
	Interpreter    string `toml:"interpreter"`
	Script         string `toml:"script"`
	SkipProcessing bool   `toml:"skip_processing"`
}

func ReadConfig(filePath string) (*Config, error) {
	var config Config
	meta, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		return nil, err
	}

	if config.Checks.Interpreter == "" {
		// Use the `sh` interpreter by default.
		config.Checks.Interpreter = "sh"
	}

	if config.Autofix.Interpreter == "" {
		// Use the `sh` interpreter by default.
		config.Autofix.Interpreter = "sh"
	}

	if config.Processor.Interpreter == "" {
		// Use the `sh` interpreter by default.
		config.Processor.Interpreter = "sh"
	}

	if !meta.IsDefined("test_checks") {
		// Only test the checks if `test_checks` in the TOML is true, or if the key
		// `test_checks` is not present while the key `checks` is present.
		config.TestChecks = meta.IsDefined("checks")
	}

	if !meta.IsDefined("test_autofix") {
		// Only test the checks if `test_autofix` in the TOML is true, or if the key
		// `test_autofix` is not present while the key `autofix` is present.
		config.TestAutofix = meta.IsDefined("autofix")
	}

	return &config, nil
}
