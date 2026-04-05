package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
	"github.com/EurFelux/gh-cherry/internal/issue"
	"github.com/spf13/cobra"
)

func init() {
	issueCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an issue with optional type support",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runIssueCreate(cmd)
		},
	}

	issueCreateCmd.Flags().StringP("title", "t", "", "Issue title (required)")
	issueCreateCmd.Flags().StringP("body", "b", "", "Issue body")
	issueCreateCmd.Flags().StringSliceP("label", "l", nil, "Labels to add")
	issueCreateCmd.Flags().StringSliceP("assignee", "a", nil, "Assignees")
	issueCreateCmd.Flags().StringP("milestone", "m", "", "Milestone")
	issueCreateCmd.Flags().StringSliceP("project", "p", nil, "Projects")
	issueCreateCmd.Flags().StringP("type", "T", "", "Issue type (e.g. Bug, Feature)")
	issueCreateCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	_ = issueCreateCmd.MarkFlagRequired("title")

	issueTypesCmd := &cobra.Command{
		Use:   "types",
		Short: "List available issue types for a repository",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runIssueTypes(cmd)
		},
	}
	issueTypesCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	subIssueAddCmd := &cobra.Command{
		Use:   "add <parent-number> <child-number>",
		Short: "Add a sub-issue relationship",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubIssueAddRemove(cmd, args, "add")
		},
	}
	subIssueAddCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	subIssueRemoveCmd := &cobra.Command{
		Use:   "remove <parent-number> <child-number>",
		Short: "Remove a sub-issue relationship",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubIssueAddRemove(cmd, args, "remove")
		},
	}
	subIssueRemoveCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	subIssueListCmd := &cobra.Command{
		Use:   "list <issue-number>",
		Short: "List sub-issues of an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubIssueList(cmd, args)
		},
	}
	subIssueListCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	subIssueCmd := &cobra.Command{
		Use:   "sub-issue",
		Short: "Manage sub-issue relationships",
	}
	subIssueCmd.AddCommand(subIssueAddCmd)
	subIssueCmd.AddCommand(subIssueRemoveCmd)
	subIssueCmd.AddCommand(subIssueListCmd)

	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Enhanced issue management",
	}
	issueCmd.AddCommand(issueCreateCmd)
	issueCmd.AddCommand(issueTypesCmd)
	issueCmd.AddCommand(subIssueCmd)
	rootCmd.AddCommand(issueCmd)
}

func runIssueCreate(cmd *cobra.Command) error {
	getString := func(name string) string {
		v, _ := cmd.Flags().GetString(name)
		return v
	}
	getStringSlice := func(name string) []string {
		v, _ := cmd.Flags().GetStringSlice(name)
		return v
	}

	typeName := getString("type")

	var client ghcli.Querier
	if typeName != "" {
		c, err := ghcli.NewClient()
		if err != nil {
			return err
		}
		client = c
	}

	return issue.Create(issue.CreateOptions{
		Title:     getString("title"),
		Body:      getString("body"),
		Labels:    getStringSlice("label"),
		Assignees: getStringSlice("assignee"),
		Milestone: getString("milestone"),
		Projects:  getStringSlice("project"),
		Type:      typeName,
		Repo:      getString("repo"),
		Client:    client,
	})
}

func runSubIssueAddRemove(cmd *cobra.Command, args []string, action string) error {
	parentNumber, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid parent issue number %q: %w", args[0], err)
	}

	childNumber, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid child issue number %q: %w", args[1], err)
	}

	repoFlag, _ := cmd.Flags().GetString("repo")

	owner, repo, err := issue.ResolveRepo(repoFlag)
	if err != nil {
		return err
	}

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	if action == "add" {
		return issue.AddSubIssue(client, owner, repo, parentNumber, childNumber, os.Stdout)
	}
	return issue.RemoveSubIssue(client, owner, repo, parentNumber, childNumber, os.Stdout)
}

func runSubIssueList(cmd *cobra.Command, args []string) error {
	issueNumber, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid issue number %q: %w", args[0], err)
	}

	repoFlag, _ := cmd.Flags().GetString("repo")

	owner, repo, err := issue.ResolveRepo(repoFlag)
	if err != nil {
		return err
	}

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	return issue.ListSubIssues(client, owner, repo, issueNumber, os.Stdout)
}

func runIssueTypes(cmd *cobra.Command) error {
	repo, _ := cmd.Flags().GetString("repo")

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	return issue.ListTypes(client, repo, os.Stdout)
}
