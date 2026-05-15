function Write-Log {
    param([string]$Level, [string]$Message, [ConsoleColor]$Color)
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $Color
}

Write-Host "`n=== KUBE-BACKLAB FAILURE SIMULATOR ===" -ForegroundColor Red

$pods = kubectl get pods -n dev -o name
if (-not $pods) {
    Write-Log "ERROR" "No pods found in 'dev' namespace." Red
    exit
}

$podToKill = $pods | Get-Random
Write-Log "ACTION" "Simulating failure by killing pod: $podToKill" Yellow

kubectl delete $podToKill -n dev --grace-period=0 --force

Write-Log "INFO" "Waiting for Kubernetes to self-heal..." White
Start-Sleep -Seconds 2

Write-Host "`n--- [ HEALING STATUS ] ---" -ForegroundColor Cyan
kubectl get pods -n dev
Write-Log "SUCCESS" "Kubernetes is recreating the pod automatically!" Green
