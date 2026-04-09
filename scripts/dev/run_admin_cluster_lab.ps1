#!/usr/bin/env pwsh
[CmdletBinding()]
param(
    [string]$RedisUrl = "redis://127.0.0.1:6379/0",
    [int]$NodeAPort = 8091,
    [int]$NodeBPort = 8092,
    [int]$LbPort = 8090,
    [string]$NodeADb = "sqlite://examples_mvc_api_node_a.db",
    [string]$NodeBDb = "sqlite://examples_mvc_api_node_b.db",
    [string]$ClusterChannel = "goframe:admin:live:v1",
    [string]$ClusterToken = "dev-cluster-token",
    [string]$TraceUrlTemplate = "",
    [string]$WorkDir = "",
    [ValidateSet("auto", "docker", "never")]
    [string]$StartRedis = "auto",
    [switch]$Detach
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$rootDir = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
if ([string]::IsNullOrWhiteSpace($WorkDir)) {
    $WorkDir = Join-Path $rootDir ".tmp/admin_cluster_lab"
}
New-Item -Path $WorkDir -ItemType Directory -Force | Out-Null

$nodeALog = Join-Path $WorkDir "node-a.log"
$nodeBLog = Join-Path $WorkDir "node-b.log"
$lbLog = Join-Path $WorkDir "lb.log"
$pidsFile = Join-Path $WorkDir "pids.json"
$redisContainer = "goframe-admin-cluster-redis"
$redisStartedByScript = $false

if (Test-Path -LiteralPath $pidsFile -PathType Leaf) {
    throw "Existing pid file found at $pidsFile. Stop previous run first with scripts/dev/stop_admin_cluster_lab.ps1 -WorkDir `"$WorkDir`""
}

function Test-RedisReady {
    try {
        $cmd = Get-Command redis-cli -ErrorAction Stop
        $out = & $cmd.Source -u $RedisUrl ping 2>$null
        return ($LASTEXITCODE -eq 0 -and $out -match "PONG")
    }
    catch {
        return $false
    }
}

function Start-RedisIfNeeded {
    if (Test-RedisReady) {
        return $true
    }
    if ($StartRedis -eq "never") {
        return $false
    }
    $docker = Get-Command docker -ErrorAction SilentlyContinue
    if ($null -eq $docker) {
        return $false
    }
    $running = (& docker ps --format "{{.Names}}") -contains $redisContainer
    if (-not $running) {
        $existing = (& docker ps -a --format "{{.Names}}") -contains $redisContainer
        if ($existing) {
            & docker start $redisContainer | Out-Null
        }
        else {
            & docker run -d --name $redisContainer -p 6379:6379 redis:7-alpine | Out-Null
        }
        $script:redisStartedByScript = $true
        Start-Sleep -Seconds 1
    }
    return (Test-RedisReady)
}

function Start-ProcessWithEnv {
    param(
        [string]$LogPath,
        [hashtable]$EnvMap,
        [string[]]$Args
    )

    $previous = @{}
    foreach ($k in $EnvMap.Keys) {
        $previous[$k] = [Environment]::GetEnvironmentVariable($k, "Process")
        [Environment]::SetEnvironmentVariable($k, [string]$EnvMap[$k], "Process")
    }
    try {
        return Start-Process -FilePath "go" -ArgumentList $Args -WorkingDirectory $rootDir -PassThru -RedirectStandardOutput $LogPath -RedirectStandardError $LogPath
    }
    finally {
        foreach ($k in $EnvMap.Keys) {
            [Environment]::SetEnvironmentVariable($k, $previous[$k], "Process")
        }
    }
}

function Wait-HttpReady {
    param([string]$Url)
    for ($i = 0; $i -lt 120; $i++) {
        try {
            $resp = Invoke-WebRequest -Uri $Url -Method GET -TimeoutSec 1 -UseBasicParsing
            if ($resp.StatusCode -ge 200 -and $resp.StatusCode -lt 500) {
                return $true
            }
        }
        catch {
        }
        Start-Sleep -Milliseconds 500
    }
    return $false
}

if (-not (Start-RedisIfNeeded)) {
    throw "Redis is not reachable at $RedisUrl. Start redis manually or use -StartRedis docker."
}

Write-Host "==> Starting admin cluster lab (PowerShell)"
Write-Host "    redis:         $RedisUrl"
Write-Host "    node-a:        http://127.0.0.1:$NodeAPort"
Write-Host "    node-b:        http://127.0.0.1:$NodeBPort"
Write-Host "    load balancer: http://127.0.0.1:$LbPort"

$nodeAEnv = @{
    GOFRAME_EXAMPLE_PORT = "$NodeAPort"
    GOFRAME_EXAMPLE_DB_URL = "$NodeADb"
    GOFRAME_EXAMPLE_REDIS_URL = "$RedisUrl"
    GOFRAME_EXAMPLE_SESSION_STORE = "redis"
    GOFRAME_EXAMPLE_SESSION_REDIS_URL = "$RedisUrl"
    GOFRAME_EXAMPLE_ADMIN_CLUSTER_ENABLED = "true"
    GOFRAME_EXAMPLE_ADMIN_CLUSTER_REDIS_URL = "$RedisUrl"
    GOFRAME_EXAMPLE_ADMIN_CLUSTER_CHANNEL = "$ClusterChannel"
    GOFRAME_EXAMPLE_ADMIN_CLUSTER_NODE_ID = "node-a"
    GOFRAME_EXAMPLE_ADMIN_CLUSTER_TOKEN = "$ClusterToken"
    GOFRAME_EXAMPLE_ADMIN_TRACE_URL_TEMPLATE = "$TraceUrlTemplate"
    GOFRAME_EXAMPLE_ADMIN_TITLE = "GoFrame Admin Node A"
}
$nodeBEnv = @{
    GOFRAME_EXAMPLE_PORT = "$NodeBPort"
    GOFRAME_EXAMPLE_DB_URL = "$NodeBDb"
    GOFRAME_EXAMPLE_REDIS_URL = "$RedisUrl"
    GOFRAME_EXAMPLE_SESSION_STORE = "redis"
    GOFRAME_EXAMPLE_SESSION_REDIS_URL = "$RedisUrl"
    GOFRAME_EXAMPLE_ADMIN_CLUSTER_ENABLED = "true"
    GOFRAME_EXAMPLE_ADMIN_CLUSTER_REDIS_URL = "$RedisUrl"
    GOFRAME_EXAMPLE_ADMIN_CLUSTER_CHANNEL = "$ClusterChannel"
    GOFRAME_EXAMPLE_ADMIN_CLUSTER_NODE_ID = "node-b"
    GOFRAME_EXAMPLE_ADMIN_CLUSTER_TOKEN = "$ClusterToken"
    GOFRAME_EXAMPLE_ADMIN_TRACE_URL_TEMPLATE = "$TraceUrlTemplate"
    GOFRAME_EXAMPLE_ADMIN_TITLE = "GoFrame Admin Node B"
}

$nodeAProc = Start-ProcessWithEnv -LogPath $nodeALog -EnvMap $nodeAEnv -Args @("run", "./examples/mvc_api")
$nodeBProc = Start-ProcessWithEnv -LogPath $nodeBLog -EnvMap $nodeBEnv -Args @("run", "./examples/mvc_api")
$lbProc = Start-Process -FilePath "go" -ArgumentList @(
    "run", "./scripts/dev/local_lb.go",
    "--listen", ":$LbPort",
    "--targets", "http://127.0.0.1:$NodeAPort,http://127.0.0.1:$NodeBPort"
) -WorkingDirectory $rootDir -PassThru -RedirectStandardOutput $lbLog -RedirectStandardError $lbLog

if (-not (Wait-HttpReady -Url "http://127.0.0.1:$NodeAPort/api/health")) {
    throw "node-a did not become ready. See $nodeALog"
}
if (-not (Wait-HttpReady -Url "http://127.0.0.1:$NodeBPort/api/health")) {
    throw "node-b did not become ready. See $nodeBLog"
}
if (-not (Wait-HttpReady -Url "http://127.0.0.1:$LbPort/api/health")) {
    throw "load balancer did not become ready. See $lbLog"
}

$payload = [ordered]@{
    root_dir = $rootDir
    work_dir = $WorkDir
    redis_url = $RedisUrl
    redis_container_name = $redisContainer
    redis_started_by_script = $redisStartedByScript
    node_a_pid = $nodeAProc.Id
    node_b_pid = $nodeBProc.Id
    lb_pid = $lbProc.Id
}
$payload | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath $pidsFile -NoNewline

Write-Host "==> Ready"
Write-Host "    dashboard via LB: http://127.0.0.1:$LbPort/admin"
Write-Host "    node-a direct:    http://127.0.0.1:$NodeAPort/admin"
Write-Host "    node-b direct:    http://127.0.0.1:$NodeBPort/admin"
Write-Host "    logs:"
Write-Host "      $nodeALog"
Write-Host "      $nodeBLog"
Write-Host "      $lbLog"
Write-Host "    stop:"
Write-Host "      pwsh -File scripts/dev/stop_admin_cluster_lab.ps1 -WorkDir `"$WorkDir`""

if ($Detach.IsPresent) {
    return
}

Write-Host "==> Running in foreground. Press Ctrl+C to stop."
try {
    Wait-Process -Id $lbProc.Id
}
finally {
    & pwsh -File (Join-Path $PSScriptRoot "stop_admin_cluster_lab.ps1") -WorkDir $WorkDir | Out-Null
}

