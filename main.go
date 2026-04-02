package main

import (
	"os"

	"github.com/replay/replay/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		println("CRITICAL ERROR: " + err.Error())
		os.Exit(1)
	}
}
