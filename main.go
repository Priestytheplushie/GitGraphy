package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gitgraphy",
	Short: "A simple CLI tool for visualizing Git graphs",
	Long:  "GitGraphy is a CLI tool for visualizing Git repositories",
}

