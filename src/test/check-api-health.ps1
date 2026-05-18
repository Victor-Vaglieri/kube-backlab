# Custom Test Script - Auto-detect Host
$ProjectName = $env:PROJECT_NAME
if ($null -eq $ProjectName -or $ProjectName -eq "") {
    $HostHeader = "hello.dev.local"
} else {
    $HostHeader = $ProjectName + ".dev.local"
}

$Url = "http://localhost:8080/healthz/liveness"

Write-Host "   [CUSTOM] Probing $Url with Host: $HostHeader..." -ForegroundColor Gray
try {
    $resp = curl.exe -s -L -k -H "Host: $HostHeader" $Url
    # Simplificando a verificação para evitar erros de parser
    if ($resp -match "OK") {
        Write-Host "   [SUCCESS] Application responded with 'OK'" -ForegroundColor Green
        exit 0
    } else {
        Write-Host "   [ERROR] Unexpected response: $resp" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "   [ERROR] Connection failed" -ForegroundColor Red
    exit 1
}
