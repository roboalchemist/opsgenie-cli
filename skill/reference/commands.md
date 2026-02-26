# opsgenie-cli Command Reference

## Table of Contents
- [Global Flags](#global-flags)
- [Alert Management](#alert-management)
- [Incident Management](#incident-management)
- [Team / User Management](#team--user-management)
- [Scheduling](#scheduling)
- [Operations](#operations)
- [Escalations & Policies](#escalations--policies)
- [Integrations & Routing](#integrations--routing)
- [Notifications](#notifications)
- [Incident Infrastructure](#incident-infrastructure)
- [Account & Utilities](#account--utilities)
- [API Behavior](#api-behavior)

Complete reference for all commands. Global flags apply to every command.

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--json` | `-j` | false | JSON output |
| `--plaintext` | `-p` | false | Tab-separated output for piping |
| `--no-color` | | false | Disable colored output |
| `--debug` | | false | Verbose logging to stderr |
| `--region` | | `us` | OpsGenie region (`us` or `eu`) |
| `--fields` | | | Comma-separated fields to display (JSON mode) |
| `--jq` | | | JQ expression to filter JSON output |
| `--silent` | | false | Synonym for `--quiet` |

---

## Alert Management

### `alerts list`

List alerts with optional filtering.

```bash
opsgenie-cli alerts list [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--query` | | Search query (OpsGenie query syntax) |
| `--limit` | 20 | Maximum number of alerts to return |
| `--offset` | 0 | Start offset for pagination |
| `--sort` | | Sort field (e.g. `createdAt`, `updatedAt`) |
| `--all` | false | Fetch all alerts (paginate through all pages) |

```bash
# List open P1 alerts
opsgenie-cli alerts list --query "status:open AND priority:P1"

# List all alerts, JSON output
opsgenie-cli alerts list --all --json

# List unacknowledged alerts
opsgenie-cli alerts list --query "status:open AND acknowledged:false"
```

### `alerts get <id>`

Get a single alert by ID.

```bash
opsgenie-cli alerts get <alert-id>
opsgenie-cli alerts get abc123 --json
```

### `alerts create`

Create a new alert.

```bash
opsgenie-cli alerts create [flags]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--message` | Yes | Alert message |
| `--description` | | Alert description |
| `--priority` | | Priority: `P1`–`P5` |
| `--tags` | | Comma-separated tags |
| `--responders` | | Comma-separated responders, e.g. `team:ops,user:alice@example.com` |

```bash
opsgenie-cli alerts create --message "High CPU" --priority P2 --responders "team:platform"
```

### `alerts delete <id>`

Delete an alert.

```bash
opsgenie-cli alerts delete <alert-id>
```

### `alerts acknowledge <id>`

Acknowledge an alert.

```bash
opsgenie-cli alerts acknowledge <alert-id>
```

### `alerts close <id>`

Close an alert.

| Flag | Description |
|------|-------------|
| `--note` | Note to add when closing |

```bash
opsgenie-cli alerts close <alert-id> --note "Resolved by deploy"
```

### `alerts snooze <id>`

Snooze an alert until a specified time.

| Flag | Required | Description |
|------|----------|-------------|
| `--end-time` | Yes | Snooze until this time (RFC3339, e.g. `2024-01-15T10:00:00Z`) |

```bash
opsgenie-cli alerts snooze <alert-id> --end-time "2024-01-15T10:00:00Z"
```

### `alerts escalate <id>`

Escalate an alert to an escalation policy.

| Flag | Required | Description |
|------|----------|-------------|
| `--escalation` | Yes | Escalation policy name |

```bash
opsgenie-cli alerts escalate <alert-id> --escalation "Critical Escalation"
```

### `alerts assign <id>`

Assign an alert to an owner.

| Flag | Required | Description |
|------|----------|-------------|
| `--owner` | Yes | Username of the new owner |

```bash
opsgenie-cli alerts assign <alert-id> --owner alice@example.com
```

### `alerts add-note <id>`

Add a note to an alert.

| Flag | Required | Description |
|------|----------|-------------|
| `--note` | Yes | Note text |

```bash
opsgenie-cli alerts add-note <alert-id> --note "Investigating disk usage"
```

### `alerts add-tags <id>`

Add tags to an alert.

| Flag | Required | Description |
|------|----------|-------------|
| `--tags` | Yes | Comma-separated tags to add |

```bash
opsgenie-cli alerts add-tags <alert-id> --tags "infra,prod"
```

### `alerts remove-tags <id>`

Remove tags from an alert.

| Flag | Required | Description |
|------|----------|-------------|
| `--tags` | Yes | Comma-separated tags to remove |

```bash
opsgenie-cli alerts remove-tags <alert-id> --tags "infra"
```

### `alerts count`

Count alerts matching a query.

| Flag | Description |
|------|-------------|
| `--query` | Search query to count matching alerts |

```bash
opsgenie-cli alerts count --query "status:open AND priority:P1"
```

---

## Incident Management

### `incidents list`

List incidents.

| Flag | Default | Description |
|------|---------|-------------|
| `--query` | | Search query |
| `--limit` | 0 (all) | Maximum number of incidents |
| `--offset` | 0 | Start offset |
| `--sort` | | Sort field |
| `--order` | | Sort order: `asc` or `desc` |

```bash
opsgenie-cli incidents list --query "status:open" --json
```

### `incidents get <id>`

Get an incident by ID.

```bash
opsgenie-cli incidents get <incident-id> --json
```

### `incidents create`

Create a new incident.

| Flag | Required | Description |
|------|----------|-------------|
| `--message` | Yes | Incident message |
| `--description` | | Incident description |
| `--priority` | | Priority: `P1`–`P5` |
| `--tags` | | Comma-separated tags |
| `--responders` | | Comma-separated responders, e.g. `team:ops,user:alice@example.com` |

```bash
opsgenie-cli incidents create --message "Payment outage" --priority P1 --responders "team:payments"
```

### `incidents close <id>`

Close an incident.

| Flag | Description |
|------|-------------|
| `--note` | Note to add when closing |

### `incidents resolve <id>`

Resolve an incident.

| Flag | Description |
|------|-------------|
| `--note` | Note to add when resolving |

### `incidents reopen <id>`

Reopen a closed or resolved incident.

| Flag | Description |
|------|-------------|
| `--note` | Note to add when reopening |

### `incidents delete <id>`

Delete an incident.

### `incidents add-note <id>`

Add a note to an incident.

| Flag | Required | Description |
|------|----------|-------------|
| `--note` | Yes | Note text |

### `incidents add-tags <id>`

Add tags to an incident.

| Flag | Required | Description |
|------|----------|-------------|
| `--tags` | Yes | Comma-separated tags |

---

## Team / User Management

### `teams list`

List all teams.

```bash
opsgenie-cli teams list --json
```

### `teams get <id>`

Get a team by ID or name.

```bash
opsgenie-cli teams get platform-team
```

### `teams create`

Create a new team.

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | Team name |
| `--description` | | Team description |

### `teams update <id>`

Update a team by ID or name.

| Flag | Description |
|------|-------------|
| `--name` | New team name |
| `--description` | New description |

### `teams delete <id>`

Delete a team by ID or name.

### `team-members add`

Add a member to a team.

| Flag | Required | Description |
|------|----------|-------------|
| `--team` | Yes | Team ID or name |
| `--user` | Yes | User ID or username |
| `--role` | | Member role |

### `team-members remove`

Remove a member from a team.

| Flag | Required | Description |
|------|----------|-------------|
| `--team` | Yes | Team ID or name |
| `--user` | Yes | User ID or username |

### `users list`

List all users. Fetches all users with automatic pagination. Supports `--fields` and `--jq` (plus global flags) for output filtering.

### `users get <id>`

Get a user by ID or username.

### `users create`

Create a new user.

| Flag | Required | Description |
|------|----------|-------------|
| `--username` | Yes | User email/username |
| `--full-name` | | Full display name |
| `--role` | | User role (default: "user") |

### `users update <id>`

Update a user by ID or username.

### `users delete <id>`

Delete a user by ID or username.

### `custom-roles list`

List all custom roles.

### `custom-roles get <id>`

Get a custom role by ID.

### `custom-roles create`

Create a new custom role.

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | Role name |
| `--rights` | | Comma-separated granted rights |

### `custom-roles update <id>`

Update a custom role.

### `custom-roles delete <id>`

Delete a custom role.

---

## Scheduling

### `schedules list`

List all on-call schedules.

```bash
opsgenie-cli schedules list
```

### `schedules get <id>`

Get a schedule by ID or name.

```bash
opsgenie-cli schedules get "Primary On-Call" --json
```

### `schedules create`

Create a new schedule.

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | Schedule name |
| `--timezone` | Yes | Timezone (e.g. `America/New_York`) |
| `--description` | | Schedule description |
| `--enabled` | | Whether schedule is enabled |

### `schedules update <id>`

Update a schedule by ID or name.

### `schedules delete <id>`

Delete a schedule by ID or name.

### `on-call get`

Get current on-call participants for a schedule.

| Flag | Required | Description |
|------|----------|-------------|
| `--schedule` | Yes | Schedule ID or name |
| `--flat` | | Return flat list of participants |

```bash
opsgenie-cli on-call get --schedule "Primary On-Call"
opsgenie-cli on-call get --schedule "Primary On-Call" --flat --json
```

### `on-call next`

Get next on-call participants for a schedule.

| Flag | Required | Description |
|------|----------|-------------|
| `--schedule` | Yes | Schedule ID or name |
| `--flat` | | Return flat list of participants |

```bash
opsgenie-cli on-call next --schedule "Primary On-Call"
```

### `schedule-rotations list`

List rotations for a schedule.

| Flag | Required | Description |
|------|----------|-------------|
| `--schedule` | Yes | Schedule ID or name |

### `schedule-rotations get`

Get a schedule rotation by ID.

| Flag | Required | Description |
|------|----------|-------------|
| `--schedule` | Yes | Schedule ID or name |
| `--id` | Yes | Rotation ID |

### `schedule-rotations create`

Create a rotation for a schedule.

### `schedule-rotations update`

Update a schedule rotation.

### `schedule-rotations delete`

Delete a schedule rotation.

### `schedule-overrides list`

List overrides for a schedule.

| Flag | Required | Description |
|------|----------|-------------|
| `--schedule` | Yes | Schedule ID or name |

### `schedule-overrides get`

Get a schedule override by alias.

### `schedule-overrides create`

Create an override for a schedule.

| Flag | Required | Description |
|------|----------|-------------|
| `--schedule` | Yes | Schedule ID or name |
| `--user` | Yes | User to override with |
| `--start-date` | Yes | Override start (RFC3339) |
| `--end-date` | Yes | Override end (RFC3339) |

### `schedule-overrides update`

Update a schedule override.

### `schedule-overrides delete`

Delete a schedule override by alias.

---

## Operations

### `heartbeats list`

List all heartbeat monitors.

```bash
opsgenie-cli heartbeats list
```

### `heartbeats get <name>`

Get a heartbeat by name.

### `heartbeats create`

Create a new heartbeat monitor.

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | Heartbeat name |
| `--description` | | Description |
| `--interval` | | Ping interval (default: 10) |
| `--interval-unit` | | Interval unit: `minutes`, `hours`, `days` (default: `minutes`) |
| `--enabled` | | Whether enabled (default: true) |

```bash
opsgenie-cli heartbeats create --name "my-service" --interval 5 --interval-unit minutes
```

### `heartbeats update <name>`

Update a heartbeat.

| Flag | Description |
|------|-------------|
| `--description` | New description |
| `--interval` | New ping interval |
| `--interval-unit` | New interval unit |
| `--enabled` | Enable or disable |

### `heartbeats delete <name>`

Delete a heartbeat.

### `heartbeats enable <name>`

Enable a heartbeat.

### `heartbeats disable <name>`

Disable a heartbeat.

### `heartbeats ping <name>`

Ping a heartbeat (reset the expiry timer).

```bash
opsgenie-cli heartbeats ping my-service
```

### `maintenance list`

List all maintenance windows.

| Flag | Description |
|------|-------------|
| `--type` | Filter: `past`, `present`, `future` |

### `maintenance get <id>`

Get a maintenance window by ID.

### `maintenance create`

Create a maintenance window.

| Flag | Required | Description |
|------|----------|-------------|
| `--description` | Yes | Description |
| `--start-date` | Yes | Start time (RFC3339) |
| `--end-date` | Yes | End time (RFC3339) |
| `--rules` | | Comma-separated entity identifiers to suppress |

### `maintenance update <id>`

Update a maintenance window.

### `maintenance delete <id>`

Delete a maintenance window.

### `maintenance cancel <id>`

Cancel an active maintenance window.

---

## Escalations & Policies

### `escalations list`

List all escalation policies.

### `escalations get <id>`

Get an escalation policy by ID or name.

### `escalations create`

Create an escalation policy.

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | Policy name |
| `--description` | | Description |
| `--rules` | | JSON array of escalation rules |

### `escalations update <id>`

Update an escalation policy by ID or name.

### `escalations delete <id>`

Delete an escalation policy by ID or name.

### `policies list`

List all alert/notification policies.

### `policies get <id>`

Get a policy by ID.

### `policies create`

Create a new policy.

### `policies update <id>`

Update a policy.

### `policies delete <id>`

Delete a policy.

### `policies enable <id>`

Enable a policy.

### `policies disable <id>`

Disable a policy.

---

## Integrations & Routing

### `integrations list`

List all integrations.

### `integrations get <id>`

Get an integration by ID.

### `integrations create`

Create a new integration.

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | Integration name |
| `--type` | Yes | Integration type (e.g. `Webhook`) |

### `integrations update <id>`

Update an integration.

### `integrations delete <id>`

Delete an integration.

### `integrations enable <id>`

Enable an integration.

### `integrations disable <id>`

Disable an integration.

### `team-routing-rules list`

List routing rules for a team.

| Flag | Required | Description |
|------|----------|-------------|
| `--team` | Yes | Team ID or name |

### `team-routing-rules get`

Get a routing rule by ID.

| Flag | Required | Description |
|------|----------|-------------|
| `--team` | Yes | Team ID or name |
| `--id` | Yes | Routing rule ID |

### `team-routing-rules create`

Create a routing rule for a team.

### `team-routing-rules update`

Update a routing rule.

### `team-routing-rules delete`

Delete a routing rule.

---

## Notifications

### `contacts list`

List contact methods for a user.

| Flag | Required | Description |
|------|----------|-------------|
| `--user` | Yes | User ID or username |

### `contacts get`

Get a contact by ID.

| Flag | Required | Description |
|------|----------|-------------|
| `--user` | Yes | User ID or username |
| `--id` | Yes | Contact ID |

### `contacts create`

Create a contact for a user.

| Flag | Required | Description |
|------|----------|-------------|
| `--user` | Yes | User ID or username |
| `--method` | Yes | Contact method type |
| `--to` | Yes | Contact address |

### `contacts update`

Update a contact.

### `contacts delete`

Delete a contact.

### `contacts enable`

Enable a contact.

### `contacts disable`

Disable a contact.

| Flag | Required | Description |
|------|----------|-------------|
| `--user` | Yes | User ID or username |
| `--contact-id` | Yes | Contact ID |

### `notification-rules list`

List notification rules for a user.

| Flag | Required | Description |
|------|----------|-------------|
| `--user` | Yes | User ID or username |

### `notification-rules get`

Get a notification rule.

### `notification-rules create`

Create a notification rule for a user.

### `notification-rules update`

Update a notification rule.

### `notification-rules delete`

Delete a notification rule.

### `notification-rules enable`

Enable a notification rule.

### `notification-rules disable`

Disable a notification rule.

| Flag | Required | Description |
|------|----------|-------------|
| `--user` | Yes | User ID or username |
| `--id` | Yes | Notification rule ID |

### `forwarding-rules list`

List all forwarding rules.

### `forwarding-rules get <id>`

Get a forwarding rule by ID.

### `forwarding-rules create`

Create a forwarding rule.

| Flag | Required | Description |
|------|----------|-------------|
| `--from-user` | Yes | Source user |
| `--to-user` | Yes | Destination user |
| `--start-date` | Yes | Start time (RFC3339) |
| `--end-date` | Yes | End time (RFC3339) |

### `forwarding-rules update <id>`

Update a forwarding rule.

### `forwarding-rules delete <id>`

Delete a forwarding rule.

---

## Incident Infrastructure

### `services list`

List all services.

### `services get <id>`

Get a service by ID.

### `services create`

Create a new service.

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | Service name |
| `--team-id` | Yes | Owning team ID |
| `--description` | | Description |

### `services update <id>`

Update a service.

### `services delete <id>`

Delete a service.

### `postmortems get <id>`

Get a postmortem by ID.

### `postmortems create`

Create a postmortem linked to an incident.

| Flag | Required | Description |
|------|----------|-------------|
| `--incident-id` | Yes | Linked incident ID |
| `--description` | | Description |

### `postmortems update <id>`

Update a postmortem.

### `postmortems delete <id>`

Delete a postmortem.

### `deployments list`

List deployments for a service.

| Flag | Required | Description |
|------|----------|-------------|
| `--service` | Yes | Service ID to list deployments for |
| `--environment` | | Filter by environment |

### `deployments get <id>`

Get a deployment by ID.

### `deployments create`

Create a deployment record.

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | Deployment name |
| `--description` | | Deployment description |
| `--service-id` | | Service ID |
| `--environment` | | Deployment environment |

### `deployments update <id>`

Update a deployment.

### `deployments search`

Search deployments.

| Flag | Required | Description |
|------|----------|-------------|
| `--service` | Yes | Service ID to search deployments for |
| `--environment` | | Filter by environment |

---

## API Behavior

### Rate Limiting
The client automatically retries on 429 (rate limited) responses with exponential backoff, up to 3 retries.

### Async Operations
Some operations return HTTP 202 (Accepted). The client automatically polls until the operation completes.

### Pagination
List commands with `--all` use offset-based pagination to fetch all pages automatically.

### Error Format
API errors return structured JSON:
```json
{"code": "ERROR", "message": "OpsGenie API error 404: Alert with id [abc] not found.", "recoverable": false}
```

---

## Account & Utilities

### `account get`

Get account information (verify connectivity).

```bash
opsgenie-cli account get
```

### `docs`

Display the full documentation (README.md).

```bash
opsgenie-cli docs
```

### `skill print`

Print the embedded SKILL.md to stdout.

```bash
opsgenie-cli skill print
```

### `skill add`

Install the skill to `~/.claude/skills/opsgenie-cli/`.

```bash
opsgenie-cli skill add
```

### `completion`

Generate shell completion scripts.

```bash
opsgenie-cli completion bash >> ~/.bashrc
opsgenie-cli completion zsh >> ~/.zshrc
opsgenie-cli completion fish > ~/.config/fish/completions/opsgenie-cli.fish
```
