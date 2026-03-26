# JIRA Cloud REST API Reference

This document provides reference information for the JIRA Cloud REST API v3.

## Authentication

JIRA Cloud uses Basic Authentication with:
- **Email**: Your Atlassian account email
- **API Token**: Generated at https://id.atlassian.com/manage-profile/security/api-tokens

The Authorization header format:
```
Authorization: Basic base64(email:api_token)
```

## Base URL

```
https://{your-domain}.atlassian.net
```

## Common Endpoints

### Issues

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/rest/api/3/issue/{issueIdOrKey}` | Get issue details |
| POST | `/rest/api/3/issue` | Create an issue |
| PUT | `/rest/api/3/issue/{issueIdOrKey}` | Update an issue |
| DELETE | `/rest/api/3/issue/{issueIdOrKey}` | Delete an issue |
| POST | `/rest/api/3/search` | Search issues using JQL |
| GET | `/rest/api/3/issue/{issueIdOrKey}/transitions` | Get available transitions |
| POST | `/rest/api/3/issue/{issueIdOrKey}/transitions` | Transition an issue |

### Issue Links

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/rest/api/3/issueLinkType` | List available link types |
| GET | `/rest/api/3/issue/{issueIdOrKey}?fields=issuelinks` | Get links on an issue |
| POST | `/rest/api/3/issueLink` | Create an issue link |
| DELETE | `/rest/api/3/issueLink/{linkId}` | Delete an issue link |

### Comments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/rest/api/3/issue/{issueIdOrKey}/comment` | Get comments |
| POST | `/rest/api/3/issue/{issueIdOrKey}/comment` | Add a comment |
| PUT | `/rest/api/3/issue/{issueIdOrKey}/comment/{id}` | Update a comment |
| DELETE | `/rest/api/3/issue/{issueIdOrKey}/comment/{id}` | Delete a comment |

### Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/rest/api/3/project/search` | List projects |
| GET | `/rest/api/3/project/{projectIdOrKey}` | Get project details |
| GET | `/rest/api/3/project/{projectIdOrKey}/statuses` | Get project statuses |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/rest/api/3/myself` | Get current user |
| GET | `/rest/api/3/user/search` | Search users |
| GET | `/rest/api/3/user/assignable/search` | Get assignable users |

### Metadata

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/rest/api/3/issuetype` | List issue types |
| GET | `/rest/api/3/priority` | List priorities |
| GET | `/rest/api/3/status` | List statuses |

## JQL (JIRA Query Language)

JQL is used to search for issues. Common operators and fields:

### Fields
- `project` - Project key or name
- `status` - Issue status
- `assignee` - Assigned user (use `currentUser()` for yourself)
- `reporter` - Reporter user
- `priority` - Priority level
- `issuetype` - Issue type (Bug, Task, Story, etc.)
- `created` - Creation date
- `updated` - Last updated date
- `resolution` - Resolution status
- `labels` - Issue labels
- `text` - Full-text search across summary and description

### Operators
- `=`, `!=` - Equality
- `~`, `!~` - Contains/not contains (for text)
- `>`, `>=`, `<`, `<=` - Comparison
- `IN`, `NOT IN` - List membership
- `IS`, `IS NOT` - Null checks (e.g., `IS EMPTY`)
- `WAS`, `WAS NOT` - Historical values
- `CHANGED` - Field change detection

### Functions
- `currentUser()` - The logged-in user
- `now()` - Current date/time
- `startOfDay()`, `endOfDay()` - Day boundaries
- `startOfWeek()`, `endOfWeek()` - Week boundaries
- `startOfMonth()`, `endOfMonth()` - Month boundaries

### Example Queries

```jql
# Issues assigned to me
assignee = currentUser()

# Open bugs in PROJECT
project = PROJECT AND issuetype = Bug AND status != Done

# High priority issues updated this week
priority = High AND updated >= startOfWeek()

# Unassigned tasks
project = PROJECT AND issuetype = Task AND assignee IS EMPTY

# Issues containing "error" in summary or description
text ~ "error"

# Issues created in the last 7 days
created >= -7d

# Issues with specific labels
labels IN ("frontend", "urgent")

# Complex query
project = PROJECT AND status IN ("To Do", "In Progress") AND
(priority = High OR labels = urgent) ORDER BY created DESC
```

## Atlassian Document Format (ADF)

JIRA uses ADF for rich text fields like description and comments.

### Basic Text
```json
{
  "type": "doc",
  "version": 1,
  "content": [
    {
      "type": "paragraph",
      "content": [
        {"type": "text", "text": "Hello World"}
      ]
    }
  ]
}
```

### Text with Formatting
```json
{
  "type": "doc",
  "version": 1,
  "content": [
    {
      "type": "paragraph",
      "content": [
        {"type": "text", "text": "Bold text", "marks": [{"type": "strong"}]},
        {"type": "text", "text": " and "},
        {"type": "text", "text": "italic text", "marks": [{"type": "em"}]}
      ]
    }
  ]
}
```

### Code Block
```json
{
  "type": "doc",
  "version": 1,
  "content": [
    {
      "type": "codeBlock",
      "attrs": {"language": "python"},
      "content": [
        {"type": "text", "text": "print('Hello')"}
      ]
    }
  ]
}
```

### Bullet List
```json
{
  "type": "doc",
  "version": 1,
  "content": [
    {
      "type": "bulletList",
      "content": [
        {
          "type": "listItem",
          "content": [
            {"type": "paragraph", "content": [{"type": "text", "text": "Item 1"}]}
          ]
        },
        {
          "type": "listItem",
          "content": [
            {"type": "paragraph", "content": [{"type": "text", "text": "Item 2"}]}
          ]
        }
      ]
    }
  ]
}
```

## Issue Types

Common issue types (may vary by project configuration):
- **Bug** - A problem or error
- **Task** - A work item
- **Story** - A user story
- **Epic** - A large body of work
- **Sub-task** - A child task of another issue

## Priorities

Common priorities (may vary by instance configuration):
- Highest
- High
- Medium
- Low
- Lowest

## Error Handling

JIRA API returns errors in JSON format:
```json
{
  "errorMessages": ["Issue does not exist or you do not have permission to see it."],
  "errors": {}
}
```

Or for field-level errors:
```json
{
  "errorMessages": [],
  "errors": {
    "summary": "You must specify a summary of the issue."
  }
}
```

## Rate Limiting

JIRA Cloud has rate limits. When exceeded, API returns HTTP 429 with:
- `Retry-After` header indicating seconds to wait
- `X-RateLimit-*` headers with limit information

## Resources

- [JIRA Cloud REST API Documentation](https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/)
- [JQL Reference](https://support.atlassian.com/jira-software-cloud/docs/use-advanced-search-with-jira-query-language-jql/)
- [Atlassian Document Format](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/)
