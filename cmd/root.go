package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitgraphy",
	Short: "A simple CLI tool for visualizing Git graphs",
	Long:  "GitGraphy is a CLI tool for visualizing Git repositories",
}

func Execute() error {
	return rootCmd.Execute()
}