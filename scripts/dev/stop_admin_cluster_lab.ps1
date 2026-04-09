#!/usr/bin/env pwsh
[CmdletBinding()]
param(
    [string]$WorkDir = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$rootDir = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
if ([string]::IsNullOrWhiteSpace($WorkDir)) {
    $WorkDir = Join-Path $rootDir ".tmp/admin_cluster_lab"
}
$pidsFile = Join-Path $WorkDir "pids.json"

if (-not (Test-Path -LiteralPath $pidsFile -PathType Leaf)) {
    Write-Host "No running admin cluster lab found ($pidsFile)"
    return
}

$payload = Get-Content -LiteralPath $pidsFile -Raw | ConvertFrom-Json

function Stop-ByPid {
    param([int]$Pid)
    if ($Pid -le 0) {
        return
    }
    $proc = Get-Process -Id $Pid -ErrorAction SilentlyContinue
    if ($null -ne $proc) {
        Stop-Process -Id $Pid -Force -ErrorAction SilentlyContinue
    }
}

Stop-ByPid -Pid ([int]$payload.lb_pid)
Stop-ByPid -Pid ([int]$payload.node_a_pid)
Stop-ByPid -Pid ([int]$payload.node_b_pid)

if ($payload.redis_started_by_script -eq $true) {
    $docker = Get-Command docker -ErrorAction SilentlyContinue
    if ($null -ne $docker) {
        & docker stop $payload.redis_container_name | Out-Null
    }
}

Remove-Item -LiteralPath $pidsFile -Force
Write-Host "Stopped admin cluster lab (workdir: $WorkDir)"

