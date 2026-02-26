---
name: opsgenie-cli
description: CLI for the OpsGenie REST API v2. Use when managing alerts, incidents, teams, schedules, on-call, heartbeats, escalations, integrations, and more via OpsGenie.
scope: both
---

# opsgenie-cli

CLI for the OpsGenie REST API v2. Use when managing alerts, incidents, teams, schedules, on-call, heartbeats, escalations, integrations, and more via OpsGenie.

**Binary**: `opsgenie-cli` | **Auth**: `OPSGENIE_API_KEY` env or `~/.opsgenie-cli-auth.json`

## Setup

```bash
export OPSGENIE_API_KEY="your-api-key"   # Set API key
opsgenie-cli account get                 # Verify connectivity
```

Or write to config file:
```bash
echo '{"api_key":"your-api-key"}' > ~/.opsgenie-cli-auth.json
chmod 600 ~/.opsgenie-cli-auth.json
```

## Output Formats

| Flag | Format | Use case |
|------|--------|----------|
| (none) | Colored table | Human terminal |
| `-p` / `--plaintext` | Tab-separated | Piping, agents |
| `-j` / `--json` | JSON | Programmatic parsing |

**Always use `--json` for programmatic parsing.**

## EU Region

```bash
opsgenie-cli --region eu alerts list
```

<examples>
<example>
Task: List open P1 alerts

```bash
opsgenie-cli alerts list --query "status:open AND priority:P1" --json
```

Output:
```json
[
  {"id": "abc123", "message": "Database CPU high", "status": "open", "priority": "P1", ...}
]
```
</example>

<example>
Task: Acknowledge an alert

```bash
opsgenie-cli alerts acknowledge abc123-alert-id
```

Output:
```
✓ Alert acknowledged
```
</example>

<example>
Task: Check who is on-call for a schedule

```bash
opsgenie-cli on-call get --schedule "Primary On-Call" --json
```

Output:
```json
{"scheduleRef": {"name": "Primary On-Call"}, "onCallParticipants": [{"name": "alice@example.com", ...}]}
```
</example>

<example>
Task: List all teams

```bash
opsgenie-cli teams list
```

Output:
```
NAME           ID
Platform       abc-123
Backend        def-456
```
</example>

<example>
Task: Create a heartbeat monitor

```bash
opsgenie-cli heartbeats create --name "my-service-heartbeat" --interval 5 --interval-unit minutes
```

Output:
```
✓ Heartbeat "my-service-heartbeat" created
```
</example>

<example>
Task: List open incidents

```bash
opsgenie-cli incidents list --query "status:open" --json
```

Output:
```json
[{"id": "inc-1", "message": "Payment service down", "status": "open", "priority": "P1", ...}]
```
</example>
</examples>

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | JSON output |
| `--plaintext` | `-p` | Tab-separated output for piping |
| `--no-color` | | Disable colored output |
| `--debug` | | Verbose logging to stderr |
| `--region` | | OpsGenie region: `us` (default) or `eu` |
| `--fields` | | Comma-separated fields to display (JSON mode) |
| `--jq` | | JQ expression to filter JSON output |

## Authentication

Priority order:
1. `OPSGENIE_API_KEY` environment variable
2. `~/.opsgenie-cli-auth.json` — `{"api_key": "..."}`

## Available Commands

| Command | Description |
|---------|-------------|
| `account` | Account information |
| `alerts` | Alert management (list, get, create, delete, acknowledge, close, snooze, escalate, assign, add-note, add-tags, remove-tags, count) |
| `contacts` | User contact methods |
| `custom-roles` | Custom role management |
| `deployments` | Deployment tracking |
| `escalations` | Escalation policies |
| `forwarding-rules` | Notification forwarding rules |
| `heartbeats` | Heartbeat monitors |
| `incidents` | Incident management (list, get, create, close, resolve, reopen, delete, add-note, add-tags) |
| `integrations` | Integrations (list, get, create, update, delete, enable, disable) |
| `maintenance` | Maintenance windows |
| `notification-rules` | User notification rules |
| `on-call` | On-call schedule queries |
| `policies` | Alert and notification policies |
| `postmortems` | Postmortem management |
| `schedule-overrides` | Schedule overrides |
| `schedule-rotations` | Schedule rotations |
| `schedules` | On-call schedules |
| `services` | Service catalog |
| `team-members` | Team membership |
| `team-routing-rules` | Team routing rules |
| `teams` | Team management |
| `users` | User management |

See [reference/commands.md](reference/commands.md) for full command reference with all flags.
