# gh-cherry

A [GitHub CLI](https://cli.github.com/) extension that provides enhanced issue management and PR review capabilities not available natively in `gh`.

## Features

- **Issue types** — Create issues with [sub-issue type support](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/about-issue-types) (`gh cherry issue create -T Bug`)
- **Annotated PR diffs** — View diffs with left/right line numbers, designed for AI agent review workflows
- **Full review lifecycle** — Start, preview, submit, view, and edit PR reviews entirely from the terminal
- **Review threads** — View and reply to review comment threads

## Installation

```bash
gh extension install EurFelux/gh-cherry
```

Requires [GitHub CLI](https://cli.github.com/) v2.0+.

## Usage

### Issues

```bash
# Create an issue with a type
gh cherry issue create -T Bug

# List available issue types for a repository
gh cherry issue types
```

### PR Diff

```bash
# View annotated diff for a PR
gh cherry pr diff 123
```

### Reviews

```bash
# Start a pending review
gh cherry review start 123

# Add comments, then preview before submitting
gh cherry review preview 123

# Submit the review
gh cherry review submit 123 --event APPROVE --body "LGTM"

# View all reviews and threads on a PR
gh cherry review view 123

# Edit a submitted review's body
gh cherry review edit 123 <review-id> --body "Updated feedback"

# Reply to an existing review thread
gh cherry review thread reply 123 <thread-id> --body "Thanks, fixed"
```

## Claude Code Skill

This repo includes a [Claude Code skill](https://docs.anthropic.com/en/docs/claude-code/skills) at `.claude/skills/gh-cherry-usage/` that teaches Claude how to use `gh cherry` commands. Once installed, Claude can assist with issue creation, PR reviews, and diff viewing through natural language.

## Building from source

```bash
go build -o gh-cherry .
```

## License

[MIT](LICENSE)
