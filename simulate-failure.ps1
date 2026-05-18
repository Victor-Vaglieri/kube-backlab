param(
    [string]$Project = "dev"
)

function Write-Log {
    param([string]$Level, [string]$Message, [ConsoleColor]$Color)
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $Color
}

$Namespace = $Project.ToLower()

Write-Host "`n=== KUBE-BACKLAB FAILURE SIMULATOR ===" -ForegroundColor Red
Write-Host "Target Namespace: $Namespace" -ForegroundColor Gray

$pods = kubectl get pods -n $Namespace -o name
if (-not $pods) {
    Write-Log "ERROR" "No pods found in '$Namespace' namespace." Red
    exit
}

$podToKill = $pods | Get-Random
Write-Log "ACTION" "Simulating failure by killing pod: $podToKill" Yellow

kubectl delete $podToKill -n $Namespace --grace-period=0 --force

Write-Log "INFO" "Waiting for Kubernetes to self-heal..." White
Start-Sleep -Seconds 2

Write-Host "`n--- [ HEALING STATUS ] ---" -ForegroundColor Cyan
kubectl get pods -n $Namespace
Write-Log "SUCCESS" "Kubernetes is recreating the pod automatically!" Green
