package issue

import (
	"fmt"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// SubIssueResult represents the result of an add or remove sub-issue operation.
type SubIssueResult struct {
	Parent int    `json:"parent"`
	Child  int    `json:"child"`
	Action string `json:"action"`
}

// SubIssueInfo represents a sub-issue in list output.
type SubIssueInfo struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
}

// AddSubIssue creates a sub-issue relationship between parent and child issues.
func AddSubIssue(client ghcli.Querier, owner, repo string, parentNumber, childNumber int) (*SubIssueResult, error) {
	parentID, err := GetIssueNodeID(client, owner, repo, parentNumber)
	if err != nil {
		return nil, fmt.Errorf("resolve parent issue: %w", err)
	}

	childID, err := GetIssueNodeID(client, owner, repo, childNumber)
	if err != nil {
		return nil, fmt.Errorf("resolve child issue: %w", err)
	}

	query := `mutation($issueId: ID!, $subIssueId: ID!) {
		addSubIssue(input: {issueId: $issueId, subIssueId: $subIssueId}) {
			issue {
				id
			}
		}
	}`

	var result struct {
		AddSubIssue struct {
			Issue struct {
				ID string `json:"id"`
			} `json:"issue"`
		} `json:"addSubIssue"`
	}

	if err := client.Query(query, map[string]any{
		"issueId":    parentID,
		"subIssueId": childID,
	}, &result); err != nil {
		return nil, fmt.Errorf("add sub-issue: %w", err)
	}

	return &SubIssueResult{
		Parent: parentNumber,
		Child:  childNumber,
		Action: "added",
	}, nil
}

// RemoveSubIssue removes a sub-issue relationship between parent and child issues.
func RemoveSubIssue(client ghcli.Querier, owner, repo string, parentNumber, childNumber int) (*SubIssueResult, error) {
	parentID, err := GetIssueNodeID(client, owner, repo, parentNumber)
	if err != nil {
		return nil, fmt.Errorf("resolve parent issue: %w", err)
	}

	childID, err := GetIssueNodeID(client, owner, repo, childNumber)
	if err != nil {
		return nil, fmt.Errorf("resolve child issue: %w", err)
	}

	query := `mutation($issueId: ID!, $subIssueId: ID!) {
		removeSubIssue(input: {issueId: $issueId, subIssueId: $subIssueId}) {
			issue {
				id
			}
		}
	}`

	var result struct {
		RemoveSubIssue struct {
			Issue struct {
				ID string `json:"id"`
			} `json:"issue"`
		} `json:"removeSubIssue"`
	}

	if err := client.Query(query, map[string]any{
		"issueId":    parentID,
		"subIssueId": childID,
	}, &result); err != nil {
		return nil, fmt.Errorf("remove sub-issue: %w", err)
	}

	return &SubIssueResult{
		Parent: parentNumber,
		Child:  childNumber,
		Action: "removed",
	}, nil
}

// ListSubIssuesResult represents the result of listing sub-issues.
type ListSubIssuesResult struct {
	SubIssues  []SubIssueInfo `json:"subIssues"`
	TotalCount int            `json:"totalCount"`
}

// ListSubIssues lists all sub-issues of the given issue.
func ListSubIssues(client ghcli.Querier, owner, repo string, issueNumber int) (*ListSubIssuesResult, error) {
	issueID, err := GetIssueNodeID(client, owner, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("resolve issue: %w", err)
	}

	query := `query($id: ID!) {
		node(id: $id) {
			... on Issue {
				subIssues(first: 100) {
					totalCount
					nodes {
						number
						title
						state
					}
				}
			}
		}
	}`

	var result struct {
		Node struct {
			SubIssues struct {
				TotalCount int            `json:"totalCount"`
				Nodes      []SubIssueInfo `json:"nodes"`
			} `json:"subIssues"`
		} `json:"node"`
	}

	if err := client.Query(query, map[string]any{
		"id": issueID,
	}, &result); err != nil {
		return nil, fmt.Errorf("list sub-issues: %w", err)
	}

	nodes := result.Node.SubIssues.Nodes
	if nodes == nil {
		nodes = []SubIssueInfo{}
	}

	return &ListSubIssuesResult{
		SubIssues:  nodes,
		TotalCount: result.Node.SubIssues.TotalCount,
	}, nil
}
