# Goal: Build opsgenie-cli — a complete Go CLI for the OpsGenie API

A comprehensive CLI tool for interacting with the Atlassian OpsGenie REST API v2, covering all 23+ resources (alerts, incidents, teams, schedules, users, heartbeats, escalations, integrations, and more). Hosted privately on Gitea with brew distribution via roboalchemist/private tap. Success = all 6 Phase 7 reviewers pass + brew formula installs + skill works in Claude Code.

## Phase 1 — Foundation {⬜ NOT STARTED}
Scaffold + pkg/ infrastructure. Binary builds, --help works, auth works against live API.
### 1A — Scaffold: directory structure, go.mod, Makefile
### 1B — pkg/api: HTTP client with rate limiting, pagination, error handling, async request polling
### 1C — pkg/auth: env var → config file chain (OPSGENIE_API_KEY)
### 1D — pkg/output: table/JSON/plaintext + --fields + --jq

## Phase 2 — Commands {⬜ NOT STARTED}
All API resources implemented as cobra commands. Full CRUD per resource.
### 2A — cmd/root.go: global flags (--json, --plaintext, --no-color, --debug, --region), GNU standards, exit codes
### 2B — cmd/<resource>.go: one file per API resource
- alerts (list, get, create, delete, acknowledge, close, snooze, escalate, assign, add-note, add-tags, remove-tags, count)
- incidents (list, get, create, close, resolve, reopen, delete, add-note, add-tags)
- teams (list, get, create, update, delete)
- team-members (add, remove)
- team-routing-rules (list, get, create, update, delete)
- schedules (list, get, create, update, delete, export-ical)
- schedule-rotations (list, get, create, update, delete)
- schedule-overrides (list, get, create, update, delete)
- on-call (get, next)
- escalations (list, get, create, update, delete)
- users (list, get, create, update, delete)
- contacts (list, get, create, update, delete, enable, disable)
- notification-rules (list, get, create, update, delete, enable, disable)
- heartbeats (list, get, create, update, delete, enable, disable, ping)
- integrations (list, get, create, update, delete, enable, disable)
- maintenance (list, get, create, update, delete, cancel)
- services (list, get, create, update, delete)
- policies (list, get, create, update, delete, enable, disable)
- forwarding-rules (list, get, create, update, delete)
- custom-roles (list, get, create, update, delete)
- postmortems (list, get, create, update, delete)
- deployments (list, get, create, update, search)
- account (get)
### 2C — Built-ins: docs, completion, skill print/add

## Phase 3 — Tests {⬜ NOT STARTED}
All three tiers pass. 90%+ unit coverage. Integration covers every command × every output flag.
### 3A — Unit tests (pkg/api, pkg/auth, pkg/output via httptest mocks)
### 3B — Integration tests (CRUDL lifecycle, cross-product, READONLY=1 gate — NOTE: we only have a RO key, so write tests always skip)
### 3C — Smoke tests (make test passes, no API key needed)

## Phase 4 — Docs & Skill {⬜ NOT STARTED}
### 4A — skill/SKILL.md + skill/reference/commands.md
### 4B — README.md + llms.txt

## Phase 5 — Release Automation {⬜ NOT STARTED}
### 5A — .gitea/workflows/bump-tap.yml
### 5B — CI workflow + required secrets documented
### 5C — Initial brew formula in roboalchemist/private tap

## Phase 6 — Definition of Done {⬜ NOT STARTED}
Every check verified with actual output. No assumptions.

## Phase 7 — Review (6 Reviewers) {⬜ NOT STARTED}
All 6 reviewers pass. Loop on failures.

## Phase 8 — Claude Integration {⬜ NOT STARTED}
Skill installed, enforcer updated, jset sync-all complete.

## Available Resources
- **Skill**: creating-go-cli (all templates + patterns)
- **API docs**: /tmp/opsgenie_api_v2_comprehensive_reference.md (35KB), /tmp/opsgenie_api_v2_quick_reference.md (9KB)
- **Credentials**: "Eng Opsgenie RO Key" in 1Password (item UUID: fpgdsvrbgnavuhxneamumyxnme, field: credential) — READ-ONLY
- **Hosting**: Gitea (private) — roboalchemist/private tap
- **Env var**: OPSGENIE_API_KEY
- **Base URLs**: api.opsgenie.com (US), api.eu.opsgenie.com (EU)
- **Auth header**: Authorization: GenieKey <key>

## API Quirks to Handle
- Incidents use /v1/incidents (not /v2)
- Teams don't paginate (returns all at once)
- Async operations return 202 with requestId — must poll
- Services & Logs APIs are Enterprise/Standard only
- Integration API doesn't support Zendesk/Slack/Incoming Call
- Flexible identifiers (ID or name) with identifierType param
- Pagination is offset-based (limit + offset)

## Success Criteria
- Phase 1: `make build` succeeds, `opsgenie-cli --help` runs, auth connects to live API
- Phase 3: `make test-unit` ≥ 90%, `make test-integration` passes with real key (READONLY=1 default), `make test` passes
- Phase 7: All 6 reviewers PASS
- Phase 8: `brew install --build-from-source` + `opsgenie-cli skill add` both work
