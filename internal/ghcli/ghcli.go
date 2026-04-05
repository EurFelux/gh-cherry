package ghcli

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Client wraps a GitHub GraphQL client.
type Client struct {
	gql *api.GraphQLClient
}

// NewClient creates a new Client using the default gh authentication.
func NewClient() (*Client, error) {
	gql, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("create graphql client: %w", err)
	}
	return &Client{gql: gql}, nil
}

// Query executes a GraphQL query and decodes the response into result.
func (c *Client) Query(query string, variables map[string]interface{}, result interface{}) error {
	return c.gql.Do(query, variables, result)
}
