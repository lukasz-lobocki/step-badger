#!/usr/bin/env bash
# Runs the full Go verification suite before the agent is allowed to stop.
set -uo pipefail

fail() {
  jq -n --arg r "$1" \
    '{hookSpecificOutput: {hookEventName: "Stop", decision: "block", reason: $r}}'
  exit 0
}

git diff --quiet HEAD -- '*.go' 'go.mod' 'go.sum' && exit 0

if out=$(gofmt -s -l . 2>&1) && [[ -n "$out" ]]; then
  fail "Unformatted Go files:\n$out\nRun: gofmt -s -w ."
fi

if ! out=$(go build ./... 2>&1); then
  fail "go build failed:\n$out"
fi

if ! out=$(go vet ./... 2>&1); then
  fail "go vet failed:\n$out"
fi

if ! out=$(go test ./... -race -count=1 2>&1); then
  fail "Tests failed:\n$(printf '%s' "$out" | tail -60)"
fi

if command -v golangci-lint >/dev/null 2>&1 && [[ -f .golangci.yml || -f .golangci.yaml ]]; then
  if ! out=$(golangci-lint run 2>&1); then
    fail "golangci-lint failed:\n$(printf '%s' "$out" | tail -60)"
  fi
fi

exit 0