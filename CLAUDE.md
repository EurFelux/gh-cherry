# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`gh-cherry` is a GitHub CLI extension written in Go. It enhances `gh` with features not available natively, such as issue type support. Installed as `gh cherry`.

## Commands

```bash
# Build
go build -o gh-cherry .

# Lint (golangci-lint v2)
golangci-lint run

# Format
gofumpt -w .

# Test
go test ./...

# Test single package
go test ./internal/issue/

# Test single function
go test ./internal/issue/ -run TestFindTypeByName

# Install git hooks (uses prek, a pre-commit alternative)
prek install
```

## Architecture

- **Entry point**: `main.go` → `cmd.Execute()`
- **CLI layer** (`cmd/`): Cobra commands. `root.go` defines the root command; subcommand files (`issue.go`, `pr.go`, `review.go`) register themselves via `init()`.
- **Business logic** (`internal/`):
  - `ghcli/` — `Querier` (GraphQL) and `RESTQuerier` (REST) interfaces wrapping `go-gh`'s API clients. Both are mockable for testing.
  - `issue/` — issue creation with type support, sub-issue management. Creates via `gh issue create` then sets type via GraphQL mutation. Also: `FetchTypes()`, `AddSubIssue()`, `RemoveSubIssue()`, `ListSubIssues()`.
  - `prdiff/` — annotated PR diff output. Pipeline: `FetchPRFiles()` (REST) → `AnnotateFile()` (parse unified diff) → `Format()` (L/R line annotations).
  - `review/` — full PR review lifecycle. `StartReview`, `SubmitReview`, `EditReview`, `PreviewReview`, `ViewReviews`, `AddThread`, `ReplyToThread`. Each returns a typed result struct.
- GraphQL operations go through `ghcli.Querier`, REST through `ghcli.RESTQuerier`. Only `issue.Create()` uses `gh.Exec()` for its interactive features.
- All command handlers output JSON via `json.NewEncoder(os.Stdout).Encode()` for consistent, parseable output.

## Development Rules

- Every feature must have a plan approved by the user before implementation
- Every feature must include tests
- Before committing, all changes must be reviewed by the code-reviewer agent until approved
- PR description must include `Fixes #<issue_number>` to auto-close the related issue

## Branching & Merge Strategy

- **Feature development and bug fixes** must be done on a separate branch, include a CHANGELOG entry, and be squash-merged into `main`
- **Documentation-only changes** (e.g., CLAUDE.md, README) can be committed directly to `main`

## Conventions

- Formatter: **gofumpt** with extra rules enabled
- Linter: **golangci-lint v2** with revive (exported comments required, unused parameters flagged)
- Pre-commit hooks managed by **prek** (`.pre-commit-config.yaml`)
- Releases: tag `v*` triggers GitHub Actions with `cli/gh-extension-precompile` for cross-platform binaries
