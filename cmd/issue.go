package cmd

import (
	"os"

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

	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Enhanced issue management",
	}
	issueCmd.AddCommand(issueCreateCmd)
	issueCmd.AddCommand(issueTypesCmd)
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

func runIssueTypes(cmd *cobra.Command) error {
	repo, _ := cmd.Flags().GetString("repo")

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	return issue.ListTypes(client, repo, os.Stdout)
}
