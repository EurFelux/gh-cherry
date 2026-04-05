// Package jqutil provides JSON output with optional jq filtering.
package jqutil

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/itchyny/gojq"
)

// Encode writes v as JSON to w. If jqExpr is non-empty, the value is
// filtered through the jq expression first and each result is written
// as a separate JSON line.
func Encode(w io.Writer, v any, jqExpr string) error {
	if jqExpr == "" {
		return json.NewEncoder(w).Encode(v)
	}

	query, err := gojq.Parse(jqExpr)
	if err != nil {
		return fmt.Errorf("invalid jq expression: %w", err)
	}

	// Convert v to an interface{} that gojq can traverse.
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var input any
	if err := json.Unmarshal(raw, &input); err != nil {
		return err
	}

	iter := query.Run(input)
	for {
		val, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := val.(error); isErr {
			return fmt.Errorf("jq: %w", err)
		}
		if err := json.NewEncoder(w).Encode(val); err != nil {
			return err
		}
	}
	return nil
}
