package issue

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractIssueNumber(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    int
		wantErr bool
	}{
		{
			name: "standard URL",
			url:  "https://github.com/owner/repo/issues/42",
			want: 42,
		},
		{
			name: "URL with trailing whitespace",
			url:  "https://github.com/owner/repo/issues/7\n",
			want: 7,
		},
		{
			name:    "no slash",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "non-numeric suffix",
			url:     "https://github.com/owner/repo/issues/abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractIssueNumber(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseRepoFlag(t *testing.T) {
	tests := []struct {
		name      string
		flag      string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "valid",
			flag:      "owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:    "no slash",
			flag:    "owner-repo",
			wantErr: true,
		},
		{
			name:    "empty owner",
			flag:    "/repo",
			wantErr: true,
		},
		{
			name:    "empty repo",
			flag:    "owner/",
			wantErr: true,
		},
		{
			name:      "with nested slash",
			flag:      "owner/repo/extra",
			wantOwner: "owner",
			wantRepo:  "repo/extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseRepoFlag(tt.flag)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOwner, owner)
			assert.Equal(t, tt.wantRepo, repo)
		})
	}
}

func TestAddAsSubIssue(t *testing.T) {
	// buildMock creates a mock that handles: GetIssueNodeID(parent), GetIssueNodeID(child), AddSubIssue mutation.
	buildMock := func(t *testing.T, parentNum, childNum int, parentNodeID, childNodeID string) *mockQuerier {
		t.Helper()
		call := 0
		return &mockQuerier{
			queryFunc: func(_ string, variables map[string]interface{}, result interface{}) error {
				call++
				switch call {
				case 1: // GetIssueNodeID for parent
					assert.Equal(t, parentNum, variables["number"])
					type respType = struct {
						Repository struct {
							Issue struct {
								ID string `json:"id"`
							} `json:"issue"`
						} `json:"repository"`
					}
					r := result.(*respType)
					r.Repository.Issue.ID = parentNodeID
					return nil
				case 2: // GetIssueNodeID for child
					assert.Equal(t, childNum, variables["number"])
					type respType = struct {
						Repository struct {
							Issue struct {
								ID string `json:"id"`
							} `json:"issue"`
						} `json:"repository"`
					}
					r := result.(*respType)
					r.Repository.Issue.ID = childNodeID
					return nil
				case 3: // AddSubIssue mutation
					assert.Equal(t, parentNodeID, variables["issueId"])
					assert.Equal(t, childNodeID, variables["subIssueId"])
					return nil
				default:
					t.Fatalf("unexpected call %d", call)
					return nil
				}
			},
		}
	}

	t.Run("success", func(t *testing.T) {
		mock := buildMock(t, 2, 10, "I_parent", "I_child")
		opts := CreateOptions{
			Parent: 2,
			Repo:   "owner/repo",
			Client: mock,
		}
		err := addAsSubIssue(opts, "https://github.com/owner/repo/issues/10")
		require.NoError(t, err)
	})

	t.Run("parent not found", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]interface{}, _ interface{}) error {
				return nil
			},
		}
		opts := CreateOptions{
			Parent: 999,
			Repo:   "owner/repo",
			Client: mock,
		}
		err := addAsSubIssue(opts, "https://github.com/owner/repo/issues/10")
		assert.ErrorContains(t, err, "resolve parent issue")
	})

	t.Run("invalid issue URL", func(t *testing.T) {
		mock := &mockQuerier{}
		opts := CreateOptions{
			Parent: 2,
			Repo:   "owner/repo",
			Client: mock,
		}
		err := addAsSubIssue(opts, "not-a-url")
		assert.Error(t, err)
	})

	t.Run("mutation fails", func(t *testing.T) {
		call := 0
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]interface{}, result interface{}) error {
				call++
				switch call {
				case 1, 2:
					type respType = struct {
						Repository struct {
							Issue struct {
								ID string `json:"id"`
							} `json:"issue"`
						} `json:"repository"`
					}
					r := result.(*respType)
					r.Repository.Issue.ID = fmt.Sprintf("I_%d", call)
					return nil
				case 3:
					return fmt.Errorf("insufficient permissions")
				default:
					return nil
				}
			},
		}
		opts := CreateOptions{
			Parent: 2,
			Repo:   "owner/repo",
			Client: mock,
		}
		err := addAsSubIssue(opts, "https://github.com/owner/repo/issues/10")
		assert.ErrorContains(t, err, "insufficient permissions")
	})
}

func TestGetIssueNodeID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, variables map[string]interface{}, result interface{}) error {
				assert.Equal(t, "owner", variables["owner"])
				assert.Equal(t, "repo", variables["repo"])
				assert.Equal(t, 42, variables["number"])
				type respType = struct {
					Repository struct {
						Issue struct {
							ID string `json:"id"`
						} `json:"issue"`
					} `json:"repository"`
				}
				r := result.(*respType)
				r.Repository.Issue.ID = "I_abc123"
				return nil
			},
		}

		id, err := GetIssueNodeID(mock, "owner", "repo", 42)
		require.NoError(t, err)
		assert.Equal(t, "I_abc123", id)
	})

	t.Run("not found", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]interface{}, _ interface{}) error {
				return nil // empty response, ID stays ""
			},
		}

		_, err := GetIssueNodeID(mock, "owner", "repo", 999)
		assert.ErrorContains(t, err, "issue #999 not found")
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]interface{}, _ interface{}) error {
				return fmt.Errorf("timeout")
			},
		}

		_, err := GetIssueNodeID(mock, "owner", "repo", 1)
		assert.ErrorContains(t, err, "timeout")
	})
}
