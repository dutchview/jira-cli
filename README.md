# jira-cli

A command-line interface for Jira Cloud, built in Go.

## Installation

### Homebrew

```bash
brew tap dutchview/tap
brew install jira
```

### From source

```bash
go install github.com/dutchview/jira-cli@latest
```

## Configuration

Create a `.env` file at `~/.config/jira/.env`:

```
JIRA_BASE_URL=https://yourcompany.atlassian.net
JIRA_EMAIL=you@example.com
JIRA_API_TOKEN=your_api_token
```

Get your API token at: https://id.atlassian.com/manage-profile/security/api-tokens

Config is loaded from (in order):
1. Environment variables
2. `.env` in current directory
3. `~/.config/jira/.env`
4. Custom file via `--config` flag

Run `jira configure` to see setup help.

## Usage

### Issues

```bash
# Search issues
jira issues search "project = ED ORDER BY updated DESC"
jira issues search -p ED -s "In Progress"
jira issues search --my-issues

# Get issue details
jira issues get ED-123
jira issues get ED-123 --comments
jira issues get ED-123 --json

# Create issue
jira issues create -p ED -t Task -s "Fix the login bug"
jira issues create -p ED -t Bug -s "Crash on save" -d "Steps to reproduce: ..."
jira issues create -p ED -t Task -s "Deadline task" --due-date 2026-03-15

# Update issue
jira issues update ED-123 -s "Updated title"
jira issues update ED-123 -d "New description with **markdown**"
jira issues update ED-123 -a <account-id>
jira issues update ED-123 --unassign
jira issues update ED-123 --due-date 2026-04-01
jira issues update ED-123 --no-due-date

# Delete issue
jira issues delete ED-123
jira issues delete ED-123 --force

# Transition issue
jira issues transition ED-123 "In Progress"
jira issues transition ED-123 "Done"
jira issues transition ED-123 --list
```

### Comments

```bash
jira comments list ED-123
jira comments add ED-123 "This is a **markdown** comment"
jira comments add ED-123 --file comment.md
jira comments update ED-123 <comment-id> "Updated text"
jira comments delete ED-123 <comment-id>
```

### Attachments

```bash
jira attachments add ED-123 ./screenshot.png
jira attachments add ED-123 ./report.pdf --filename "Q4 Report.pdf"
```

### Projects

```bash
jira projects list
jira projects get ED
```

### Users

```bash
jira users me
jira users search "john"
jira users assignable -p ED
```

## Markdown-to-ADF

Descriptions and comments support Markdown, which is automatically converted to Atlassian Document Format (ADF). Supported syntax:

- **Bold** (`**text**`)
- *Italic* (`*text*`)
- `Inline code` (`` `code` ``)
- [Links](url) (`[text](url)`)
- @mentions (`@Display Name` — resolved to JIRA users automatically)
- Headings (`# H1` through `###### H6`)
- Bullet lists (`- item`)
- Numbered lists (`1. item`)
- Code blocks (triple backticks with optional language)
- Tables (pipe syntax)
- Horizontal rules (`---`)

### Inline Images

To embed attached images inline in descriptions or comments, first attach the file, then reference it using `!filename!` syntax:

```bash
# 1. Attach the image
jira attachments add ED-123 screenshot.png

# 2. Reference it in description or comment
jira issues update ED-123 --description "## Screenshot\n\n!screenshot.png!"

# With width option
jira issues update ED-123 --description "!screenshot.png|width=720!"
```

When inline images are detected, the CLI automatically uses JIRA's wiki markup renderer (API v2) which supports embedded attachments.

### @Mentions

Use `@Display Name` in descriptions and comments to mention JIRA users. The CLI resolves display names to JIRA account IDs automatically via user search. Mentions are rendered as proper JIRA mention nodes in the Atlassian Document Format.

```bash
# Mention a user in a comment
jira comments add ED-123 "Hey @John Smith, can you review this?"

# Mention in a description
jira issues create -p ED -t Task -s "Review needed" -d "Assigned to @Jane Doe for review"
```

## Claude Code Skill

This repository includes a [Claude Code](https://claude.com/claude-code) skill that enables Claude to interact with JIRA directly from your terminal. The skill provides full issue management, commenting, attachments, and JQL search capabilities.

### Installing the Skill

1. Clone this repository (or ensure you have a local copy):
   ```bash
   git clone https://github.com/dutchview/jira-cli.git
   ```

2. Create a symlink from the skill directory to your Claude Code skills folder:
   ```bash
   mkdir -p ~/.claude/skills
   ln -s /path/to/jira-cli/claude-skill ~/.claude/skills/jira
   ```

3. Restart Claude Code. The skill will be automatically detected and available when you ask Claude to work with JIRA issues.

### What the Skill Provides

- Search, create, update, and delete JIRA issues
- Manage comments and attachments
- Transition issue statuses
- Full JQL query support
- Markdown-to-ADF conversion for descriptions and comments
- User mentions with automatic account ID resolution

The skill requires the `jira` CLI to be installed (see [Installation](#installation) above) and configured with your JIRA credentials.

## License

MIT
