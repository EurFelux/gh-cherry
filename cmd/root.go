package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "cherry",
	Short: "Cherry - Enhanced GitHub CLI extension",
	Long:  "A GitHub CLI extension that provides enhanced issue management and PR review capabilities.",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
