#!/usr/bin/env bash
# Emits Go toolchain context at the start of a Copilot session.
set -euo pipefail

context=$(
  printf 'Go toolchain: %s\n' "$(go version 2>/dev/null || echo 'go not found')"
  printf 'Module: %s\n' "$(go list -m 2>/dev/null || echo 'n/a')"
  printf 'go directive: %s\n' "$(go list -m -f '{{.GoVersion}}' 2>/dev/null || echo 'n/a')"
  printf 'Packages: %s\n' "$(go list ./... 2>/dev/null | wc -l | tr -d ' ')"
  printf 'golangci-lint: %s\n' "$(command -v golangci-lint >/dev/null && golangci-lint --version 2>/dev/null | head -1 || echo 'not installed')"
  printf 'Dirty files: %s\n' "$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')"
)

jq -n --arg c "$context" \
  '{hookSpecificOutput: {hookEventName: "SessionStart", additionalContext: $c}}'