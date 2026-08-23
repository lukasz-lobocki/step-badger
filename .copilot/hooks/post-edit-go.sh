#!/usr/bin/env bash
# Formats and vets a Go file immediately after Copilot edits it.
set -euo pipefail

input=$(cat)
file=$(printf '%s' "$input" | jq -r '.tool_input.path // .tool_input.file_path // ""')

[[ "$file" == *.go ]] || exit 0
[[ -f "$file" ]] || exit 0

msgs=()

if command -v gofumpt >/dev/null 2>&1; then
  gofumpt -w "$file" && msgs+=("gofumpt applied to $file")
else
  gofmt -s -w "$file" && msgs+=("gofmt -s applied to $file")
fi

if command -v goimports >/dev/null 2>&1; then
  goimports -w "$file"
fi

pkg=$(dirname "$file")
if ! vet_out=$(go vet "./$pkg" 2>&1); then
  msgs+=("go vet FAILED for ./$pkg:")
  msgs+=("$vet_out")
  msgs+=("Fix these before continuing.")
fi

if [[ "$file" != *_test.go ]]; then
  base=$(basename "$file" .go)
  [[ -f "$pkg/${base}_test.go" ]] || msgs+=("Note: no ${base}_test.go exists in $pkg.")
fi

if ((${#msgs[@]})); then
  printf '%s\n' "${msgs[@]}" >&2
  jq -n --arg c "$(printf '%s\n' "${msgs[@]}")" \
    '{hookSpecificOutput: {hookEventName: "PostToolUse", additionalContext: $c}}'
fi
exit 0