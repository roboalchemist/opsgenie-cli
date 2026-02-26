# opsgenie-cli

Go CLI for the OpsGenie REST API v2. Manage alerts, incidents, teams, schedules, on-call rotations, heartbeats, escalations, integrations, and more from your terminal.

## Installation

### Homebrew (recommended)

```bash
brew tap roboalchemist/tap
brew install opsgenie-cli
```

### Go Install

```bash
go install github.com/roboalchemist/opsgenie-cli@latest
```

### Build from Source

```bash
git clone https://github.com/roboalchemist/opsgenie-cli.git
cd opsgenie-cli
make build
```

## Quick Start

```bash
# Set your API key
export OPSGENIE_API_KEY="your-api-key-here"

# Verify connectivity
opsgenie-cli account get

# List open alerts
opsgenie-cli alerts list --query "status:open"

# Check who is on-call
opsgenie-cli on-call get --schedule "Primary On-Call"
```

## Authentication

The API key is resolved in priority order:

1. `OPSGENIE_API_KEY` environment variable
2. `~/.opsgenie-cli-auth.json` — `{"api_key": "your-key"}`

To write the config file directly:

```bash
echo '{"api_key":"your-api-key"}' > ~/.opsgenie-cli-auth.json
chmod 600 ~/.opsgenie-cli-auth.json
```

Get your API key from OpsGenie: Settings → API key management → Add new API key.

## Available Commands

| Command | Subcommands | Description |
|---------|-------------|-------------|
| `account` | `get` | Account information |
| `alerts` | `list`, `get`, `create`, `delete`, `acknowledge`, `close`, `snooze`, `escalate`, `assign`, `add-note`, `add-tags`, `remove-tags`, `count` | Alert management |
| `contacts` | `list`, `get`, `create`, `update`, `delete`, `enable` | User contact methods |
| `custom-roles` | `list`, `get`, `create`, `update`, `delete` | Custom role management |
| `deployments` | `list`, `get`, `create`, `update`, `search` | Deployment tracking |
| `escalations` | `list`, `get`, `create`, `update`, `delete` | Escalation policies |
| `forwarding-rules` | `list`, `get`, `create`, `update`, `delete` | Notification forwarding |
| `heartbeats` | `list`, `get`, `create`, `update`, `delete`, `enable`, `disable`, `ping` | Heartbeat monitors |
| `incidents` | `list`, `get`, `create`, `close`, `resolve`, `reopen`, `delete`, `add-note`, `add-tags` | Incident management |
| `integrations` | `list`, `get`, `create`, `update`, `delete`, `enable`, `disable` | Integrations |
| `maintenance` | `list`, `get`, `create`, `update`, `delete`, `cancel` | Maintenance windows |
| `notification-rules` | `list`, `get`, `create`, `update`, `delete`, `enable` | Notification rules |
| `on-call` | `get`, `next` | On-call schedule queries |
| `policies` | `list`, `get`, `create`, `update`, `delete`, `enable`, `disable` | Alert/notification policies |
| `postmortems` | `get`, `create`, `update`, `delete` | Postmortem management |
| `schedule-overrides` | `list`, `get`, `create`, `update`, `delete` | Schedule overrides |
| `schedule-rotations` | `list`, `get`, `create`, `update`, `delete` | Schedule rotations |
| `schedules` | `list`, `get`, `create`, `update`, `delete` | On-call schedules |
| `services` | `list`, `get`, `create`, `update`, `delete` | Service catalog |
| `team-members` | `add`, `remove` | Team membership |
| `team-routing-rules` | `list`, `get`, `create`, `update`, `delete` | Team routing rules |
| `teams` | `list`, `get`, `create`, `update`, `delete` | Team management |
| `users` | `list`, `get`, `create`, `update`, `delete` | User management |

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | JSON output (best for scripting/agents) |
| `--plaintext` | `-p` | Tab-separated output for piping |
| `--no-color` | | Disable colored output |
| `--debug` | | Verbose logging to stderr |
| `--region` | | OpsGenie region: `us` (default) or `eu` |
| `--fields` | | Comma-separated fields to display (JSON mode) |
| `--jq` | | JQ expression to filter JSON output |

## EU Region Support

For EU-hosted OpsGenie accounts, pass `--region eu`:

```bash
opsgenie-cli --region eu alerts list
```

Or set `OPSGENIE_API_URL` to override the base URL entirely.

## Examples

```bash
# List all open P1 alerts as JSON
opsgenie-cli alerts list --query "status:open AND priority:P1" --json

# Acknowledge an alert
opsgenie-cli alerts acknowledge <alert-id>

# Close an alert with a note
opsgenie-cli alerts close <alert-id> --note "Fixed by reverting deploy abc"

# Create an alert
opsgenie-cli alerts create --message "Disk usage > 90%" --priority P2 --responders "team:infra"

# Check who is on-call right now
opsgenie-cli on-call get --schedule "Primary On-Call"

# Check who is on-call next
opsgenie-cli on-call next --schedule "Primary On-Call" --json

# Create a heartbeat monitor
opsgenie-cli heartbeats create --name "payments-cron" --interval 10 --interval-unit minutes

# Ping a heartbeat
opsgenie-cli heartbeats ping payments-cron

# List teams
opsgenie-cli teams list

# Add a user to a team
opsgenie-cli team-members add --team platform --user alice@example.com

# Create a maintenance window
opsgenie-cli maintenance create \
  --description "Scheduled DB maintenance" \
  --start-date "2024-01-15T02:00:00Z" \
  --end-date "2024-01-15T04:00:00Z"

# List open incidents
opsgenie-cli incidents list --query "status:open"

# Resolve an incident
opsgenie-cli incidents resolve <incident-id> --note "Root cause addressed"
```

## Shell Completion

```bash
# Bash
opsgenie-cli completion bash >> ~/.bashrc

# Zsh
opsgenie-cli completion zsh >> ~/.zshrc

# Fish
opsgenie-cli completion fish > ~/.config/fish/completions/opsgenie-cli.fish
```

## Development

```bash
make build          # Build the binary
make test           # Run smoke tests (no API key needed)
make test-unit      # Run unit tests with coverage
make test-integration  # Run integration tests (requires OPSGENIE_API_KEY)
make check          # fmt + lint + test + test-unit
make fmt            # Format code
make lint           # Lint with golangci-lint
make install        # Install to /usr/local/bin/
make man            # Generate man pages
make docs-gen       # Generate markdown CLI reference
make help           # Show all targets
```

## License

MIT
