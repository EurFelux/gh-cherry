package main

import (
	"os"

	"github.com/EurFelux/gh-cherry/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
