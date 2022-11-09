package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/deepsourcelabs/SCATR/runner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	runCwd     string
	pretty     bool
	verbose    bool
	files      []string
	autofixDir string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the tests in a provided directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := os.Chdir(runCwd)
		if err != nil {
			return err
		}

		var printer runner.IssuePrinter
		if pretty {
			printer = runner.NewPrettyIssuePrinter()
		} else {
			printer = &runner.DefaultIssuePrinter{}
		}

		if !verbose {
			log.SetOutput(io.Discard)
		}

		passed, err := runner.Run(printer, files, autofixDir)
		if err != nil {
			fmt.Println(err)
			// skipcq: RVV-A0003
			os.Exit(1)
		}

		if !passed {
			// skipcq: RVV-A0003
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringVarP(
		&runCwd, "cwd", "c", ".",
		"Set the current working directory of the runner.",
	)
	runCmd.Flags().BoolVarP(
		&pretty, "pretty", "p", term.IsTerminal(int(os.Stdout.Fd())),
		"Pretty print the results",
	)
	runCmd.Flags().BoolVarP(
		&verbose, "verbose", "v", false,
		"Use verbose logging",
	)
	runCmd.Flags().StringArrayVarP(
		&files, "files", "f", []string{},
		"Set the list of files to run the tests on. This is relative to the specified cwd.",
	)
	runCmd.Flags().StringVarP(
		&autofixDir, "autofix-dir", "a", "",
		"Sets the directory where Autofix testing takes place. The Autofix runner is run inside this "+
			"directory. It uses the cwd as the autofix-dir if nothing is specified."+
			"This accepts an absolute path or a path relative to the set cwd. SCATR sets the absolute "+
			"path of the user directory as the OUTPUT_PATH environment variable for the Autofix script. "+
			"It does not clean up the directory in case autofix-dir is set.",
	)

	rootCmd.AddCommand(runCmd)
}
