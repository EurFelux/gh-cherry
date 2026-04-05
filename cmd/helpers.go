package cmd

import (
	"fmt"
	"os"

	"github.com/EurFelux/gh-cherry/internal/jqutil"
	"github.com/spf13/cobra"
)

// encodeJSON writes v as JSON to stdout, applying the --jq filter if set.
func encodeJSON(cmd *cobra.Command, v any) error {
	jqExpr, _ := cmd.Root().PersistentFlags().GetString("jq")
	return jqutil.Encode(os.Stdout, v, jqExpr)
}

// resolveBody reads the review body from --body or --body-file.
func resolveBody(cmd *cobra.Command) (string, error) {
	body, _ := cmd.Flags().GetString("body")
	if body != "" {
		return body, nil
	}

	bodyFile, _ := cmd.Flags().GetString("body-file")
	if bodyFile == "" {
		return "", nil
	}

	data, err := os.ReadFile(bodyFile)
	if err != nil {
		return "", fmt.Errorf("read body file: %w", err)
	}
	return string(data), nil
}
