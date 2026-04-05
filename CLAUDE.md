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
- **CLI layer** (`cmd/`): Cobra commands. `root.go` defines the root command; subcommand files register themselves via `init()`.
- **Business logic** (`internal/`):
  - `ghcli/` — thin wrapper around `gh api graphql` for executing GraphQL queries
  - `issue/` — issue creation with type support. Creates via `gh issue create`, then sets type via GraphQL mutation.
- GitHub API interactions use `gh.Exec()` (shelling out to `gh`) or GraphQL via `internal/ghcli`.

## Development Rules

- Every feature must have a plan approved by the user before implementation
- Every feature must include tests
- Before committing, all changes must be reviewed by the code-reviewer agent until approved

## Conventions

- Formatter: **gofumpt** with extra rules enabled
- Linter: **golangci-lint v2** with revive (exported comments required, unused parameters flagged)
- Pre-commit hooks managed by **prek** (`.pre-commit-config.yaml`)
- Releases: tag `v*` triggers GitHub Actions with `cli/gh-extension-precompile` for cross-platform binaries
