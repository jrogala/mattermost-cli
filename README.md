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
$ mattermost unread
--- general ---
  Mar 24 10:15  john   Meeting at 2pm in the conference room
  Mar 24 10:10  sarah  Can anyone review the new design mockup?
--- dev-team ---
  Mar 24 14:22  alice  Deployed v1.2.3 to production

$ mattermost channel list
ID            TYPE     NAME
5dkc57v8yj..  public   general
6dkc57v8yj..  private  dev-team
7dkc57v8yj..  dm       @john

$ mattermost channel send --channel dev-team "Deployment complete"
Message sent to dev-team
```

## JSON Output

All commands support `--json` for machine-readable output.
