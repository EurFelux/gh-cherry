package issue

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// ListTypes fetches and prints available issue types for a repository.
func ListTypes(client ghcli.Querier, repoFlag string, w io.Writer) error {
	owner, repo, err := ResolveRepo(repoFlag)
	if err != nil {
		return err
	}

	types, err := FetchTypes(client, owner, repo)
	if err != nil {
		return err
	}

	if len(types) == 0 {
		_, err := fmt.Fprintf(w, "No issue types configured for %s/%s\n", owner, repo)
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	for _, t := range types {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", t.Name, t.ID); err != nil {
			return err
		}
	}
	return tw.Flush()
}
