# ZeroExec Installation Script

Write-Host "🛠️ Installing ZeroExec..." -ForegroundColor Cyan

$installPath = "C:\ZeroExec"
if (!(Test-Path $installPath)) {
    New-Item -ItemType Directory -Path $installPath | Out-Null
}

Write-Host "📁 Deploying files to $installPath..."
Copy-Item -Recurse dist\* $installPath -Force

Write-Host "🔑 Generating Development JWT Secret..."
$secret = [Guid]::NewGuid().ToString()
$configPath = Join-Path $installPath "config.yaml"
(Get-Content $configPath).Replace("developer-secret-change-in-production", $secret) | Set-Content $configPath

Write-Host "✅ Installation successful!" -ForegroundColor Green
Write-Host "Run with: cd $installPath; .\zeroexec.exe"
