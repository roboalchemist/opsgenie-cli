---
name: opsgenie-cli
description: CLI for the OpsGenie REST API v2. Use when managing alerts, incidents, teams, schedules, on-call, heartbeats, escalations, integrations, and more via OpsGenie.
scope: both
---

# opsgenie-cli

CLI for the OpsGenie REST API v2. Use when managing alerts, incidents, teams, schedules, on-call, heartbeats, escalations, integrations, and more via OpsGenie.

**Binary**: `opsgenie-cli` | **Auth**: `OPSGENIE_API_KEY` env or `~/.opsgenie-cli-auth.json`

<examples>
<example>
Task: List open P1 alerts as JSON

```bash
opsgenie-cli alerts list --query "status:open AND priority:P1" --limit 5 --json
```

Output:
```json
[
  {
    "id": "bff3ccbf-c7dd-4d96-8...",
    "message": "[Datadog] [P1] [Warn] Service vnet has a high error rate on env:prod2",
    "status": "open",
    "priority": "P1",
    "acknowledged": true,
    "createdAt": "2026-02-26T03:31:05.985Z"
  }
]
```
</example>

<example>
Task: List alerts as a table with specific fields

```bash
opsgenie-cli alerts list --limit 2 --json --fields id,message,priority
```

Output:
```json
[
  {
    "id": "9fefe690-4ad4-48e0-b50e-c184e9e39c0f-1772089369305",
    "message": "[Datadog] [P5] [Triggered] Index vnetsuite-services daily logging quota approaching",
    "priority": "P5"
  }
]
```
</example>

<example>
Task: Extract a single value with --jq

```bash
opsgenie-cli alerts list --limit 1 --jq '.[0].message'
```

Output:
```
"[Datadog] [P5] [Triggered] Index vnetsuite-services daily logging quota approaching"
```
</example>

<example>
Task: Get total alert count

```bash
opsgenie-cli alerts count --json
```

Output:
```json
{
  "count": 111830
}
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
{"onCallParticipants": [{"name": "alice@example.com", "type": "user"}]}
```
</example>
</examples>

## Setup

```bash
export OPSGENIE_API_KEY="your-api-key"   # Required
opsgenie-cli alerts list --limit 1       # Verify connectivity
```

Or config file: `echo '{"api_key":"..."}' > ~/.opsgenie-cli-auth.json && chmod 600 ~/.opsgenie-cli-auth.json`

## Output Formats

| Flag | Format | Use case |
|------|--------|----------|
| (none) | Colored table | Human terminal |
| `-p` / `--plaintext` | Tab-separated | Piping, scripts |
| `-j` / `--json` | JSON | Programmatic parsing |
| `--fields` | Filtered JSON | Reduce output to specific fields |
| `--jq` | JQ-filtered JSON | Complex filtering expressions |

**Always use `--json` for programmatic parsing. `--fields` and `--jq` implicitly enable JSON mode.**

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | JSON output |
| `--plaintext` | `-p` | Tab-separated output |
| `--no-color` | | Disable colored output |
| `--verbose` | `-v` | Verbose output |
| `--debug` | | Debug logging to stderr |
| `--quiet` | `-q` | Suppress progress output |
| `--silent` | | Synonym for `--quiet` |
| `--region` | | OpsGenie region: `us` (default) or `eu` |
| `--fields` | | Comma-separated fields (implicitly enables JSON) |
| `--jq` | | JQ expression (implicitly enables JSON) |

## Authentication

Priority: `OPSGENIE_API_KEY` env var → `~/.opsgenie-cli-auth.json`

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OPSGENIE_API_KEY` | API key for authentication (required) |
| `OPSGENIE_API_URL` | Override API base URL (default: `https://api.opsgenie.com`) |
| `NO_COLOR` | Disable colored output when set |

## Available Commands

| Command | Description |
|---------|-------------|
| `alerts` | list, get, create, delete, acknowledge, close, snooze, escalate, assign, add-note, add-tags, remove-tags, count |
| `incidents` | list, get, create, close, resolve, reopen, delete, add-note, add-tags (**uses /v1 API**) |
| `teams` | list, get, create, update, delete |
| `team-members` | add, remove |
| `team-routing-rules` | list, get, create, update, delete |
| `users` | list, get, create, update, delete |
| `contacts` | list, get, create, update, delete, enable, disable |
| `notification-rules` | list, get, create, update, delete, enable, disable |
| `schedules` | list, get, create, update, delete |
| `schedule-rotations` | list, get, create, update, delete |
| `schedule-overrides` | list, get, create, update, delete |
| `on-call` | get, next |
| `escalations` | list, get, create, update, delete |
| `heartbeats` | list, get, create, update, delete, enable, disable, ping |
| `integrations` | list, get, create, update, delete, enable, disable |
| `maintenance` | list, get, create, update, delete, cancel |
| `services` | list, get, create, update, delete |
| `policies` | list, get, create, update, delete, enable, disable |
| `forwarding-rules` | list, get, create, update, delete |
| `custom-roles` | list, get, create, update, delete |
| `postmortems` | get, create, update, delete |
| `deployments` | list, get, create, update, search |
| `account` | get |

See [reference/commands.md](reference/commands.md) for full command reference with all flags and options.
