---
name: outline
description: >
  Manage notes and documents in Outline knowledge base — create, read, edit, delete, list, search, archive, and restore documents across collections
bins: ["outline"]
env: ["OUTLINE_API_KEY"]
---
# Outline CLI
Use the `outline` binary to manage documents in Outline. All commands accept `--json` for machine-readable output. Always use `--json` when you need to extract IDs or structured data.

## Setup
### First-time authentication
Copy `sample.env`, fill in your API key, then load it:
```
outline auth credentials ~/path/to/api-key.env
```
Credentials are stored at `~/.config/outlinecli/credentials.json` (mode 0600). Check status or remove at any time:
```
outline auth status
outline auth remove
```
The `OUTLINE_API_KEY` and `OUTLINE_URL` environment variables take precedence over stored credentials if set.

## Collections
Before creating documents, get a collection ID:
```
outline collection list --json
outline collection get <id> --json
```

## Documents
### Create a note
```
outline doc add --title "My Note" --body "# Hello
Content here" --collection <collectionId>
outline doc add --title "My Note" --file ./note.md --collection <collectionId>
outline doc add --title "Draft" --body "WIP" --collection <collectionId> --draft
```

### Read a note
```
outline doc get <id> --json
```

### Edit a note
```
outline doc edit <id> --title "New Title"
outline doc edit <id> --body "Updated content"
outline doc edit <id> --file ./updated.md
```

### Delete a note
```
outline doc delete <id> # moves to trash (recoverable for 30 days)
outline doc delete <id> --permanent # irreversible
```

### List notes
```
outline doc list --json
outline doc list --collection <collectionId> --json
outline doc list --status draft --json
outline doc list --status archived --json
```

### Search notes
```
outline doc search "keyword" --json
outline doc search "keyword" --collection <collectionId> --json
```

### Archive / restore
```
outline doc archive <id>
outline doc restore <id>
```

## Common workflows
**Find a document and edit it:**
1. `outline doc search "topic" --json` → get the `id` from results
2. `outline doc edit <id> --body "New content"`
**Create a note in the right collection:**
1. `outline doc list --json` → identify relevant collectionId from existing docs, or
2. `outline collection list --json` → pick collection by name
3. `outline doc add --title "..." --body "..." --collection <collectionId>`
**Clean up drafts:**
1. `outline doc list --status draft --json` → get IDs
2. `outline doc delete <id>` for each unwanted draft

## Output format
All `--json` responses follow the Outline API structure with a `data` field containing document or collection objects. Key fields:
- `id` — UUID used in all commands
- `title` — document title
- `text` — markdown content
- `collectionId` — parent collection UUID
- `updatedAt`, `createdAt` — ISO 8601 timestamps
- `archivedAt`, `deletedAt` — set when archived/deleted