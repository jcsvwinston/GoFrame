# Admin Cluster Lab (2 Nodes + LB + Redis)

Reference date: 2026-04-09.
Status: Current.

This guide provides a local, reproducible lab to validate cluster-aware admin behavior:

- two app nodes
- one Redis relay/session backend
- one local round-robin load balancer

## Objective

Validate that `/admin` live telemetry can aggregate cross-node activity and that admin auth/session works through a non-sticky load balancer.

## Prerequisites

- Go available in `PATH`
- curl available in `PATH`
- Redis reachable at `redis://127.0.0.1:6379/0`
- optional: Docker (for auto-start Redis mode in helper scripts)

## Run (bash)

```bash
scripts/dev/run_admin_cluster_lab.sh
```

Optional detached mode:

```bash
scripts/dev/run_admin_cluster_lab.sh --detach
```

Stop:

```bash
scripts/dev/stop_admin_cluster_lab.sh
```

## Run (PowerShell)

```powershell
pwsh -File scripts/dev/run_admin_cluster_lab.ps1
```

Optional detached mode:

```powershell
pwsh -File scripts/dev/run_admin_cluster_lab.ps1 -Detach
```

Stop:

```powershell
pwsh -File scripts/dev/stop_admin_cluster_lab.ps1
```

## Endpoints

After startup:

- LB: `http://127.0.0.1:8090/admin`
- Node A direct: `http://127.0.0.1:8091/admin`
- Node B direct: `http://127.0.0.1:8092/admin`

## What to verify in admin

On `Live`:

- `Cluster relay` shows `Connected`
- `Cluster topology` shows both `node-a` and `node-b`
- `Node scope` filter can isolate one node while telemetry still arrives from both in global mode

On `System`:

- `Telemetry` card shows OTLP and trace-link status

## Script controls

Important options shared by both scripts:

- Redis URL: `--redis-url` / `-RedisUrl`
- ports: node A/B and LB
- channel/token: cluster relay hardening
- trace URL template: `{trace_id}` target for external trace explorers
- Redis start mode: `auto|docker|never`

## Notes

- Scripts configure `examples/mvc_api` with:
  - `session_store=redis` for shared sessions through LB
  - `admin_cluster_enabled=true` plus node identity/channel/token
- Logs and pid metadata are stored under `.tmp/admin_cluster_lab`.

