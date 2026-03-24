# mattermost-cli

CLI for Mattermost team chat interaction.

## Install

Download a binary from the [latest release](https://github.com/jrogala/mattermost-cli/releases/latest), or install with Go:

```bash
go install github.com/jrogala/mattermost-cli@latest
```

## Setup

Set env vars or use login command:

```bash
export MATTERMOST_URL=https://mattermost.example.com
export MATTERMOST_TOKEN=your-access-token
```

Or authenticate interactively:

```bash
mattermost login
```

## Commands

| Command | Description |
|---|---|
| `login` | Authenticate with Mattermost server |
| `me` | Show current user info |
| `latest` | Show latest messages across channels |
| `unread` | Show unread messages |
| `mentions` | Show recent mentions |
| `channel find` | Find a channel by name |
| `channel list` | List joined channels |
| `channel read` | Read messages from a channel |
| `channel send` | Send a message to a channel |

## Examples

```bash
# Check unread messages
mattermost unread

# Read the last 20 messages from a channel
mattermost channel read --channel town-square --limit 20

# Send a message to a channel
mattermost channel send --channel dev-team "Deployment complete"

# List recent mentions
mattermost mentions
```

## JSON Output

All commands support `--json` for machine-readable output.
