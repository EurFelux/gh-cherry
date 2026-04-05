package issue

import (
	"fmt"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// Type represents a GitHub issue type.
type Type struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FetchTypes fetches available issue types for a repository.
func FetchTypes(client ghcli.Querier, owner, repo string) ([]Type, error) {
	query := `query($owner: String!, $repo: String!) {
		repository(owner: $owner, name: $repo) {
			issueTypes(first: 50) {
				nodes {
					id
					name
				}
			}
		}
	}`

	var result struct {
		Repository struct {
			IssueTypes struct {
				Nodes []Type `json:"nodes"`
			} `json:"issueTypes"`
		} `json:"repository"`
	}

	err := client.Query(query, map[string]interface{}{
		"owner": owner,
		"repo":  repo,
	}, &result)
	if err != nil {
		return nil, err
	}
	return result.Repository.IssueTypes.Nodes, nil
}

// SetType sets the type of an issue using GraphQL mutation.
func SetType(client ghcli.Querier, issueID, issueTypeID string) error {
	query := `mutation($issueId: ID!, $issueTypeId: ID!) {
		updateIssue(input: {id: $issueId, issueTypeId: $issueTypeId}) {
			issue {
				id
			}
		}
	}`

	var result struct {
		UpdateIssue struct {
			Issue struct {
				ID string `json:"id"`
			} `json:"issue"`
		} `json:"updateIssue"`
	}

	return client.Query(query, map[string]interface{}{
		"issueId":     issueID,
		"issueTypeId": issueTypeID,
	}, &result)
}

// FindTypeByName finds an issue type by name from the list.
func FindTypeByName(types []Type, name string) (*Type, error) {
	for _, t := range types {
		if t.Name == name {
			return &t, nil
		}
	}
	available := make([]string, len(types))
	for i, t := range types {
		available[i] = t.Name
	}
	return nil, fmt.Errorf("issue type %q not found, available types: %v", name, available)
}
