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

	reviewViewCmd := &cobra.Command{
		Use:   "view <pr-number>",
		Short: "View all reviews and threads for a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReviewView(cmd, args)
		},
	}
	reviewViewCmd.Flags().String("reviewer", "", "Filter by reviewer login")
	reviewViewCmd.Flags().String("state", "", "Filter by review state")
	reviewViewCmd.Flags().Bool("unresolved", false, "Only show unresolved threads")
	reviewViewCmd.Flags().Int("tail", 0, "Show only last N comments per thread")
	reviewViewCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	reviewEditCmd := &cobra.Command{
		Use:   "edit <review-id>",
		Short: "Edit a submitted review's body text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReviewEdit(cmd, args)
		},
	}
	reviewEditCmd.Flags().StringP("body", "b", "", "New review body")
	reviewEditCmd.Flags().String("body-file", "", "Read body from file")
	reviewEditCmd.MarkFlagsMutuallyExclusive("body", "body-file")
	reviewEditCmd.MarkFlagsOneRequired("body", "body-file")

	reviewPreviewCmd := &cobra.Command{
		Use:   "preview <review-id>",
		Short: "Preview a pending review's comments before submitting",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runReviewPreview(args)
		},
	}

	threadAddCmd := &cobra.Command{
		Use:   "add <review-id>",
		Short: "Add an inline comment thread to a pending review",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReviewThreadAdd(cmd, args)
		},
	}
	threadAddCmd.Flags().String("path", "", "File path to comment on")
	threadAddCmd.Flags().Int("line", 0, "End line number")
	threadAddCmd.Flags().StringP("body", "b", "", "Comment text")
	threadAddCmd.Flags().String("body-file", "", "Read body from file")
	threadAddCmd.Flags().String("side", "", "LEFT or RIGHT (default: RIGHT)")
	threadAddCmd.Flags().Int("start-line", 0, "Multi-line start line")
	threadAddCmd.Flags().String("start-side", "", "Multi-line start side (LEFT or RIGHT)")
	_ = threadAddCmd.MarkFlagRequired("path")
	_ = threadAddCmd.MarkFlagRequired("line")
	threadAddCmd.MarkFlagsMutuallyExclusive("body", "body-file")
	threadAddCmd.MarkFlagsOneRequired("body", "body-file")

	threadReplyCmd := &cobra.Command{
		Use:   "reply <thread-id>",
		Short: "Reply to an existing review thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReviewThreadReply(cmd, args)
		},
	}
	threadReplyCmd.Flags().StringP("body", "b", "", "Reply text")
	threadReplyCmd.Flags().String("body-file", "", "Read body from file")
	threadReplyCmd.MarkFlagsMutuallyExclusive("body", "body-file")
	threadReplyCmd.MarkFlagsOneRequired("body", "body-file")

	threadListCmd := &cobra.Command{
		Use:   "list <pr-number>",
		Short: "List review threads for a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReviewThreadList(cmd, args)
		},
	}
	threadListCmd.Flags().Bool("unresolved", false, "Only show unresolved threads")
	threadListCmd.Flags().Bool("mine", false, "Only show threads started by me")
	threadListCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	threadResolveCmd := &cobra.Command{
		Use:   "resolve <thread-id>",
		Short: "Resolve a review thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runReviewThreadResolve(args, true)
		},
	}

	threadUnresolveCmd := &cobra.Command{
		Use:   "unresolve <thread-id>",
		Short: "Unresolve a review thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runReviewThreadResolve(args, false)
		},
	}

	threadEditCommentCmd := &cobra.Command{
		Use:   "edit-comment <comment-id>",
		Short: "Edit a review comment's body",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReviewThreadEditComment(cmd, args)
		},
	}
	threadEditCommentCmd.Flags().StringP("body", "b", "", "New comment body")
	threadEditCommentCmd.Flags().String("body-file", "", "Read body from file")
	threadEditCommentCmd.MarkFlagsMutuallyExclusive("body", "body-file")
	threadEditCommentCmd.MarkFlagsOneRequired("body", "body-file")

	threadDeleteCommentCmd := &cobra.Command{
		Use:   "delete-comment <comment-id>",
		Short: "Delete a review comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runReviewThreadDeleteComment(args)
		},
	}

	threadCmd := &cobra.Command{
		Use:   "thread",
		Short: "Manage review comment threads",
	}
	threadCmd.AddCommand(threadAddCmd)
	threadCmd.AddCommand(threadReplyCmd)
	threadCmd.AddCommand(threadListCmd)
	threadCmd.AddCommand(threadResolveCmd)
	threadCmd.AddCommand(threadUnresolveCmd)
	threadCmd.AddCommand(threadEditCommentCmd)
	threadCmd.AddCommand(threadDeleteCommentCmd)

	reviewCmd := &cobra.Command{
		Use:   "review",
		Short: "PR review operations",
	}
	reviewCmd.AddCommand(reviewStartCmd)
	reviewCmd.AddCommand(reviewSubmitCmd)
	reviewCmd.AddCommand(reviewViewCmd)
	reviewCmd.AddCommand(reviewEditCmd)
	reviewCmd.AddCommand(reviewPreviewCmd)
	reviewCmd.AddCommand(threadCmd)
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

func runReviewView(cmd *cobra.Command, args []string) error {
	prNumber, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid PR number %q: %w", args[0], err)
	}

	repo, _ := cmd.Flags().GetString("repo")
	owner, repoName, err := issue.ResolveRepo(repo)
	if err != nil {
		return err
	}

	reviewer, _ := cmd.Flags().GetString("reviewer")
	state, _ := cmd.Flags().GetString("state")
	unresolved, _ := cmd.Flags().GetBool("unresolved")
	tail, _ := cmd.Flags().GetInt("tail")

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	result, err := review.ViewReviews(client, review.ViewOptions{
		Owner:      owner,
		Repo:       repoName,
		PRNumber:   prNumber,
		Reviewer:   reviewer,
		State:      state,
		Unresolved: unresolved,
		Tail:       tail,
	})
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func runReviewEdit(cmd *cobra.Command, args []string) error {
	body, err := resolveBody(cmd)
	if err != nil {
		return err
	}

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	result, err := review.EditReview(client, args[0], body)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func runReviewPreview(args []string) error {
	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	result, err := review.PreviewReview(client, args[0])
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func runReviewThreadAdd(cmd *cobra.Command, args []string) error {
	body, err := resolveBody(cmd)
	if err != nil {
		return err
	}

	path, _ := cmd.Flags().GetString("path")
	line, _ := cmd.Flags().GetInt("line")
	side, _ := cmd.Flags().GetString("side")
	startLine, _ := cmd.Flags().GetInt("start-line")
	startSide, _ := cmd.Flags().GetString("start-side")

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	result, err := review.AddThread(client, review.AddThreadOptions{
		ReviewID:  args[0],
		Path:      path,
		Line:      line,
		Body:      body,
		Side:      side,
		StartLine: startLine,
		StartSide: startSide,
	})
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func runReviewThreadReply(cmd *cobra.Command, args []string) error {
	body, err := resolveBody(cmd)
	if err != nil {
		return err
	}

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	result, err := review.ReplyToThread(client, review.ReplyThreadOptions{
		ThreadID: args[0],
		Body:     body,
	})
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func runReviewThreadList(cmd *cobra.Command, args []string) error {
	prNumber, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid PR number %q: %w", args[0], err)
	}

	repo, _ := cmd.Flags().GetString("repo")
	owner, repoName, err := issue.ResolveRepo(repo)
	if err != nil {
		return err
	}

	unresolved, _ := cmd.Flags().GetBool("unresolved")
	mine, _ := cmd.Flags().GetBool("mine")

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	threads, err := review.ListThreads(client, review.ListThreadsOptions{
		Owner:      owner,
		Repo:       repoName,
		PRNumber:   prNumber,
		Unresolved: unresolved,
		Mine:       mine,
	})
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(threads)
}

func runReviewThreadResolve(args []string, resolve bool) error {
	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	var result *review.ResolveThreadResult
	if resolve {
		result, err = review.ResolveThread(client, args[0])
	} else {
		result, err = review.UnresolveThread(client, args[0])
	}
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func runReviewThreadEditComment(cmd *cobra.Command, args []string) error {
	body, err := resolveBody(cmd)
	if err != nil {
		return err
	}

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	result, err := review.EditComment(client, args[0], body)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func runReviewThreadDeleteComment(args []string) error {
	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	result, err := review.DeleteComment(client, args[0])
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
