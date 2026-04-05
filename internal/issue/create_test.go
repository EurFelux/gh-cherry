package issue

import (
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
