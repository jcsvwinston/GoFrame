#!/usr/bin/env bash
set -euo pipefail

print_usage() {
  cat <<'USAGE'
Usage:
  scripts/dev/run_admin_cluster_lab.sh [options]

Options:
  --redis-url <url>          Redis URL shared by nodes and admin cluster relay (default: redis://127.0.0.1:6379/0)
  --node-a-port <port>       Node A HTTP port (default: 8091)
  --node-b-port <port>       Node B HTTP port (default: 8092)
  --lb-port <port>           Local load balancer port (default: 8090)
  --node-a-db <url>          Node A DB URL (default: sqlite://examples_mvc_api_node_a.db)
  --node-b-db <url>          Node B DB URL (default: sqlite://examples_mvc_api_node_b.db)
  --cluster-channel <name>   Admin live relay channel (default: goframe:admin:live:v1)
  --cluster-token <token>    Admin live relay token (default: dev-cluster-token)
  --trace-url-template <u>   Optional trace explorer template ({trace_id})
  --workdir <path>           Working directory for logs and pid file (default: .tmp/admin_cluster_lab)
  --start-redis <mode>       auto|docker|never (default: auto)
  --detach                   Start processes and exit immediately
  -h, --help                 Show help

Examples:
  scripts/dev/run_admin_cluster_lab.sh
  scripts/dev/run_admin_cluster_lab.sh --redis-url redis://127.0.0.1:6380/0 --lb-port 9000
  scripts/dev/run_admin_cluster_lab.sh --detach
USAGE
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REDIS_URL="redis://127.0.0.1:6379/0"
NODE_A_PORT=8091
NODE_B_PORT=8092
LB_PORT=8090
NODE_A_DB_URL="sqlite://examples_mvc_api_node_a.db"
NODE_B_DB_URL="sqlite://examples_mvc_api_node_b.db"
CLUSTER_CHANNEL="goframe:admin:live:v1"
CLUSTER_TOKEN="dev-cluster-token"
TRACE_URL_TEMPLATE=""
WORK_DIR="${ROOT_DIR}/.tmp/admin_cluster_lab"
START_REDIS_MODE="auto"
DETACH=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --redis-url)
      REDIS_URL="${2:-}"
      shift 2
      ;;
    --node-a-port)
      NODE_A_PORT="${2:-}"
      shift 2
      ;;
    --node-b-port)
      NODE_B_PORT="${2:-}"
      shift 2
      ;;
    --lb-port)
      LB_PORT="${2:-}"
      shift 2
      ;;
    --node-a-db)
      NODE_A_DB_URL="${2:-}"
      shift 2
      ;;
    --node-b-db)
      NODE_B_DB_URL="${2:-}"
      shift 2
      ;;
    --cluster-channel)
      CLUSTER_CHANNEL="${2:-}"
      shift 2
      ;;
    --cluster-token)
      CLUSTER_TOKEN="${2:-}"
      shift 2
      ;;
    --trace-url-template)
      TRACE_URL_TEMPLATE="${2:-}"
      shift 2
      ;;
    --workdir)
      WORK_DIR="${2:-}"
      shift 2
      ;;
    --start-redis)
      START_REDIS_MODE="${2:-}"
      shift 2
      ;;
    --detach)
      DETACH=1
      shift
      ;;
    -h|--help)
      print_usage
      exit 0
      ;;
    *)
      echo "error: unknown option: $1" >&2
      print_usage >&2
      exit 1
      ;;
  esac
done

if [[ "${START_REDIS_MODE}" != "auto" && "${START_REDIS_MODE}" != "docker" && "${START_REDIS_MODE}" != "never" ]]; then
  echo "error: --start-redis must be one of auto|docker|never" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "error: go is required" >&2
  exit 1
fi
if ! command -v curl >/dev/null 2>&1; then
  echo "error: curl is required" >&2
  exit 1
fi

mkdir -p "${WORK_DIR}"
NODE_A_LOG="${WORK_DIR}/node-a.log"
NODE_B_LOG="${WORK_DIR}/node-b.log"
LB_LOG="${WORK_DIR}/lb.log"
PIDS_FILE="${WORK_DIR}/pids.env"
NODE_BIN="${WORK_DIR}/mvc_api.bin"
LB_BIN="${WORK_DIR}/local_lb.bin"
REDIS_CONTAINER_NAME="goframe-admin-cluster-redis"
REDIS_STARTED_BY_SCRIPT=0

redis_is_ready() {
  if command -v redis-cli >/dev/null 2>&1; then
    if redis-cli -u "${REDIS_URL}" ping >/dev/null 2>&1; then
      return 0
    fi
  fi
  return 1
}

start_redis_if_needed() {
  if redis_is_ready; then
    return 0
  fi
  if [[ "${START_REDIS_MODE}" == "never" ]]; then
    return 1
  fi
  if ! command -v docker >/dev/null 2>&1; then
    return 1
  fi

  if docker ps --format '{{.Names}}' | grep -qx "${REDIS_CONTAINER_NAME}"; then
    return 0
  fi
  if docker ps -a --format '{{.Names}}' | grep -qx "${REDIS_CONTAINER_NAME}"; then
    docker start "${REDIS_CONTAINER_NAME}" >/dev/null
  else
    docker run -d --name "${REDIS_CONTAINER_NAME}" -p 6379:6379 redis:7-alpine >/dev/null
  fi
  REDIS_STARTED_BY_SCRIPT=1
  sleep 1
  redis_is_ready
}

wait_http_ready() {
  local url="$1"
  local attempts=60
  local i
  for ((i=1; i<=attempts; i++)); do
    if curl -fsS --max-time 1 "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.5
  done
  return 1
}

build_detach_binaries() {
  if [[ "${DETACH}" -ne 1 ]]; then
    return 0
  fi

  go -C "${ROOT_DIR}" build -o "${NODE_BIN}" ./examples/mvc_api
  go -C "${ROOT_DIR}" build -o "${LB_BIN}" ./scripts/dev/local_lb.go
  chmod +x "${NODE_BIN}" "${LB_BIN}"
}

if [[ -f "${PIDS_FILE}" ]]; then
  echo "error: existing pid file found at ${PIDS_FILE}" >&2
  echo "hint: stop previous lab with scripts/dev/stop_admin_cluster_lab.sh --workdir ${WORK_DIR}" >&2
  exit 1
fi

if ! start_redis_if_needed; then
  echo "error: Redis is not reachable at ${REDIS_URL}" >&2
  echo "hint: start redis manually or run with --start-redis docker" >&2
  exit 1
fi

build_detach_binaries

: > "${NODE_A_LOG}"
: > "${NODE_B_LOG}"
: > "${LB_LOG}"

echo "==> Starting admin cluster lab"
echo "    redis:        ${REDIS_URL}"
echo "    node-a:       http://127.0.0.1:${NODE_A_PORT}"
echo "    node-b:       http://127.0.0.1:${NODE_B_PORT}"
echo "    load balancer:http://127.0.0.1:${LB_PORT}"

start_node() {
  local node_id="$1"
  local port="$2"
  local db_url="$3"
  local title="$4"
  local log_file="$5"
  local -a env_vars=(
    "GOFRAME_EXAMPLE_PORT=${port}"
    "GOFRAME_EXAMPLE_DB_URL=${db_url}"
    "GOFRAME_EXAMPLE_REDIS_URL=${REDIS_URL}"
    "GOFRAME_EXAMPLE_SESSION_STORE=redis"
    "GOFRAME_EXAMPLE_SESSION_REDIS_URL=${REDIS_URL}"
    "GOFRAME_EXAMPLE_ADMIN_CLUSTER_ENABLED=true"
    "GOFRAME_EXAMPLE_ADMIN_CLUSTER_REDIS_URL=${REDIS_URL}"
    "GOFRAME_EXAMPLE_ADMIN_CLUSTER_CHANNEL=${CLUSTER_CHANNEL}"
    "GOFRAME_EXAMPLE_ADMIN_CLUSTER_NODE_ID=${node_id}"
    "GOFRAME_EXAMPLE_ADMIN_CLUSTER_TOKEN=${CLUSTER_TOKEN}"
    "GOFRAME_EXAMPLE_ADMIN_TRACE_URL_TEMPLATE=${TRACE_URL_TEMPLATE}"
    "GOFRAME_EXAMPLE_ADMIN_TITLE=${title}"
  )

  if [[ "${DETACH}" -eq 1 ]]; then
    nohup env "${env_vars[@]}" "${NODE_BIN}" >>"${log_file}" 2>&1 < /dev/null &
  else
    env "${env_vars[@]}" go -C "${ROOT_DIR}" run ./examples/mvc_api >"${log_file}" 2>&1 &
  fi
  STARTED_PID="$!"
}

STARTED_PID=""
start_node "node-a" "${NODE_A_PORT}" "${NODE_A_DB_URL}" "GoFrame Admin Node A" "${NODE_A_LOG}"
NODE_A_PID="${STARTED_PID}"
start_node "node-b" "${NODE_B_PORT}" "${NODE_B_DB_URL}" "GoFrame Admin Node B" "${NODE_B_LOG}"
NODE_B_PID="${STARTED_PID}"

start_lb() {
  if [[ "${DETACH}" -eq 1 ]]; then
    nohup "${LB_BIN}" \
      --listen ":${LB_PORT}" \
      --targets "http://127.0.0.1:${NODE_A_PORT},http://127.0.0.1:${NODE_B_PORT}" >>"${LB_LOG}" 2>&1 < /dev/null &
  else
    go -C "${ROOT_DIR}" run ./scripts/dev/local_lb.go \
      --listen ":${LB_PORT}" \
      --targets "http://127.0.0.1:${NODE_A_PORT},http://127.0.0.1:${NODE_B_PORT}" >"${LB_LOG}" 2>&1 &
  fi
  STARTED_PID="$!"
}

start_lb
LB_PID="${STARTED_PID}"

if ! wait_http_ready "http://127.0.0.1:${NODE_A_PORT}/api/health"; then
  echo "error: node-a did not become ready. See ${NODE_A_LOG}" >&2
  exit 1
fi
if ! wait_http_ready "http://127.0.0.1:${NODE_B_PORT}/api/health"; then
  echo "error: node-b did not become ready. See ${NODE_B_LOG}" >&2
  exit 1
fi
if ! wait_http_ready "http://127.0.0.1:${LB_PORT}/api/health"; then
  echo "error: load balancer did not become ready. See ${LB_LOG}" >&2
  exit 1
fi

cat > "${PIDS_FILE}" <<EOF
ROOT_DIR='${ROOT_DIR}'
WORK_DIR='${WORK_DIR}'
REDIS_URL='${REDIS_URL}'
REDIS_CONTAINER_NAME='${REDIS_CONTAINER_NAME}'
REDIS_STARTED_BY_SCRIPT='${REDIS_STARTED_BY_SCRIPT}'
NODE_A_PID='${NODE_A_PID}'
NODE_B_PID='${NODE_B_PID}'
LB_PID='${LB_PID}'
EOF

echo "==> Ready"
echo "    dashboard via LB: http://127.0.0.1:${LB_PORT}/admin"
echo "    node-a direct:    http://127.0.0.1:${NODE_A_PORT}/admin"
echo "    node-b direct:    http://127.0.0.1:${NODE_B_PORT}/admin"
echo "    logs:"
echo "      ${NODE_A_LOG}"
echo "      ${NODE_B_LOG}"
echo "      ${LB_LOG}"
echo "    stop:"
echo "      scripts/dev/stop_admin_cluster_lab.sh --workdir ${WORK_DIR}"

if [[ "${DETACH}" -eq 1 ]]; then
  echo "==> Detached mode enabled (nohup)"
  exit 0
fi

cleanup() {
  set +e
  scripts/dev/stop_admin_cluster_lab.sh --workdir "${WORK_DIR}" >/dev/null 2>&1
}
trap cleanup INT TERM

echo "==> Running in foreground. Press Ctrl+C to stop."
wait "${LB_PID}" || true
