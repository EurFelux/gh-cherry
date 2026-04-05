package ghcli

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Querier executes GraphQL queries against the GitHub API.
type Querier interface {
	Query(query string, variables map[string]interface{}, result interface{}) error
}

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

// RESTQuerier executes REST API requests against the GitHub API.
type RESTQuerier interface {
	Get(path string, response interface{}) error
}

// RESTClient wraps a GitHub REST client.
type RESTClient struct {
	rest *api.RESTClient
}

// NewRESTClient creates a new RESTClient using the default gh authentication.
func NewRESTClient() (*RESTClient, error) {
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("create rest client: %w", err)
	}
	return &RESTClient{rest: rest}, nil
}

// Get executes a GET request and decodes the JSON response.
func (c *RESTClient) Get(path string, response interface{}) error {
	return c.rest.Get(path, response)
}
