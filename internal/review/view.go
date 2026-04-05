package review

import (
	"fmt"
	"strings"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// ViewOptions holds the filter options for viewing reviews and threads.
type ViewOptions struct {
	Owner      string
	Repo       string
	PRNumber   int
	Reviewer   string
	State      string
	Unresolved bool
	Tail       int
}

// ViewReview represents a single review in the view output.
type ViewReview struct {
	ID          string `json:"id"`
	Author      string `json:"author"`
	State       string `json:"state"`
	Body        string `json:"body,omitempty"`
	SubmittedAt string `json:"submittedAt,omitempty"`
}

// ViewThreadComment represents a comment within a review thread.
type ViewThreadComment struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

// ViewThread represents a review thread in the view output.
type ViewThread struct {
	ID         string              `json:"id"`
	Path       string              `json:"path"`
	Line       int                 `json:"line"`
	IsResolved bool                `json:"isResolved"`
	Comments   []ViewThreadComment `json:"comments"`
}

// ViewResult holds the complete view of reviews and threads for a PR.
type ViewResult struct {
	Reviews []ViewReview `json:"reviews"`
	Threads []ViewThread `json:"threads"`
}

// ViewReviews fetches all reviews and review threads for a PR with optional filtering.
func ViewReviews(client ghcli.Querier, opts ViewOptions) (*ViewResult, error) {
	reviews, err := fetchAllReviews(client, opts.Owner, opts.Repo, opts.PRNumber)
	if err != nil {
		return nil, fmt.Errorf("fetch reviews: %w", err)
	}

	threads, err := fetchAllThreads(client, opts.Owner, opts.Repo, opts.PRNumber)
	if err != nil {
		return nil, fmt.Errorf("fetch threads: %w", err)
	}

	reviews = filterReviews(reviews, opts)
	threads = filterThreads(threads, opts)

	return &ViewResult{
		Reviews: reviews,
		Threads: threads,
	}, nil
}

func fetchAllReviews(client ghcli.Querier, owner, repo string, prNumber int) ([]ViewReview, error) {
	query := `query($owner: String!, $repo: String!, $number: Int!, $cursor: String) {
		repository(owner: $owner, name: $repo) {
			pullRequest(number: $number) {
				reviews(first: 100, after: $cursor) {
					nodes {
						id
						author { login }
						state
						body
						submittedAt
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}`

	var allReviews []ViewReview
	var cursor *string

	for {
		vars := map[string]any{
			"owner":  owner,
			"repo":   repo,
			"number": prNumber,
		}
		if cursor != nil {
			vars["cursor"] = *cursor
		}

		var result struct {
			Repository struct {
				PullRequest struct {
					Reviews struct {
						Nodes []struct {
							ID     string `json:"id"`
							Author struct {
								Login string `json:"login"`
							} `json:"author"`
							State       string `json:"state"`
							Body        string `json:"body"`
							SubmittedAt string `json:"submittedAt"`
						} `json:"nodes"`
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"reviews"`
				} `json:"pullRequest"`
			} `json:"repository"`
		}

		if err := client.Query(query, vars, &result); err != nil {
			return nil, err
		}

		for _, n := range result.Repository.PullRequest.Reviews.Nodes {
			allReviews = append(allReviews, ViewReview{
				ID:          n.ID,
				Author:      n.Author.Login,
				State:       n.State,
				Body:        n.Body,
				SubmittedAt: n.SubmittedAt,
			})
		}

		pi := result.Repository.PullRequest.Reviews.PageInfo
		if !pi.HasNextPage {
			break
		}
		cursor = &pi.EndCursor
	}

	return allReviews, nil
}

func fetchAllThreads(client ghcli.Querier, owner, repo string, prNumber int) ([]ViewThread, error) {
	query := `query($owner: String!, $repo: String!, $number: Int!, $cursor: String) {
		repository(owner: $owner, name: $repo) {
			pullRequest(number: $number) {
				reviewThreads(first: 100, after: $cursor) {
					nodes {
						id
						path
						line
						isResolved
						comments(first: 100) {
							nodes {
								id
								author { login }
								body
								createdAt
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}`

	var allThreads []ViewThread
	var cursor *string

	for {
		vars := map[string]any{
			"owner":  owner,
			"repo":   repo,
			"number": prNumber,
		}
		if cursor != nil {
			vars["cursor"] = *cursor
		}

		var result struct {
			Repository struct {
				PullRequest struct {
					ReviewThreads struct {
						Nodes []struct {
							ID         string `json:"id"`
							Path       string `json:"path"`
							Line       int    `json:"line"`
							IsResolved bool   `json:"isResolved"`
							Comments   struct {
								Nodes []struct {
									ID     string `json:"id"`
									Author struct {
										Login string `json:"login"`
									} `json:"author"`
									Body      string `json:"body"`
									CreatedAt string `json:"createdAt"`
								} `json:"nodes"`
							} `json:"comments"`
						} `json:"nodes"`
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"reviewThreads"`
				} `json:"pullRequest"`
			} `json:"repository"`
		}

		if err := client.Query(query, vars, &result); err != nil {
			return nil, err
		}

		for _, n := range result.Repository.PullRequest.ReviewThreads.Nodes {
			comments := make([]ViewThreadComment, len(n.Comments.Nodes))
			for i, c := range n.Comments.Nodes {
				comments[i] = ViewThreadComment{
					ID:        c.ID,
					Author:    c.Author.Login,
					Body:      c.Body,
					CreatedAt: c.CreatedAt,
				}
			}
			allThreads = append(allThreads, ViewThread{
				ID:         n.ID,
				Path:       n.Path,
				Line:       n.Line,
				IsResolved: n.IsResolved,
				Comments:   comments,
			})
		}

		pi := result.Repository.PullRequest.ReviewThreads.PageInfo
		if !pi.HasNextPage {
			break
		}
		cursor = &pi.EndCursor
	}

	return allThreads, nil
}

func filterReviews(reviews []ViewReview, opts ViewOptions) []ViewReview {
	var filtered []ViewReview
	for _, r := range reviews {
		if opts.Reviewer != "" && !strings.EqualFold(r.Author, opts.Reviewer) {
			continue
		}
		if opts.State != "" && !strings.EqualFold(r.State, opts.State) {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

func filterThreads(threads []ViewThread, opts ViewOptions) []ViewThread {
	var filtered []ViewThread
	for _, t := range threads {
		if opts.Unresolved && t.IsResolved {
			continue
		}
		if opts.Tail > 0 && len(t.Comments) > opts.Tail {
			t.Comments = t.Comments[len(t.Comments)-opts.Tail:]
		}
		filtered = append(filtered, t)
	}
	return filtered
}
