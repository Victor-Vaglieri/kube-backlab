param(
    [string]$Path = "src"
)

function Write-Log {
    param([string]$Level, [string]$Message, [ConsoleColor]$Color)
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $Color
}

# 0. Setup Environment
$env:PROJECT_PATH = Resolve-Path $Path
Write-Host "`n=== KUBE-BACKLAB MASTER SETUP ===" -ForegroundColor Cyan
Write-Host "Target Project: $($env:PROJECT_PATH)" -ForegroundColor Gray

# 1. Check Cluster
Write-Log "INFO" "Checking k3d cluster..." White
$cluster = k3d cluster list | Select-String "dev-cluster"
if (-not $cluster) {
    Write-Log "ACTION" "Creating dev-cluster..." Yellow
    k3d cluster create dev-cluster -p "8080:80@loadbalancer"
} else {
    Write-Log "SUCCESS" "Cluster dev-cluster is ready." Green
}

# 2. Setup Namespaces
Write-Log "INFO" "Ensuring namespaces exist..." White
kubectl create namespace dev --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace infra --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -

# 3. Apply Project Specific Configs
$projectConfig = Join-Path $env:PROJECT_PATH "k8s-config.yaml"
if (Test-Path $projectConfig) {
    Write-Log "SUCCESS" "Detected k8s-config.yaml in project. Applying..." Green
    kubectl apply -f $projectConfig
} else {
    Write-Log "WARN" "No k8s-config.yaml found in project. Using lab defaults." Yellow
}

# 4. Setup Helm Charts (Infra)
# ... (rest of the script)
