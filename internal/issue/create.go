package issue

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/repository"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// CreateOptions holds the options for creating an issue.
type CreateOptions struct {
	Title     string
	Body      string
	Labels    []string
	Assignees []string
	Milestone string
	Projects  []string
	Type      string
	Repo      string
}

// Create creates a new issue, optionally setting the issue type.
// It delegates basic creation to `gh issue create` and then sets the type via GraphQL if specified.
func Create(opts CreateOptions) error {
	args := []string{"issue", "create", "--title", opts.Title}

	if opts.Body != "" {
		args = append(args, "--body", opts.Body)
	}
	for _, l := range opts.Labels {
		args = append(args, "--label", l)
	}
	for _, a := range opts.Assignees {
		args = append(args, "--assignee", a)
	}
	if opts.Milestone != "" {
		args = append(args, "--milestone", opts.Milestone)
	}
	for _, p := range opts.Projects {
		args = append(args, "--project", p)
	}
	if opts.Repo != "" {
		args = append(args, "--repo", opts.Repo)
	}

	stdout, stderr, err := gh.Exec(args...)
	if err != nil {
		return fmt.Errorf("gh issue create: %w\nstderr: %s", err, stderr.String())
	}

	issueURL := strings.TrimSpace(stdout.String())
	fmt.Println(issueURL)

	if opts.Type != "" {
		if err := setTypeForNewIssue(opts, issueURL); err != nil {
			return fmt.Errorf("issue created but failed to set type: %w", err)
		}
		fmt.Printf("Issue type set to %q\n", opts.Type)
	}

	return nil
}

func setTypeForNewIssue(opts CreateOptions, issueURL string) error {
	owner, repo, err := resolveRepo(opts.Repo)
	if err != nil {
		return err
	}

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	types, err := FetchTypes(client, owner, repo)
	if err != nil {
		return fmt.Errorf("fetch issue types: %w", err)
	}

	issueType, err := FindTypeByName(types, opts.Type)
	if err != nil {
		return err
	}

	number, err := ExtractIssueNumber(issueURL)
	if err != nil {
		return err
	}

	nodeID, err := GetIssueNodeID(client, owner, repo, number)
	if err != nil {
		return fmt.Errorf("get issue node ID: %w", err)
	}

	return SetType(client, nodeID, issueType.ID)
}

// ExtractIssueNumber extracts the issue number from a GitHub issue URL.
func ExtractIssueNumber(issueURL string) (int, error) {
	issueURL = strings.TrimSpace(issueURL)
	idx := strings.LastIndex(issueURL, "/")
	if idx == -1 {
		return 0, fmt.Errorf("invalid issue URL: %s", issueURL)
	}
	number, err := strconv.Atoi(issueURL[idx+1:])
	if err != nil {
		return 0, fmt.Errorf("invalid issue number in URL %q: %w", issueURL, err)
	}
	return number, nil
}

// GetIssueNodeID fetches the node ID of an issue by its number via GraphQL.
func GetIssueNodeID(client *ghcli.Client, owner, repo string, number int) (string, error) {
	query := `query($owner: String!, $repo: String!, $number: Int!) {
		repository(owner: $owner, name: $repo) {
			issue(number: $number) {
				id
			}
		}
	}`

	var result struct {
		Repository struct {
			Issue struct {
				ID string `json:"id"`
			} `json:"issue"`
		} `json:"repository"`
	}

	err := client.Query(query, map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}, &result)
	if err != nil {
		return "", err
	}
	if result.Repository.Issue.ID == "" {
		return "", fmt.Errorf("issue #%d not found in %s/%s", number, owner, repo)
	}
	return result.Repository.Issue.ID, nil
}

func resolveRepo(repoFlag string) (string, string, error) {
	if repoFlag != "" {
		return parseRepoFlag(repoFlag)
	}
	repo, err := repository.Current()
	if err != nil {
		return "", "", fmt.Errorf("detect current repo: %w", err)
	}
	return repo.Owner, repo.Name, nil
}

func parseRepoFlag(flag string) (string, string, error) {
	parts := strings.SplitN(flag, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo format %q, expected owner/repo", flag)
	}
	return parts[0], parts[1], nil
}
