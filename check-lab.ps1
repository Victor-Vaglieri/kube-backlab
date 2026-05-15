param(
    [string]$ProjectPath = ".",
    [string]$HostHeader = "hello.dev.local",
    [string]$Endpoint = "http://localhost:8080"
)

function Write-Log {
    param([string]$Level, [string]$Message, [ConsoleColor]$Color)
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $Color
}

function Run-InfrastructureChecks {
    Write-Host "`n--- [ INFRASTRUCTURE VALIDATION ] ---" -ForegroundColor Cyan
    
    # 1. Gateway Check
    try {
        $tcp = New-Object System.Net.Sockets.TcpClient
        $tcp.Connect("localhost", 8080)
        Write-Log "SUCCESS" "Gateway (localhost:8080) is reachable." Green
        $tcp.Close()
    } catch {
        Write-Log "ERROR" "Gateway is down. Ensure k3d/skaffold is running." Red
        return $false
    }

    # 2. Ingress Routing Check
    $response = curl.exe -s -I -H "Host: $HostHeader" "$Endpoint/"
    if ($response -match "HTTP/1.1 200" -or $response -match "HTTP/1.1 302" -or $response -match "HTTP/1.1 308") {
        Write-Log "SUCCESS" "Ingress routing for $HostHeader is functional (Response: $($response | Select-Object -First 1))." Green
    } else {
        Write-Log "ERROR" "Ingress routing failed. Expected 200/302/308, but received:`n$response" Red
        return $false
    }
    return $true
}

function Run-ProjectTests {
    Write-Host "`n--- [ PROJECT TEST EXECUTION ] ---" -ForegroundColor Cyan
    
    $testPaths = @(
        (Join-Path $ProjectPath "tests"),
        (Join-Path $ProjectPath "src/test"),
        (Join-Path $ProjectPath "src/tests")
    )
    
    $srcPath = Join-Path $ProjectPath "src"
    
    # 1. Run Standard Package Tests if they exist
    if (Test-Path (Join-Path $srcPath "package.json")) {
        Write-Log "INFO" "Node.js project detected. Running npm test..." White
        npm test --prefix $srcPath
    }

    # 2. Run any custom .ps1 scripts found in test directories
    $foundCustom = $false
    foreach ($path in $testPaths) {
        if (Test-Path $path) {
            Get-ChildItem $path -Filter "*.ps1" | ForEach-Object { 
                $foundCustom = $true
                Write-Log "INFO" "Running custom test: $($_.Name)" White
                & $_.FullName
            }
        }
    }

    if (-not $foundCustom -and -not (Test-Path (Join-Path $srcPath "package.json"))) {
        Write-Log "WARN" "No tests found in standard locations." Yellow
    }
}

$infraOk = Run-InfrastructureChecks
if ($infraOk) {
    Run-ProjectTests
}

Write-Host "`n--- [ PROCESS COMPLETE ] ---`n" -ForegroundColor Cyan
