#!/usr/bin/env bash
# Blocks destructive or policy-violating shell commands before execution.
set -euo pipefail

input=$(cat)
cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // .tool_input.cmd // ""')

deny() {
  jq -n --arg r "$1" \
    '{hookSpecificOutput: {hookEventName: "PreToolUse",
      permissionDecision: "deny", permissionDecisionReason: $r}}'
  exit 0
}

case "$cmd" in
  *"rm -rf /"*|*"rm -rf ~"*)          deny "Refusing recursive delete of a root path." ;;
  *"git push --force"*|*"git push -f"*) deny "Force-push is not permitted. Use --force-with-lease and ask first." ;;
  *"git reset --hard"*)                deny "Hard reset would discard work. Stash or commit instead." ;;
  *"git commit"*"--no-verify"*)        deny "Bypassing hooks is not permitted." ;;
  *"go mod edit -go"*)                 deny "Go version bumps must go through the Go Maintainer agent." ;;
  *"GOFLAGS=-mod=mod"*|*"-mod=mod"*)   deny "Use 'go mod tidy' explicitly rather than implicit module edits." ;;
  *"go clean -modcache"*)              deny "Clearing the module cache is disruptive; not needed for this task." ;;
  *"curl "*"| sh"*|*"wget "*"| sh"*)   deny "Piping remote scripts to a shell is not permitted." ;;
esac

# Warn (but allow) on skipped-verification test runs.
if [[ "$cmd" == *"go test"* && "$cmd" != *"-race"* && "$cmd" != *"-run"* ]]; then
  jq -n '{hookSpecificOutput: {hookEventName: "PreToolUse",
    permissionDecision: "allow",
    permissionDecisionReason: "Reminder: full test runs should include -race -count=1."}}'
  exit 0
fi

exit 0