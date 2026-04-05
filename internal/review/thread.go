package review

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// ValidSides contains the allowed values for the side parameter.
var ValidSides = []string{"LEFT", "RIGHT"}

// AddThreadOptions holds the options for adding a review thread.
type AddThreadOptions struct {
	ReviewID  string
	Path      string
	Line      int
	Body      string
	Side      string
	StartLine int
	StartSide string
}

// AddThreadResult holds the result of adding a review thread.
type AddThreadResult struct {
	ThreadID string `json:"threadId"`
}

// AddThread adds an inline comment thread to a pending pull request review.
func AddThread(client ghcli.Querier, opts AddThreadOptions) (*AddThreadResult, error) {
	if opts.Side == "" {
		opts.Side = "RIGHT"
	}
	if err := validateSide(opts.Side); err != nil {
		return nil, err
	}
	if opts.StartSide != "" {
		if err := validateSide(opts.StartSide); err != nil {
			return nil, fmt.Errorf("invalid start-side: %w", err)
		}
	}

	variables := map[string]any{
		"reviewId": opts.ReviewID,
		"path":     opts.Path,
		"line":     opts.Line,
		"side":     opts.Side,
		"body":     opts.Body,
	}
	if opts.StartLine > 0 {
		variables["startLine"] = opts.StartLine
		if opts.StartSide == "" {
			opts.StartSide = "RIGHT"
		}
		variables["startSide"] = opts.StartSide
	}

	query := `mutation($reviewId: ID!, $path: String!, $line: Int!, $side: DiffSide!, $body: String!, $startLine: Int, $startSide: DiffSide) {
		addPullRequestReviewThread(input: {
			pullRequestReviewId: $reviewId,
			path: $path,
			line: $line,
			side: $side,
			body: $body,
			startLine: $startLine,
			startSide: $startSide
		}) {
			thread {
				id
			}
		}
	}`

	var result struct {
		AddPullRequestReviewThread struct {
			Thread struct {
				ID string `json:"id"`
			} `json:"thread"`
		} `json:"addPullRequestReviewThread"`
	}

	if err := client.Query(query, variables, &result); err != nil {
		return nil, fmt.Errorf("add review thread: %w", err)
	}

	return &AddThreadResult{
		ThreadID: result.AddPullRequestReviewThread.Thread.ID,
	}, nil
}

// PrintResult writes the result as JSON to the given writer.
func PrintResult(result *AddThreadResult, w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(result)
}

func validateSide(side string) error {
	if !slices.Contains(ValidSides, side) {
		return fmt.Errorf("invalid side %q, must be LEFT or RIGHT", side)
	}
	return nil
}

// ReadBodyFile reads the body content from a file.
func ReadBodyFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read body file: %w", err)
	}
	return string(data), nil
}
