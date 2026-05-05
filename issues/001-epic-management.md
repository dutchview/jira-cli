# Parent (epic) management

## Problem

The CLI has no support for the parent / Epic Link relationship. `cmd/links.go` only
handles issue links (Relates, Blocks, etc.). Concrete pain hit on 2026-05-05:

- Wanted to attach ED-12104 / ED-12105 ("switch contracts offline") as **children**
  of epic ED-7281 (SecurED). Had to fall back to a `1. Relates` link, which is not
  the same as a parent/epic relationship — those issues won't appear in the epic's
  child panel.
- Wanted to list children of an epic (incl. Done items the JIRA UI hides by default)
  to take stock of work on ED-7281. No CLI command exists.
- After setting a parent, no way to verify from the CLI — `issues get` doesn't show
  the parent field today.

## Goals

1. Set, change, or clear an issue's parent from the CLI, in bulk.
2. Create an issue already attached to a parent in one shot.
3. List all children of a given parent.
4. See an issue's parent on `issues get`.

Non-goals: managing the "Epic Name" field, converting an issue type, bulk operations
beyond the parent field, project-type detection, fallback to the legacy
`customfield_10014` Epic Link field.

## Naming: parent, not epic

JIRA Cloud unifies the old "Epic Link" onto the standard `parent` field
(`fields.parent.key`). The same field also expresses Story → Subtask in modern
projects. This CLI uses **`parent`** consistently across read and write commands so
the surface stays honest if subtasks come up later. Help text on every parent-related
command and flag explicitly notes: *"In JIRA's UI, the parent of a Story under an
Epic is shown as the 'Epic'. Same field — JIRA Cloud unifies them."*

## API approach

- Write: PUT `/rest/api/3/issue/<KEY>` with `{"fields":{"parent":{"key":"<PARENT>"}}}`
  to set, `{"fields":{"parent":null}}` to clear.
- Create: POST `/rest/api/3/issue` with `parent.key` included in `fields`.
- Read children: JQL `parent = <KEY>` via the existing search endpoint.
- Read parent of an issue: returned in the issue payload's `fields.parent`.

We do **not** write `customfield_10014`. If a project rejects `parent` we surface
JIRA's error verbatim — the user can fix it via the UI.

## CLI surface

### Write — new `issues parent` group

```bash
jira issues parent set ED-12104 ED-12105 --to ED-7281    # attach one or many to ED-7281
jira issues parent clear ED-12104 ED-12105               # detach one or many
```

- Children are positional and variadic; the parent (only on `set`) is a `--to` flag.
  This keeps roles explicit (no chance of swapping parent and child positions) and
  makes `set` and `clear` symmetric.
- `--to` is required on `set`.
- Bulk semantics: each child is a separate API call. **Continue on error**, print a
  per-key result, exit non-zero if any failed. Example:
  ```
  Set parent of ED-12104 to ED-7281.
  Set parent of ED-12105 to ED-7281.
  Failed on ED-12106: <API error message>
  ```
- Setting a parent on a child that already has a different parent overwrites it
  silently (no extra GET, no warning) — this is the "move" use case.

### Write — `--parent` on `issues create`

```bash
jira issues create --project ED --type Story --summary "..." --parent ED-7281
```

`--parent` is optional; when present it adds `parent.key` to the create payload.

### Read — `--parent` on `issues search`

```bash
jira issues search --parent ED-7281            # children, including Done
jira issues search --parent ED-7281 --json
jira issues search 'parent = ED-7281 AND statusCategory != Done'   # raw JQL escape hatch
```

`--parent <KEY>` appends `parent = <KEY>` to the JQL builder. No `--no-done` flag —
the raw-JQL positional argument already covers status-category filtering, and the
common case here (cleanup) wants Done included.

### Read — Parent on `issues get`

`issues get <KEY>` gains a `Parent: <KEY>` line, rendered conditionally like Priority
and Assignee (only when present). The JSON output includes `fields.parent`
unchanged from the API response.

## Implementation plan (TDD)

Tests live next to code: `internal/api/client_test.go`, `cmd/issues_test.go`. The
HTTP client is exercised against `httptest.Server` — write the test first, watch it
fail, then implement.

1. **API: `Parent` on `IssueFields`** — extend the struct with `Parent *Issue` (or a
   trimmed struct exposing at least `Key`). Test asserts `GetIssue` round-trips a
   parent payload.
2. **API: parent on update** — assert that `UpdateIssue` with
   `fields["parent"] = map[string]string{"key": "ED-7281"}` marshals to
   `{"fields":{"parent":{"key":"ED-7281"}}}`, and that
   `fields["parent"] = nil` marshals to `{"fields":{"parent":null}}`. No new client
   method needed — `UpdateIssue` is generic.
3. **API: parent on create** — extend `CreateIssue` and `CreateIssueWiki` with a
   `parentKey string` parameter (positional, consistent with the existing signature;
   no options-struct refactor). Test asserts the request body contains `parent.key`
   when set, and omits it when empty.
4. **CLI: `issues parent set` / `clear`** — new `IssuesParentCmd` with `Set` and
   `Clear` subcommands, mounted under `IssuesCmd`. Tests cover:
   - flag wiring (`--to` required on `set`, rejected on `clear`),
   - the field map sent to the fake client per child,
   - bulk continue-on-error and non-zero exit when any child fails.
5. **CLI: `--parent` on `issues create`** — flag wiring; test asserts `parentKey`
   reaches `CreateIssue` / `CreateIssueWiki`.
6. **CLI: `--parent` on `issues search`** — flag appends `parent = <KEY>` to the
   JQL builder. Test asserts the final JQL string for combinations
   (`--parent` alone, `--parent` + `--status`, etc.).
7. **CLI: `Parent` on `issues get`** — render `Parent: <KEY>` when present, skip
   when nil. Test asserts the formatted output for both cases.
8. **Manual smoke test** against ED-7281: set ED-12105's parent, list children with
   `issues search --parent ED-7281`, verify with `issues get ED-12105`, clear the
   parent again to leave the ticket in its prior state.

## Out of scope / deferred

- **`customfield_10014` fallback**: not implemented. If a Dutchview project rejects
  the unified `parent` field in practice, revisit.
- **Project-type detection**: assumed unnecessary; ED is company-managed but uses
  next-gen schema.
- **Generic subtask vocabulary in docs**: the help text mentions parent ↔ epic
  equivalence; a similar note for Story → Subtask is not added until that workflow
  becomes a real pain.
- **`--no-done` convenience flag**: dropped in favor of raw JQL.
- **Top-level `epics` command group**: not created. If a future need arises that
  doesn't fit under `issues`, revisit.
