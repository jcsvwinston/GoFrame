#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Validate stable contract freeze baselines (no removals) for CLI and config.

Usage:
  bash scripts/ci/check_contract_freeze.sh
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

go test ./contracts -run '^TestContractFreeze_' -count=1 -v
