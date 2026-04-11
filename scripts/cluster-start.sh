#!/usr/bin/env bash
set -e

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOGDIR="$REPO_ROOT/.tmp/cluster"
mkdir -p "$LOGDIR"

export PATH=$PATH:/usr/local/go/bin

REDIS_URL="redis://127.0.0.1:6379"

echo "==> Building binaries..."
cd "$REPO_ROOT"
go build -o .tmp/goframe-node    ./examples/mvc_api
go build -o .tmp/goframe-lb      ./cmd/balancer
go build -o .tmp/goframe-store   ./cmd/ministore

echo "==> Launching ministore (Redis-compatible) on :6379 ..."
.tmp/goframe-store > "$LOGDIR/store.log" 2>&1 &
STORE_PID=$!
echo "    ministore PID=$STORE_PID  log=$LOGDIR/store.log"
sleep 1.0  # increased delay for Mac stability 


echo "==> Launching node1 on :8091 ..."
GOFRAME_EXAMPLE_PORT=8091 \
GOFRAME_EXAMPLE_DB_URL="sqlite://examples_mvc_api_cluster.db" \
GOFRAME_EXAMPLE_ADMIN_CLUSTER_NODE_ID=node1 \
GOFRAME_EXAMPLE_ADMIN_CLUSTER_ENABLED=1 \
GOFRAME_EXAMPLE_ADMIN_CLUSTER_REDIS_URL="$REDIS_URL" \
GOFRAME_EXAMPLE_REDIS_URL="$REDIS_URL" \
GOFRAME_EXAMPLE_SESSION_STORE=redis \
GOFRAME_EXAMPLE_SESSION_REDIS_URL="$REDIS_URL" \
  .tmp/goframe-node > "$LOGDIR/node1.log" 2>&1 &
NODE1_PID=$!
echo "    node1 PID=$NODE1_PID  log=$LOGDIR/node1.log"

echo "==> Launching node2 on :8092 ..."
GOFRAME_EXAMPLE_PORT=8092 \
GOFRAME_EXAMPLE_DB_URL="sqlite://examples_mvc_api_cluster.db" \
GOFRAME_EXAMPLE_ADMIN_CLUSTER_NODE_ID=node2 \
GOFRAME_EXAMPLE_ADMIN_CLUSTER_ENABLED=1 \
GOFRAME_EXAMPLE_ADMIN_CLUSTER_REDIS_URL="$REDIS_URL" \
GOFRAME_EXAMPLE_REDIS_URL="$REDIS_URL" \
GOFRAME_EXAMPLE_SESSION_STORE=redis \
GOFRAME_EXAMPLE_SESSION_REDIS_URL="$REDIS_URL" \
  .tmp/goframe-node > "$LOGDIR/node2.log" 2>&1 &
NODE2_PID=$!
echo "    node2 PID=$NODE2_PID  log=$LOGDIR/node2.log"

sleep 1

echo "==> Launching load balancer on :8090 -> :8091, :8092 ..."
.tmp/goframe-lb > "$LOGDIR/lb.log" 2>&1 &
LB_PID=$!
echo "    lb    PID=$LB_PID     log=$LOGDIR/lb.log"

echo ""
echo "Cluster running:"
echo "  Shared store  : $REDIS_URL"
echo "  Load balancer : http://localhost:8090/admin  (sticky sessions)"
echo "  Node1 direct  : http://localhost:8091/admin"
echo "  Node2 direct  : http://localhost:8092/admin"
echo ""
echo "PIDs: store=$STORE_PID  node1=$NODE1_PID  node2=$NODE2_PID  lb=$LB_PID"
echo "$STORE_PID $NODE1_PID $NODE2_PID $LB_PID" > "$LOGDIR/pids"
echo ""
echo "Stop all with: bash scripts/cluster-stop.sh"
echo "Stream logs  : tail -f .tmp/cluster/*.log"
