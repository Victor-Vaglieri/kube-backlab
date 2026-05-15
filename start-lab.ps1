param(
    [string]$Path = "src",
    [string]$Project = "dev"
)

function Write-Log {
    param([string]$Level, [string]$Message, [ConsoleColor]$Color)
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $Color
}

function Write-Box {
    param([string]$Title, [string[]]$Lines, [ConsoleColor]$Color = "Cyan")
    $width = 60
    Write-Host ("+" + ("-" * ($width - 2)) + "+") -ForegroundColor $Color
    Write-Host ("| " + $Title.PadRight($width - 4) + " |") -ForegroundColor $Color
    Write-Host ("+" + ("-" * ($width - 2)) + "+") -ForegroundColor $Color
    foreach ($line in $Lines) {
        Write-Host ("| " + $line.PadRight($width - 4) + " |") -ForegroundColor $Color
    }
    Write-Host ("+" + ("-" * ($width - 2)) + "+") -ForegroundColor $Color
}

$env:PROJECT_PATH = Resolve-Path $Path
$env:PROJECT_NAME = $Project
Write-Host "`n=== KUBE-BACKLAB MASTER SETUP ===" -ForegroundColor Cyan
Write-Host "Target Project: $($env:PROJECT_PATH)" -ForegroundColor Gray
Write-Host "Namespace:      $($env:PROJECT_NAME)" -ForegroundColor Gray

# 1. Check Cluster
Write-Log "INFO" "Checking k3d cluster status..." White
$clusterStatus = k3d cluster list dev-cluster --no-headers
if (-not $clusterStatus) {
    Write-Log "ACTION" "Creating dev-cluster..." Yellow
    k3d cluster create dev-cluster -p "8080:80@loadbalancer"
} elseif ($clusterStatus -match "0/1") {
    Write-Log "ACTION" "Cluster exists but is stopped. Starting dev-cluster..." Yellow
    k3d cluster start dev-cluster
} else {
    Write-Log "SUCCESS" "Cluster dev-cluster is running and ready." Green
}

# 2. Setup Namespaces
Write-Log "INFO" "Ensuring namespaces exist..." White
kubectl create namespace $env:PROJECT_NAME --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace infra --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -

# 3. Apply Project Specific Configs
$projectConfig = Join-Path $env:PROJECT_PATH "k8s-config.yaml"
if (Test-Path $projectConfig) {
    Write-Log "SUCCESS" "Detected k8s-config.yaml in project. Applying to namespace $env:PROJECT_NAME..." Green
    kubectl apply -f $projectConfig -n $env:PROJECT_NAME
} else {
    Write-Log "INFO" "Using root k8s manifests via Skaffold." White
}

# 4. Infrastructure Status
Write-Host "`n--- [ INFRASTRUCTURE STATUS ] ---" -ForegroundColor Cyan
$pods = kubectl get pods -A --no-headers
$running = ($pods | Select-String "Running").Count
$total = $pods.Count
$healthColor = if ($running -eq $total) {"Green"} else {"Yellow"}
Write-Log "INFO" "Cluster Health: $running/$total pods running." $healthColor

# 5. Start Application (if not running)
Write-Log "INFO" "Ensuring application is deployed to namespace '$env:PROJECT_NAME'..." White
skaffold run --status-check=false --namespace $env:PROJECT_NAME

# 6. Apply Dynamic Routing (Isolation)
$appHost = if ($env:PROJECT_NAME -eq "dev") { "hello.dev.local" } else { "$($env:PROJECT_NAME).dev.local" }
Write-Log "INFO" "Configuring dynamic routing for $appHost..." White
# Patch the Ingress host to match the project name for true isolation
kubectl patch ingress hello-ingress -n $env:PROJECT_NAME --type='json' -p="[{'op': 'replace', 'path': '/spec/rules/0/host', 'value':'$appHost'}]"
# Also patch TLS if present
kubectl patch ingress hello-ingress -n $env:PROJECT_NAME --type='json' -p="[{'op': 'replace', 'path': '/spec/tls/0/hosts/0', 'value':'$appHost'}]"

# 7. Access Information
Write-Host ""
$appUrl = "http://$($appHost):8080"
Write-Box "ACCESS ENDPOINTS (Namespace: $env:PROJECT_NAME)" @(
    "Application: $appUrl",
    "Grafana:     http://grafana.dev.local:8080 (admin/admin123)",
    "Prometheus:  http://prometheus.dev.local:8080 (if enabled)"
)

Write-Box "DATABASE INFO (PostgreSQL)" @(
    "Host:     postgres-postgresql.infra.svc.cluster.local",
    "Port:     5432",
    "User:     postgres",
    "DB Name:  postgres",
    "Password: [Check secret.yaml or env]"
)

Write-Box "DEBUGGING & LOGS" @(
    "App Logs:   kubectl logs -n $env:PROJECT_NAME -l app=hello-app -f",
    "Failure Sim: powershell ./simulate-failure.ps1",
    "Stop Lab:    powershell ./stop-lab.ps1",
    "Live View:   k9s -n $env:PROJECT_NAME (Recommended)",
    "Test Lab:    powershell ./check-lab.ps1 -HostHeader $appHost"
)

Write-Host "TIP: To work on this project with hot-reload, run:" -ForegroundColor Yellow
Write-Host "   skaffold dev --namespace $env:PROJECT_NAME" -ForegroundColor White
Write-Host "`n=== SETUP COMPLETED SUCCESSFULLY ===`n" -ForegroundColor Green

