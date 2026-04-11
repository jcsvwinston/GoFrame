#!/usr/bin/env bash
PIDFILE="$(cd "$(dirname "$0")/.." && pwd)/.tmp/cluster/pids"
if [ ! -f "$PIDFILE" ]; then
  echo "No running cluster found (pids file missing)."
  exit 0
fi
PIDS=$(cat "$PIDFILE")
echo "Stopping cluster PIDs: $PIDS"
for pid in $PIDS; do
  kill "$pid" 2>/dev/null && echo "  killed $pid" || echo "  $pid already gone"
done
rm -f "$PIDFILE"
echo "Done."
