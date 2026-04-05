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

// ReplyThreadOptions holds the options for replying to a review thread.
type ReplyThreadOptions struct {
	ThreadID string
	Body     string
}

// ReplyThreadResult holds the result of replying to a review thread.
type ReplyThreadResult struct {
	CommentID string `json:"commentId"`
}

// ReplyToThread adds a reply to an existing pull request review thread.
func ReplyToThread(client ghcli.Querier, opts ReplyThreadOptions) (*ReplyThreadResult, error) {
	query := `mutation($threadId: ID!, $body: String!) {
		addPullRequestReviewThreadReply(input: {
			pullRequestReviewThreadId: $threadId,
			body: $body
		}) {
			comment {
				id
			}
		}
	}`

	var result struct {
		AddPullRequestReviewThreadReply struct {
			Comment struct {
				ID string `json:"id"`
			} `json:"comment"`
		} `json:"addPullRequestReviewThreadReply"`
	}

	if err := client.Query(query, map[string]any{
		"threadId": opts.ThreadID,
		"body":     opts.Body,
	}, &result); err != nil {
		return nil, fmt.Errorf("reply to thread: %w", err)
	}

	return &ReplyThreadResult{
		CommentID: result.AddPullRequestReviewThreadReply.Comment.ID,
	}, nil
}

// ListThreadsOptions holds the options for listing review threads.
type ListThreadsOptions struct {
	Owner      string
	Repo       string
	PRNumber   int
	Unresolved bool
	Mine       bool
}

// ThreadInfo represents a review thread in the list output.
type ThreadInfo struct {
	ID           string       `json:"id"`
	Path         string       `json:"path"`
	Line         int          `json:"line"`
	IsResolved   bool         `json:"isResolved"`
	CommentCount int          `json:"commentCount"`
	FirstComment FirstComment `json:"firstComment"`
}

// FirstComment represents the first comment in a thread.
type FirstComment struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

// ListThreads lists all review threads for a pull request.
func ListThreads(client ghcli.Querier, opts ListThreadsOptions) ([]ThreadInfo, error) {
	var viewer string
	if opts.Mine {
		v, err := fetchViewer(client)
		if err != nil {
			return nil, err
		}
		viewer = v
	}

	var allThreads []ThreadInfo
	var cursor *string

	for {
		threads, pageInfo, err := fetchThreadPage(client, opts.Owner, opts.Repo, opts.PRNumber, cursor)
		if err != nil {
			return nil, fmt.Errorf("list review threads: %w", err)
		}

		for _, t := range threads {
			if opts.Unresolved && t.IsResolved {
				continue
			}
			if opts.Mine && t.FirstComment.Author != viewer {
				continue
			}
			allThreads = append(allThreads, t)
		}

		if !pageInfo.HasNextPage {
			break
		}
		cursor = &pageInfo.EndCursor
	}

	if allThreads == nil {
		allThreads = []ThreadInfo{}
	}

	return allThreads, nil
}

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

func fetchThreadPage(client ghcli.Querier, owner, repo string, prNumber int, after *string) ([]ThreadInfo, pageInfo, error) {
	query := `query($owner: String!, $repo: String!, $number: Int!, $after: String) {
		repository(owner: $owner, name: $repo) {
			pullRequest(number: $number) {
				reviewThreads(first: 100, after: $after) {
					pageInfo {
						hasNextPage
						endCursor
					}
					nodes {
						id
						path
						line
						isResolved
						comments(first: 1) {
							totalCount
							nodes {
								author { login }
								body
							}
						}
					}
				}
			}
		}
	}`

	vars := map[string]any{
		"owner":  owner,
		"repo":   repo,
		"number": prNumber,
	}
	if after != nil {
		vars["after"] = *after
	}

	var result struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					PageInfo pageInfo `json:"pageInfo"`
					Nodes    []struct {
						ID         string `json:"id"`
						Path       string `json:"path"`
						Line       int    `json:"line"`
						IsResolved bool   `json:"isResolved"`
						Comments   struct {
							TotalCount int `json:"totalCount"`
							Nodes      []struct {
								Author struct {
									Login string `json:"login"`
								} `json:"author"`
								Body string `json:"body"`
							} `json:"nodes"`
						} `json:"comments"`
					} `json:"nodes"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}

	if err := client.Query(query, vars, &result); err != nil {
		return nil, pageInfo{}, err
	}

	nodes := result.Repository.PullRequest.ReviewThreads.Nodes
	threads := make([]ThreadInfo, 0, len(nodes))
	for _, n := range nodes {
		t := ThreadInfo{
			ID:           n.ID,
			Path:         n.Path,
			Line:         n.Line,
			IsResolved:   n.IsResolved,
			CommentCount: n.Comments.TotalCount,
		}
		if len(n.Comments.Nodes) > 0 {
			t.FirstComment = FirstComment{
				Author: n.Comments.Nodes[0].Author.Login,
				Body:   n.Comments.Nodes[0].Body,
			}
		}
		threads = append(threads, t)
	}

	return threads, result.Repository.PullRequest.ReviewThreads.PageInfo, nil
}

func fetchViewer(client ghcli.Querier) (string, error) {
	query := `query { viewer { login } }`

	var result struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
	}

	if err := client.Query(query, nil, &result); err != nil {
		return "", fmt.Errorf("fetch viewer: %w", err)
	}

	return result.Viewer.Login, nil
}

// ResolveThreadResult holds the result of resolving or unresolving a thread.
type ResolveThreadResult struct {
	ID         string `json:"id"`
	IsResolved bool   `json:"isResolved"`
}

// ResolveThread resolves a review thread.
func ResolveThread(client ghcli.Querier, threadID string) (*ResolveThreadResult, error) {
	query := `mutation($threadId: ID!) {
		resolveReviewThread(input: { threadId: $threadId }) {
			thread {
				id
				isResolved
			}
		}
	}`

	var result struct {
		ResolveReviewThread struct {
			Thread struct {
				ID         string `json:"id"`
				IsResolved bool   `json:"isResolved"`
			} `json:"thread"`
		} `json:"resolveReviewThread"`
	}

	if err := client.Query(query, map[string]any{
		"threadId": threadID,
	}, &result); err != nil {
		return nil, fmt.Errorf("resolve thread: %w", err)
	}

	t := result.ResolveReviewThread.Thread
	return &ResolveThreadResult{ID: t.ID, IsResolved: t.IsResolved}, nil
}

// UnresolveThread unresolves a review thread.
func UnresolveThread(client ghcli.Querier, threadID string) (*ResolveThreadResult, error) {
	query := `mutation($threadId: ID!) {
		unresolveReviewThread(input: { threadId: $threadId }) {
			thread {
				id
				isResolved
			}
		}
	}`

	var result struct {
		UnresolveReviewThread struct {
			Thread struct {
				ID         string `json:"id"`
				IsResolved bool   `json:"isResolved"`
			} `json:"thread"`
		} `json:"unresolveReviewThread"`
	}

	if err := client.Query(query, map[string]any{
		"threadId": threadID,
	}, &result); err != nil {
		return nil, fmt.Errorf("unresolve thread: %w", err)
	}

	t := result.UnresolveReviewThread.Thread
	return &ResolveThreadResult{ID: t.ID, IsResolved: t.IsResolved}, nil
}

// EditCommentResult holds the result of editing a review comment.
type EditCommentResult struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

// EditComment updates the body of a pull request review comment.
func EditComment(client ghcli.Querier, commentID, body string) (*EditCommentResult, error) {
	query := `mutation($commentId: ID!, $body: String!) {
		updatePullRequestReviewComment(input: {
			pullRequestReviewCommentId: $commentId,
			body: $body
		}) {
			pullRequestReviewComment {
				id
				body
			}
		}
	}`

	var result struct {
		UpdatePullRequestReviewComment struct {
			PullRequestReviewComment struct {
				ID   string `json:"id"`
				Body string `json:"body"`
			} `json:"pullRequestReviewComment"`
		} `json:"updatePullRequestReviewComment"`
	}

	if err := client.Query(query, map[string]any{
		"commentId": commentID,
		"body":      body,
	}, &result); err != nil {
		return nil, fmt.Errorf("edit comment: %w", err)
	}

	c := result.UpdatePullRequestReviewComment.PullRequestReviewComment
	return &EditCommentResult{ID: c.ID, Body: c.Body}, nil
}

// DeleteCommentResult holds the result of deleting a review comment.
type DeleteCommentResult struct {
	Deleted string `json:"deleted"`
}

// DeleteComment deletes a pull request review comment.
func DeleteComment(client ghcli.Querier, commentID string) (*DeleteCommentResult, error) {
	query := `mutation($commentId: ID!) {
		deletePullRequestReviewComment(input: { id: $commentId }) {
			pullRequestReviewComment {
				id
			}
		}
	}`

	var result struct {
		DeletePullRequestReviewComment struct {
			PullRequestReviewComment struct {
				ID string `json:"id"`
			} `json:"pullRequestReviewComment"`
		} `json:"deletePullRequestReviewComment"`
	}

	if err := client.Query(query, map[string]any{
		"commentId": commentID,
	}, &result); err != nil {
		return nil, fmt.Errorf("delete comment: %w", err)
	}

	return &DeleteCommentResult{
		Deleted: result.DeletePullRequestReviewComment.PullRequestReviewComment.ID,
	}, nil
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
