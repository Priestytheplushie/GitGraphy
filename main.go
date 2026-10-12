package main

import (
	"os"

	"github.com/Priestytheplushie/GitGraphy/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}