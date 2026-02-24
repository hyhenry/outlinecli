# outlinecli

A command-line tool for managing notes and documents in [Outline](https://www.getoutline.com), built as an OpenClaw skill.

## Build & Install CLI

```bash
make build    # builds ./bin/outline
make install  # copies to /usr/local/bin/outline
```

Requires Go 1.21+.

## Install Skill in OpenClaw

```bash
mkdir ~/.openclaw/workspace/skills/outline/                 # create skill director
cp SKILL.md ~/.openclaw/workspace/skills/outline/SKILL.md   # copy skill.md
```

OpenClaw will automatically pick it up at session start, provided `outline` is in your `PATH` and `OUTLINE_API_KEY` is set.

## Authentication

Copy `sample.env` and fill in your credentials:

```bash
cp sample.env ~/my-outline.env
# open ~/my-outline.env and set OUTLINE_API_KEY (and OUTLINE_URL if self-hosted)
```

Then load them:

```bash
outline auth credentials ~/my-outline.env
```

Credentials are saved to `~/.config/outlinecli/credentials.json` (mode 0600).

```bash
outline auth status   # check what's configured
outline auth remove   # clear stored credentials
```

**Self-hosted Outline:** uncomment and set `OUTLINE_URL` in your env file:

```
OUTLINE_URL=https://your-outline-instance.com/api
```

Environment variables (`OUTLINE_API_KEY`, `OUTLINE_URL`) take precedence over stored credentials if set at runtime.

## Usage

### Collections

```bash
outline collection list [--json]
outline collection get <id> [--json]
```

### Documents

```bash
# Create
outline doc add --title "Title" [--body "markdown"] [--file path.md] [--collection <id>] [--draft]

# Read
outline doc get <id> [--json]

# Update
outline doc edit <id> [--title "New Title"] [--body "markdown"] [--file path.md]

# Delete (trash)
outline doc delete <id> [--permanent]

# List
outline doc list [--collection <id>] [--status draft|published|archived] [--json]

# Search
outline doc search "query" [--collection <id>] [--json]

# Archive / restore
outline doc archive <id>
outline doc restore <id>
```



