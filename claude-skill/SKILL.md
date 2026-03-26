---
name: jira
description: This skill enables interaction with JIRA Cloud for issue tracking. It should be used when the user wants to create, view, update, search, or comment on JIRA issues/tickets. Use this skill for any JIRA-related operations including listing projects, searching with JQL, transitioning issue statuses, managing comments, uploading file attachments, and linking issues.
---

# JIRA

## Overview

This skill provides tools for interacting with JIRA Cloud to manage issues (tickets), projects, and comments using the `jira` CLI (`/opt/homebrew/bin/jira`).

## Scope

This skill is restricted to the following projects only:
- **DD** - Deventer Dragons
- **FW** - Flexwhere
- **ED** - Ed Controls

When searching, listing, or updating issues, always filter by these projects unless explicitly asked otherwise.

## Prerequisites

The CLI loads credentials from `~/.config/jira/.env` automatically. A local `.env` in the current directory is also checked and takes priority if it contains JIRA keys (useful for per-project overrides).

The environment file should contain:
- `JIRA_BASE_URL` - Your JIRA instance URL
- `JIRA_EMAIL` - Your Atlassian account email
- `JIRA_API_TOKEN` - Your Atlassian API token

Run `jira configure` to see configuration help.

## Quick Start

### List Projects
```bash
jira projects list
jira projects get ED
```

### Search Issues
```bash
jira issues search "project = ED AND status = Open"
jira issues search --my-issues
jira issues search --project ED --status "In Progress"
jira issues search --project ED --type Bug --max 10
```

### View an Issue
```bash
jira issues get ED-123
jira issues get ED-123 --comments
jira issues get ED-123 --json
```

### Create an Issue
```bash
jira issues create --project ED --type Task --summary "Issue title"
jira issues create --project ED --type Bug --summary "Bug title" --description "Details here"
jira issues create --project ED --type Story --summary "User story" --assignee "5ca709b3e935e80e41c3e30c" --priority "Prio 1"
jira issues create --project ED --type Task --summary "Deadline task" --due-date 2026-03-15
```

### Update an Issue
```bash
jira issues update ED-123 --summary "New title"
jira issues update ED-123 --priority "Prio 1"
jira issues update ED-123 --assignee "5ca709b3e935e80e41c3e30c"
jira issues update ED-123 --unassign
jira issues update ED-123 --labels "frontend,urgent"
jira issues update ED-123 --due-date 2026-04-01
jira issues update ED-123 --no-due-date
```

### Transition an Issue
```bash
jira issues transition ED-123 --list          # List available transitions
jira issues transition ED-123 "In Progress"   # Move to status
jira issues transition ED-123 "Done"
```

### Delete an Issue
```bash
jira issues delete ED-123
jira issues delete ED-123 --force    # Skip confirmation
```

### Comments
```bash
jira comments list ED-123
jira comments add ED-123 "This is my comment"
jira comments add ED-123 --file /path/to/comment.md
jira comments update ED-123 12345 "Updated comment text"
jira comments delete ED-123 12345
jira comments delete ED-123 12345 --force
```

### Attachments
```bash
jira attachments add ED-123 /path/to/file.pdf
jira attachments add ED-123 /path/to/file.log --filename "session-log.txt"
```

### Inline Images in Descriptions/Comments

To embed attached images inline in descriptions or comments, first attach the file, then reference it using `!filename!` syntax in the description or comment text:

```bash
# 1. Attach the image(s)
jira attachments add ED-123 screenshot.png

# 2. Reference in description (renders inline in JIRA)
jira issues update ED-123 --description "## Screenshot\n\n!screenshot.png!"

# With width option
jira issues update ED-123 --description "!screenshot.png|width=720!"

# In comments too
jira comments add ED-123 "Here's the mockup:\n\n!mockup.jpg|width=600!"
```

When `!filename!` patterns are detected, the CLI automatically uses JIRA's wiki markup renderer (API v2) which supports embedded attachment images. All other content without inline images continues to use the standard ADF/v3 pipeline.

**Important:** The filename in `!filename!` must exactly match the attached file's name. Attach first, then reference.

### Issue Links
```bash
jira links list ED-123                         # List links on an issue
jira links list ED-123 --json                  # JSON output
jira links add ED-123 ED-456 --type "3. Blocks"  # Link two issues
jira links add ED-100 ED-200 --type "4. Duplicate"
jira links delete 12345                        # Delete a link by ID
jira links delete 12345 --force                # Skip confirmation
jira links types                               # List available link types
```

### Users
```bash
jira users me                                  # Show current user
jira users search "naveen"                     # Search by name/email
jira users assignable --project ED             # List assignable users
```

## CLI Reference

### `jira issues search [<jql>]`
Search and list issues using JQL queries or filters.

| Flag | Short | Description |
|------|-------|-------------|
| `--project` | `-p` | Filter by project key |
| `--status` | `-s` | Filter by status name |
| `--assignee` | `-a` | Filter by assignee (account ID or `currentUser()`) |
| `--my-issues` | `-m` | Show only issues assigned to current user |
| `--type` | `-t` | Filter by issue type (Bug, Task, Story, etc.) |
| `--max` | `-n` | Maximum results (default: 50) |
| `--json` | `-j` | Output as JSON |

### `jira issues get <issue-key>`
Display detailed information about an issue.

| Flag | Description |
|------|-------------|
| `--comments` | Include comments |
| `--json` | Output as JSON |

### `jira issues create`
Create a new issue in JIRA.

| Flag | Short | Description |
|------|-------|-------------|
| `--project` | `-p` | Project key (required) |
| `--type` | `-t` | Issue type: Bug, Task, Story, etc. (default: Task) |
| `--summary` | `-s` | Issue summary/title (required) |
| `--description` | `-d` | Issue description (supports markdown and @mentions) |
| `--priority` | | Priority level |
| `--assignee` | `-a` | Assignee account ID |
| `--labels` | `-l` | Comma-separated labels |
| `--due-date` | | Due date in YYYY-MM-DD format |
| `--json` | `-j` | Output as JSON |

### `jira issues update <issue-key>`
Update an existing issue.

| Flag | Short | Description |
|------|-------|-------------|
| `--summary` | `-s` | New summary |
| `--description` | `-d` | New description (supports markdown and @mentions) |
| `--priority` | | New priority |
| `--assignee` | `-a` | New assignee account ID |
| `--unassign` | | Remove assignee |
| `--labels` | `-l` | New comma-separated labels (replaces existing) |
| `--due-date` | | Due date in YYYY-MM-DD format |
| `--no-due-date` | | Remove the due date |

### `jira issues transition <issue-key> [<status>]`
Transition an issue to a new status.

| Flag | Description |
|------|-------------|
| `--list` | List available transitions instead of transitioning |

### `jira issues delete <issue-key>`
Delete an issue.

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation |

### `jira comments list <issue-key>`
List comments on an issue.

| Flag | Short | Description |
|------|-------|-------------|
| `--max` | `-n` | Maximum results (default: 50) |
| `--json` | `-j` | Output as JSON |

### `jira comments add <issue-key> [<text>]`
Add a comment to an issue.

| Flag | Description |
|------|-------------|
| `--file` | Read comment from file |
| `--json` | Output as JSON |

### `jira comments update <issue-key> <comment-id> <text>`
Update an existing comment.

### `jira comments delete <issue-key> <comment-id>`
Delete a comment.

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation |

### `jira attachments add <issue-key> <file-path>`
Add a file attachment to an issue.

| Flag | Description |
|------|-------------|
| `--filename` | Display name for the attachment (defaults to file's basename) |
| `--json` | Output as JSON |

### `jira projects list`
List all accessible projects.

| Flag | Short | Description |
|------|-------|-------------|
| `--max` | `-n` | Maximum results (default: 50) |
| `--json` | `-j` | Output as JSON |

### `jira projects get <project-key>`
Show project details.

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |

### `jira users me`
Show current authenticated user.

### `jira users search <query>`
Search for users by name or email.

| Flag | Short | Description |
|------|-------|-------------|
| `--max` | `-n` | Maximum results (default: 50) |
| `--json` | `-j` | Output as JSON |

### `jira users assignable --project=STRING`
List users assignable to a project.

| Flag | Short | Description |
|------|-------|-------------|
| `--project` | `-p` | Project key |
| `--max` | `-n` | Maximum results (default: 50) |
| `--json` | `-j` | Output as JSON |

## Important: Instance-Specific Configuration

### Priority Names

This JIRA instance uses custom priority names. Do NOT use standard names like "High", "Medium", "Low".

| Priority | Name |
|----------|------|
| Highest | `Prio 1` |
| Normal | `Prio 2` |
| Lowest | `Prio 3` |

### Text Formatting (Markdown Support)

The CLI automatically converts markdown to Atlassian Document Format (ADF) for descriptions and comments. Use standard markdown syntax:

| Syntax | Result |
|--------|--------|
| `@User Name` | User mention (notifies user) |
| `**bold**` or `__bold__` | **bold** text |
| `*italic*` or `_italic_` | *italic* text |
| `` `code` `` | inline code |
| `[text](url)` | hyperlink |
| `# Heading` | Heading (h1-h6) |
| `- item` or `* item` | Bullet list |
| `1. item` | Numbered list |
| ` ``` code ``` ` | Code block |
| `---` | Horizontal rule |
| `\| col \| col \|` | Table |

#### User Mentions (@mentions)

Mention users in comments and descriptions using `@Display Name` syntax. The CLI will automatically look up the user's JIRA account ID and convert it to a proper mention.

**Examples:**
```markdown
@Naveen R please review this issue.
CC: @Bhargav RK @Jyothi N
```

**Notes:**
- Use the user's **display name** as shown in JIRA (e.g., "Naveen R", not their email)
- Names with spaces are supported (e.g., `@Dhruva Kumar KR`)
- If the user is not found, the @mention will remain as plain text

#### Table Syntax

Tables use standard markdown pipe syntax with a header row, separator row, and data rows:

```markdown
| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Cell 1   | Cell 2   | Cell 3   |
| Cell 4   | Cell 5   | Cell 6   |
```

Tables support inline formatting (bold, italic, code) within cells.

### Issue Link Types

This JIRA instance uses numbered prefixes for link type names. Do NOT use standard names like "Relates", "Blocks", etc.

| Link Type | Name | Inward | Outward |
|-----------|------|--------|---------|
| Relates | `1. Relates` | relates to | relates to |
| Issue split | `2. Issue split` | split from | split to |
| Blocks | `3. Blocks` | is blocked by | blocks |
| Duplicate | `4. Duplicate` | is duplicated by | duplicates |
| Succeeding | `5. Succeeding` | preceeds | succeeds |

Use the link type **Name** column when creating links via `jira links add --type "..."`.

### `jira links list <issue-key>`
List all links on an issue.

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output as JSON |

### `jira links add <inward-issue> <outward-issue>`
Create a link between two issues.

| Flag | Short | Description |
|------|-------|-------------|
| `--type` | `-t` | Link type name (required, e.g., "3. Blocks") |

### `jira links delete <link-id>`
Delete an issue link.

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation |

### `jira links types`
List all available link types.

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output as JSON |

## Common JQL Queries

```jql
# Issues assigned to me
assignee = currentUser()

# Open bugs in a project
project = ED AND issuetype = Bug AND status != Done

# High priority issues
priority = "Prio 1" AND status IN ("To Do", "In Progress")

# Recently updated issues
project IN (ED, FW, DD) AND updated >= -7d ORDER BY updated DESC

# Unassigned tasks
project = ED AND assignee IS EMPTY

# Text search
text ~ "error" OR summary ~ "bug"

# Sprint-scoped
project = ED AND sprint IN openSprints()
```

## References

For detailed API documentation and JQL reference, see `references/api_reference.md`.
