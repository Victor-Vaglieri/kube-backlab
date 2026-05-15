# Example Custom Test Script
$HostHeader = "hello.dev.local"
$Url = "http://localhost:8080/healthz/liveness"

Write-Host "   [CUSTOM] Probing $Url..." -ForegroundColor Gray
try {
    $resp = curl.exe -s -L -k -H "Host: $HostHeader" $Url
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
