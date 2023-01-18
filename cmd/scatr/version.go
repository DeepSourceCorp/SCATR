package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const version = "0.3.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Returns the current SCATR version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("SCATR analyzer testing framework")
		fmt.Println("  Version:", version)
		fmt.Println()
		fmt.Println("Environment:")
		fmt.Println("  OS:  ", runtime.GOOS)
		fmt.Println("  ARCH:", runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
