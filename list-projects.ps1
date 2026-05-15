function Write-Log {
    param([string]$Level, [string]$Message, [ConsoleColor]$Color)
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $Color
}

Write-Host "`n=== KUBE-BACKLAB ACTIVE PROJECTS ===" -ForegroundColor Cyan

$projectsRaw = kubectl get deployments -A --no-headers | Select-String "hello-app"
$projectsList = if ($projectsRaw) { 
    $projectsRaw | ForEach-Object { ($_.ToString() -split "\s+")[0] } | Sort-Object -Unique 
} else { 
    $null 
}

if ($projectsList) {
    Write-Log "INFO" "Found $($projectsList.Count) active project(s):" White
    Write-Host ""
    
    foreach ($pName in $projectsList) {
        if (-not $pName) { continue }
        $pods = kubectl get pods -n $pName --no-headers 2>$null
        $runningCount = ($pods | Select-String "Running").Count
        $totalCount = $pods.Count
        
        $pHost = if ($pName -eq "dev") { "hello.dev.local" } else { "$pName.dev.local" }
        $hColor = if ($runningCount -eq $totalCount) { "Green" } else { "Yellow" }
        
        Write-Host " Project: " -NoNewline
        Write-Host "$($pName.PadRight(15))" -ForegroundColor Cyan -NoNewline
        Write-Host "[$runningCount/$totalCount Pods] " -ForegroundColor $hColor -NoNewline
        Write-Host "-> http://$($pHost):8080" -ForegroundColor Gray
    }
} else {
    Write-Log "WARN" "No active projects found in the cluster." Yellow
}

Write-Host "`nTip: Use 'powershell ./start-lab.ps1 -Project <name>' to create a new one.`n" -ForegroundColor Gray
