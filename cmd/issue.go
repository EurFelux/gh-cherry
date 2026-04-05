package cmd

import (
	"github.com/EurFelux/gh-cherry/internal/issue"
	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Enhanced issue management",
}

var issueCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an issue with optional type support",
	RunE:  runIssueCreate,
}

func init() {
	issueCreateCmd.Flags().StringP("title", "t", "", "Issue title (required)")
	issueCreateCmd.Flags().StringP("body", "b", "", "Issue body")
	issueCreateCmd.Flags().StringSliceP("label", "l", nil, "Labels to add")
	issueCreateCmd.Flags().StringSliceP("assignee", "a", nil, "Assignees")
	issueCreateCmd.Flags().StringP("milestone", "m", "", "Milestone")
	issueCreateCmd.Flags().StringSliceP("project", "p", nil, "Projects")
	issueCreateCmd.Flags().StringP("type", "T", "", "Issue type (e.g. Bug, Feature)")
	issueCreateCmd.Flags().StringP("repo", "R", "", "Repository in owner/repo format")

	_ = issueCreateCmd.MarkFlagRequired("title")

	issueCmd.AddCommand(issueCreateCmd)
	rootCmd.AddCommand(issueCmd)
}

func runIssueCreate(cmd *cobra.Command, _ []string) error {
	title, _ := cmd.Flags().GetString("title")
	body, _ := cmd.Flags().GetString("body")
	labels, _ := cmd.Flags().GetStringSlice("label")
	assignees, _ := cmd.Flags().GetStringSlice("assignee")
	milestone, _ := cmd.Flags().GetString("milestone")
	projects, _ := cmd.Flags().GetStringSlice("project")
	typeName, _ := cmd.Flags().GetString("type")
	repo, _ := cmd.Flags().GetString("repo")

	return issue.Create(issue.CreateOptions{
		Title:     title,
		Body:      body,
		Labels:    labels,
		Assignees: assignees,
		Milestone: milestone,
		Projects:  projects,
		Type:      typeName,
		Repo:      repo,
	})
}
