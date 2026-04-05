package prdiff

import (
	"fmt"
	"io"
	"strconv"
)

// Format writes annotated files to the writer in the specified format.
func Format(files []AnnotatedFile, w io.Writer) error {
	for i, f := range files {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintln(w, f.Filename); err != nil {
			return err
		}

		for _, h := range f.Hunks {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(w, h.Header); err != nil {
				return err
			}

			width := lineNumberWidth(h.Lines)
			for _, l := range h.Lines {
				label := l.Side + strconv.Itoa(l.LineNumber)
				if _, err := fmt.Fprintf(w, "%s %-*s | %s%s\n",
					l.Prefix, width, label, l.Prefix, l.Content); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func lineNumberWidth(lines []AnnotatedLine) int {
	maxWidth := 0
	for _, l := range lines {
		w := len(l.Side) + len(strconv.Itoa(l.LineNumber))
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}
