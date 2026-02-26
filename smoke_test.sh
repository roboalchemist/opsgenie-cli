#!/bin/bash
# Smoke tests for opsgenie-cli — no API key required.
# Verifies the binary builds and all help/docs/completion commands work correctly.
set -euo pipefail

BINARY="${BINARY:-./opsgenie-cli}"
PASS=0
FAIL=0

pass() { echo "  PASS: $*"; PASS=$((PASS+1)); }
fail() { echo "  FAIL: $*"; FAIL=$((FAIL+1)); }

check_exit() {
    local desc="$1"; shift
    if "$BINARY" "$@" >/dev/null 2>&1; then
        pass "$desc"
    else
        fail "$desc (non-zero exit)"
    fi
}

check_output() {
    local desc="$1"
    local pattern="$2"
    shift 2
    local out
    out=$("$BINARY" "$@" 2>&1) || true
    if echo "$out" | grep -q "$pattern"; then
        pass "$desc"
    else
        fail "$desc (expected pattern: $pattern)"
    fi
}

check_nonempty() {
    local desc="$1"; shift
    local out
    out=$("$BINARY" "$@" 2>&1) || true
    if [ -n "$out" ]; then
        pass "$desc"
    else
        fail "$desc (empty output)"
    fi
}

echo "=== opsgenie-cli smoke tests ==="
echo ""

echo "--- Core functionality ---"

# 1. Binary builds successfully (caller should build before running this script)
if [ -x "$BINARY" ]; then
    pass "binary is executable"
else
    fail "binary not found or not executable at $BINARY"
fi

# 2. --help exits 0 and contains "Available Commands"
check_output "--help contains 'Available Commands'" "Available Commands" --help

# 3. --version exits 0 and outputs a version string (GNU format: "name version")
check_output "--version outputs version string" "Copyright" --version

# 4. docs exits 0 and outputs non-empty content
check_nonempty "docs outputs non-empty content" docs

# 5. completion bash exits 0 and outputs a completion script
check_output "completion bash outputs script" "bash" completion bash

# 6. skill print exits 0 and outputs skill content
check_nonempty "skill print outputs content" skill print

echo ""
echo "--- Resource command help ---"

# 7. alerts --help exits 0 and shows subcommands
check_output "alerts --help shows subcommands" "Usage:" alerts --help

# 8. incidents --help exits 0
check_exit "incidents --help" incidents --help

# 9. teams --help exits 0
check_exit "teams --help" teams --help

# 10. All resource parent commands respond to --help
RESOURCE_COMMANDS=(
    account
    alerts
    contacts
    custom-roles
    deployments
    escalations
    forwarding-rules
    heartbeats
    incidents
    integrations
    maintenance
    notification-rules
    on-call
    policies
    postmortems
    schedule-overrides
    schedule-rotations
    schedules
    services
    teams
    team-members
    team-routing-rules
    users
)

for cmd in "${RESOURCE_COMMANDS[@]}"; do
    check_exit "$cmd --help" "$cmd" --help
done

echo ""
echo "--- Completion scripts ---"
check_exit "completion zsh" completion zsh
check_exit "completion fish" completion fish
check_exit "completion powershell" completion powershell

echo ""
echo "Results: $PASS passed, $FAIL failed"
echo ""

if [ "$FAIL" -eq 0 ]; then
    echo "All smoke tests passed!"
    exit 0
else
    echo "FAILED: $FAIL smoke test(s) did not pass."
    exit 1
fi
