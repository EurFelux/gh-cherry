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
	issueCreateCmd.Flags().IntP("parent", "P", 0, "Parent issue number to attach as sub-issue")
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

	subissueAddCmd := &cobra.Command{
		Use:   "add <parent-number> <child-number>",
		Short: "Add a sub-issue to a parent issue",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubIssueAdd(cmd, args)
		},
	}
	subissueAddCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	subissueRemoveCmd := &cobra.Command{
		Use:   "remove <parent-number> <child-number>",
		Short: "Remove a sub-issue from a parent issue",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubIssueRemove(cmd, args)
		},
	}
	subissueRemoveCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	subissueListCmd := &cobra.Command{
		Use:   "list <issue-number>",
		Short: "List sub-issues of an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubIssueList(cmd, args)
		},
	}
	subissueListCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	subissueCmd := &cobra.Command{
		Use:   "subissue",
		Short: "Manage sub-issues",
	}
	subissueCmd.AddCommand(subissueAddCmd)
	subissueCmd.AddCommand(subissueRemoveCmd)
	subissueCmd.AddCommand(subissueListCmd)

	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Enhanced issue management",
	}
	issueCmd.AddCommand(issueCreateCmd)
	issueCmd.AddCommand(issueTypesCmd)
	issueCmd.AddCommand(subissueCmd)
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
	parent, _ := cmd.Flags().GetInt("parent")

	var client ghcli.Querier
	if typeName != "" || parent > 0 {
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
		Parent:    parent,
		Repo:      getString("repo"),
		Client:    client,
	})
}

func parseIssueNumber(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid issue number %q: %w", s, err)
	}
	return n, nil
}

func runSubIssueAdd(cmd *cobra.Command, args []string) error {
	parent, err := parseIssueNumber(args[0])
	if err != nil {
		return err
	}
	child, err := parseIssueNumber(args[1])
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

	result, err := issue.AddSubIssue(client, owner, repoName, parent, child)
	if err != nil {
		return err
	}

	return encodeJSON(cmd, result)
}

func runSubIssueRemove(cmd *cobra.Command, args []string) error {
	parent, err := parseIssueNumber(args[0])
	if err != nil {
		return err
	}
	child, err := parseIssueNumber(args[1])
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

	result, err := issue.RemoveSubIssue(client, owner, repoName, parent, child)
	if err != nil {
		return err
	}

	return encodeJSON(cmd, result)
}

func runSubIssueList(cmd *cobra.Command, args []string) error {
	issueNumber, err := parseIssueNumber(args[0])
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

	result, err := issue.ListSubIssues(client, owner, repoName, issueNumber)
	if err != nil {
		return err
	}

	return encodeJSON(cmd, result)
}

func runIssueTypes(cmd *cobra.Command) error {
	repo, _ := cmd.Flags().GetString("repo")

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	return issue.ListTypes(client, repo, os.Stdout)
}
