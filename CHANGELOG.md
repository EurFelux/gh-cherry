# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `gh cherry issue create` command with issue type support (`-T` flag)
- `gh cherry issue types` command to list available issue types for a repository
- `ghcli.Querier` interface for mockable GraphQL operations
- Pre-commit hooks via prek (gofumpt, golangci-lint, typos)
- CI release workflow with `gh-extension-precompile`
- MIT license

### Changed

- Migrated GraphQL client from raw HTTP to `go-gh` `api.DefaultGraphQLClient()` with `Querier` interface
