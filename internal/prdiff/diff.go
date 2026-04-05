package prdiff

import (
	"io"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// AnnotatedDiff fetches a PR's changed files and writes an annotated diff to w.
func AnnotatedDiff(client ghcli.RESTQuerier, owner, repo string, prNumber int, w io.Writer) error {
	files, err := FetchPRFiles(client, owner, repo, prNumber)
	if err != nil {
		return err
	}

	var annotated []AnnotatedFile
	for _, f := range files {
		if f.Patch == "" {
			continue
		}
		annotated = append(annotated, AnnotateFile(f))
	}

	return Format(annotated, w)
}
