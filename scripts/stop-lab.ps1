param(
    [string]$Project = "dev",
    [switch]$Full = $false
)

function Write-Log {
    param([string]$Level, [string]$Message, [ConsoleColor]$Color)
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $Color
}

Write-Host "`n=== KUBE-BACKLAB SHUTDOWN ===" -ForegroundColor Yellow

$Namespace = $Project.ToLower()
Write-Log "INFO" "Removing application deployments for project '$Namespace'..." White
$env:PROJECT_PATH = "src"
skaffold delete --namespace $Namespace

if ($Full) {
    # 2. Stop Cluster
    Write-Log "INFO" "Stopping k3d cluster 'dev-cluster'..." White
    k3d cluster stop dev-cluster
    Write-Log "SUCCESS" "Project '$Namespace' removed and Cluster stopped." Green
} else {
    Write-Log "SUCCESS" "Project '$Namespace' removed. Cluster is still running for other projects." Green
    Write-Host "TIP: To stop the whole cluster, use: ./stop-lab.ps1 -Full" -ForegroundColor Gray
}

Write-Host "To restart, run: powershell ./start-lab.ps1 -Project $Namespace`n" -ForegroundColor Gray
