#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PI_BIN="${PI_BIN:-$(command -v pi)}"

if [[ -z "${PI_BIN:-}" ]]; then
  echo "pi not found on PATH" >&2
  exit 1
fi

cd "$ROOT"

"$PI_BIN" -e ./pi-fabbro -p '/fabbro-status' >/dev/null
"$PI_BIN" -e ./pi-fabbro -p '/fabbro-prime' >/dev/null

review_output="$($PI_BIN -e ./pi-fabbro -p '/fabbro-review Smoke test review content.' 2>&1)"
session_id="$(printf '%s\n' "$review_output" | grep -oE '[0-9]{8}-[a-f0-9]+' | head -1)"

if [[ -z "$session_id" ]]; then
  echo "failed to extract session ID from /fabbro-review output" >&2
  printf '%s\n' "$review_output" >&2
  exit 1
fi

"$PI_BIN" -e ./pi-fabbro -p "/fabbro-apply $session_id" >/dev/null
"$PI_BIN" -e ./pi-fabbro -p '/fabbro-sessions' >/dev/null

echo "pi-fabbro smoke test passed"
