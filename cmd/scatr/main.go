package main

import (
	"log"
	"os"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[test] ")
	log.SetFlags(log.LstdFlags)
}

func main() {
	rootCmd.Execute()
}
