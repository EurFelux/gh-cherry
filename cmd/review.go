package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
	"github.com/EurFelux/gh-cherry/internal/issue"
	"github.com/EurFelux/gh-cherry/internal/review"
	"github.com/spf13/cobra"
)

func init() {
	reviewStartCmd := &cobra.Command{
		Use:   "start <pr-number>",
		Short: "Create a pending review on a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReviewStart(cmd, args)
		},
	}
	reviewStartCmd.Flags().StringP("body", "b", "", "Review body text")
	reviewStartCmd.Flags().String("body-file", "", "Read body from file")
	reviewStartCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")
	reviewStartCmd.MarkFlagsMutuallyExclusive("body", "body-file")

	reviewSubmitCmd := &cobra.Command{
		Use:   "submit <review-id>",
		Short: "Submit a pending review",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReviewSubmit(cmd, args)
		},
	}
	reviewSubmitCmd.Flags().StringP("event", "e", "", "Review event: APPROVE, REQUEST_CHANGES, or COMMENT")
	reviewSubmitCmd.Flags().StringP("body", "b", "", "Summary text")
	reviewSubmitCmd.Flags().String("body-file", "", "Read summary from file")
	_ = reviewSubmitCmd.MarkFlagRequired("event")
	reviewSubmitCmd.MarkFlagsMutuallyExclusive("body", "body-file")

	reviewCmd := &cobra.Command{
		Use:   "review",
		Short: "PR review operations",
	}
	reviewCmd.AddCommand(reviewStartCmd)
	reviewCmd.AddCommand(reviewSubmitCmd)
	rootCmd.AddCommand(reviewCmd)
}

func runReviewStart(cmd *cobra.Command, args []string) error {
	prNumber, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid PR number %q: %w", args[0], err)
	}

	body, err := resolveBody(cmd)
	if err != nil {
		return err
	}

	repo, _ := cmd.Flags().GetString("repo")
	owner, repoName, err := issue.ResolveRepo(repo)
	if err != nil {
		return err
	}

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	result, err := review.StartReview(client, owner, repoName, prNumber, body)
	if err != nil {
		return err
	}

	if result.Reused && body != "" {
		fmt.Fprintln(os.Stderr, "warning: --body ignored, reusing existing pending review")
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func runReviewSubmit(cmd *cobra.Command, args []string) error {
	event, _ := cmd.Flags().GetString("event")

	body, err := resolveBody(cmd)
	if err != nil {
		return err
	}

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	result, err := review.SubmitReview(client, args[0], event, body)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

// resolveBody reads the review body from --body or --body-file.
func resolveBody(cmd *cobra.Command) (string, error) {
	body, _ := cmd.Flags().GetString("body")
	if body != "" {
		return body, nil
	}

	bodyFile, _ := cmd.Flags().GetString("body-file")
	if bodyFile == "" {
		return "", nil
	}

	data, err := os.ReadFile(bodyFile)
	if err != nil {
		return "", fmt.Errorf("read body file: %w", err)
	}
	return string(data), nil
}
