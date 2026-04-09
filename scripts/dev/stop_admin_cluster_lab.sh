#!/usr/bin/env bash
set -euo pipefail

WORK_DIR=""
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORK_DIR="${ROOT_DIR}/.tmp/admin_cluster_lab"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --workdir)
      WORK_DIR="${2:-}"
      shift 2
      ;;
    -h|--help)
      echo "Usage: scripts/dev/stop_admin_cluster_lab.sh [--workdir <path>]"
      exit 0
      ;;
    *)
      echo "error: unknown option: $1" >&2
      exit 1
      ;;
  esac
done

PIDS_FILE="${WORK_DIR}/pids.env"
if [[ ! -f "${PIDS_FILE}" ]]; then
  echo "No running admin cluster lab found (${PIDS_FILE})"
  exit 0
fi

# shellcheck source=/dev/null
source "${PIDS_FILE}"

kill_if_running() {
  local pid="$1"
  if [[ -z "${pid}" ]]; then
    return 0
  fi
  if kill -0 "${pid}" >/dev/null 2>&1; then
    kill "${pid}" >/dev/null 2>&1 || true
  fi
}

kill_if_running "${LB_PID:-}"
kill_if_running "${NODE_A_PID:-}"
kill_if_running "${NODE_B_PID:-}"

if [[ "${REDIS_STARTED_BY_SCRIPT:-0}" == "1" ]]; then
  if command -v docker >/dev/null 2>&1; then
    docker stop "${REDIS_CONTAINER_NAME:-goframe-admin-cluster-redis}" >/dev/null 2>&1 || true
  fi
fi

rm -f "${PIDS_FILE}"
echo "Stopped admin cluster lab (workdir: ${WORK_DIR})"

