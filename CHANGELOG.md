# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `gh cherry review thread list` command to list review threads with `--unresolved` and `--mine` filters
- `gh cherry review thread reply` command to reply to existing review threads
- `gh cherry issue create` command with issue type support (`-T` flag)
- `gh cherry issue types` command to list available issue types for a repository
- `gh cherry pr diff` command with annotated L/R line numbers for AI agent review workflows
- `gh cherry review start` command to create or reuse a pending PR review via GraphQL
- `gh cherry review submit` command to submit a pending review with APPROVE/REQUEST_CHANGES/COMMENT
- `gh cherry review preview` command to preview pending review comments before submitting
- `gh cherry review view` command to view all reviews and threads for a PR with filtering
- `gh cherry review edit` command to edit a submitted review's body text
- `ghcli.Querier` interface for mockable GraphQL operations
- `ghcli.RESTQuerier` interface for mockable REST operations
- Pre-commit hooks via prek (gofumpt, golangci-lint, typos)
- CI release workflow with `gh-extension-precompile`
- MIT license

### Changed

- Migrated GraphQL client from raw HTTP to `go-gh` `api.DefaultGraphQLClient()` with `Querier` interface
