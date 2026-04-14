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
"$PI_BIN" -e ./pi-fabbro -p '/fabbro-review Smoke test review content.' >/dev/null

echo "pi-fabbro smoke test passed"
