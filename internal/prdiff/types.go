package prdiff

// PRFile represents a file returned by the GitHub PR files endpoint.
type PRFile struct {
	Filename string `json:"filename"`
	Patch    string `json:"patch"`
	Status   string `json:"status"`
}

// AnnotatedLine represents a single diff line with side and line number annotation.
type AnnotatedLine struct {
	Side       string // "L" or "R"
	LineNumber int
	Prefix     string // " ", "+", or "-"
	Content    string // line content without the diff prefix
}

// AnnotatedHunk represents a parsed hunk with annotated lines.
type AnnotatedHunk struct {
	Header string
	Lines  []AnnotatedLine
}

// AnnotatedFile is a fully annotated file diff.
type AnnotatedFile struct {
	Filename string
	Hunks    []AnnotatedHunk
}
