---
name: create-issue
description: Create GitHub issues for the gh-cherry repo (EurFelux/gh-cherry) using the correct issue template. Use when the user wants to file an issue, report a bug, request a feature, or create a task for this project.
---

# Create Issue

Create issues on `EurFelux/gh-cherry` using `gh issue create` with the matching template format.

## Template Selection

| User intent | Label |
|---|---|
| New command, enhancement, new flag | `enhancement` |
| Something broken, unexpected behavior | `bug` |
| Refactoring, chore, internal work | `task` |

## Issue Body Format

### Feature (`--label enhancement`)

```
## Description
{what to add or improve}

## Motivation
{why this is needed}

## Proposed Solution
{how it should work, include CLI usage examples}

## Alternatives Considered
{other approaches, or omit if none}
```

### Bug (`--label bug`)

```
## What happened?
{description of the bug}

## Expected behavior
{what should happen}

## Steps to reproduce
1. Run `gh cherry ...`
2. ...
```

### Task (`--label task`)

```
## Description
{what needs to be done}

## Acceptance Criteria
- [ ] {criterion}
```

## Command

```bash
gh issue create --repo EurFelux/gh-cherry \
  --title "<concise title>" \
  --label "<label>" \
  --body "$(cat <<'EOF'
<body matching template above>
EOF
)"
```
