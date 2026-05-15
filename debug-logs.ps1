param(
    [switch]$Watch = $false
)

function Write-Log {
    param([string]$Level, [string]$Message, [ConsoleColor]$Color)
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $Color
}

Write-Host "`n=== KUBE-BACKLAB DIAGNOSTIC LOGS ===" -ForegroundColor Cyan

$failingPods = kubectl get pods -A --no-headers | Where-Object { $_ -notmatch "Running" -and $_ -notmatch "Completed" }

if ($failingPods) {
    Write-Log "ERROR" "Found failing pods in the cluster!" Red
    foreach ($pod in $failingPods) {
        Write-Host " > $pod" -ForegroundColor Red
    }
    
    Write-Host "`n--- [ LAST ERROR LOGS ] ---" -ForegroundColor Yellow
    foreach ($podLine in $failingPods) {
        $parts = $podLine -split "\s+"
        $ns = $parts[0]
        $name = $parts[1]
        Write-Host "[$ns/$name] Recent Logs:" -ForegroundColor Gray
        kubectl logs $name -n $ns --tail=10
        Write-Host "------------------------" -ForegroundColor Gray
    }
} else {
    Write-Log "SUCCESS" "All pods are healthy (Running/Completed)." Green
}

Write-Host "`n--- [ APPLICATION LOGS (dev) ] ---" -ForegroundColor Cyan
$appPod = kubectl get pods -n dev -l app=hello-app -o name | Select-Object -First 1
if ($appPod) {
    if ($Watch) {
        Write-Log "INFO" "Streaming application logs... (Ctrl+C to stop)" White
        kubectl logs -n dev $appPod -f
    } else {
        kubectl logs -n dev $appPod --tail=20
    }
} else {
    Write-Log "WARN" "No application pod found in 'dev' namespace." Yellow
}

Write-Host "`n--- [ INFRASTRUCTURE LOGS (infra) ] ---" -ForegroundColor Cyan
$dbPod = kubectl get pods -n infra -l app.kubernetes.io/name=postgresql -o name | Select-Object -First 1
if ($dbPod) {
    Write-Log "INFO" "Last 5 DB logs:" White
    kubectl logs -n infra $dbPod --tail=5
} else {
    Write-Log "WARN" "No DB pod found in 'infra' namespace." Yellow
}

Write-Host "`nTip: Use 'k9s' for a better interactive experience.`n" -ForegroundColor Gray
