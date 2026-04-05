package prdiff

import (
	"regexp"
	"strconv"
	"strings"
)

var hunkHeaderRe = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

// ParsePatch parses a unified diff patch string into annotated hunks.
func ParsePatch(patch string) []AnnotatedHunk {
	if patch == "" {
		return nil
	}

	lines := strings.Split(patch, "\n")
	var hunks []AnnotatedHunk
	var current *AnnotatedHunk
	var oldLine, newLine int

	for _, line := range lines {
		if matches := hunkHeaderRe.FindStringSubmatch(line); matches != nil {
			oldLine, _ = strconv.Atoi(matches[1])
			newLine, _ = strconv.Atoi(matches[2])
			hunks = append(hunks, AnnotatedHunk{Header: line})
			current = &hunks[len(hunks)-1]
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(line, `\ `) || line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "-"):
			current.Lines = append(current.Lines, AnnotatedLine{
				Side:       "L",
				LineNumber: oldLine,
				Prefix:     "-",
				Content:    line[1:],
			})
			oldLine++
		case strings.HasPrefix(line, "+"):
			current.Lines = append(current.Lines, AnnotatedLine{
				Side:       "R",
				LineNumber: newLine,
				Prefix:     "+",
				Content:    line[1:],
			})
			newLine++
		default:
			// Context line (starts with space or is empty within a hunk)
			content := line
			if len(line) > 0 && line[0] == ' ' {
				content = line[1:]
			}
			current.Lines = append(current.Lines, AnnotatedLine{
				Side:       "R",
				LineNumber: newLine,
				Prefix:     " ",
				Content:    content,
			})
			oldLine++
			newLine++
		}
	}

	return hunks
}

// AnnotateFile parses a PRFile into an AnnotatedFile.
func AnnotateFile(file PRFile) AnnotatedFile {
	return AnnotatedFile{
		Filename: file.Filename,
		Hunks:    ParsePatch(file.Patch),
	}
}
