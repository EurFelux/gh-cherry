package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
	"github.com/EurFelux/gh-cherry/internal/issue"
	"github.com/EurFelux/gh-cherry/internal/prdiff"
	"github.com/spf13/cobra"
)

func init() {
	prDiffCmd := &cobra.Command{
		Use:   "diff <number>",
		Short: "Show annotated diff for a pull request with L/R line numbers",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPRDiff(cmd, args)
		},
	}
	prDiffCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	prCmd := &cobra.Command{
		Use:   "pr",
		Short: "Enhanced pull request operations",
	}
	prCmd.AddCommand(prDiffCmd)
	rootCmd.AddCommand(prCmd)
}

func runPRDiff(cmd *cobra.Command, args []string) error {
	prNumber, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid PR number %q: %w", args[0], err)
	}

	repo, _ := cmd.Flags().GetString("repo")
	owner, repoName, err := issue.ResolveRepo(repo)
	if err != nil {
		return err
	}

	client, err := ghcli.NewRESTClient()
	if err != nil {
		return err
	}

	return prdiff.AnnotatedDiff(client, owner, repoName, prNumber, os.Stdout)
}
