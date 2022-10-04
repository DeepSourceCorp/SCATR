package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "scatr",
	Short: "Static Code Analysis Testing Framework, that just works!",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return err
		}

		// skipcq: RVV-A0003
		os.Exit(1)
		return nil
	},
}
