package prdiff

import (
	"fmt"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

const (
	perPage  = 100
	maxPages = 50
)

// FetchPRFiles fetches all changed files for a pull request, handling pagination.
func FetchPRFiles(client ghcli.RESTQuerier, owner, repo string, prNumber int) ([]PRFile, error) {
	var allFiles []PRFile

	for page := 1; page <= maxPages; page++ {
		path := fmt.Sprintf("repos/%s/%s/pulls/%d/files?per_page=%d&page=%d",
			owner, repo, prNumber, perPage, page)

		var files []PRFile
		if err := client.Get(path, &files); err != nil {
			return nil, fmt.Errorf("fetch PR files (page %d): %w", page, err)
		}

		allFiles = append(allFiles, files...)

		if len(files) < perPage {
			break
		}
	}

	return allFiles, nil
}
