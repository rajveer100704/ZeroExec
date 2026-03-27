# ZeroExec Build Script

Write-Output "Starting ZeroExec Production Build..."

$distDir = "dist"
if (Test-Path $distDir) { Remove-Item -Recurse -Force $distDir }
New-Item -ItemType Directory -Path $distDir | Out-Null

# 1. Build Gateway
Write-Output "Building Gateway..."
Set-Location gateway/cmd/gateway
go build -o ../../../dist/zeroexec.exe .
Set-Location ../../..

# 2. Build Agent
Write-Output "Building Agent..."
Set-Location agent/cmd/agent
go build -o ../../../dist/zeroexec-agent.exe .
Set-Location ../../..

# 3. Build Frontend
Write-Output "Building Frontend..."
Set-Location frontend
npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Output "Frontend build failed!"
    exit $LASTEXITCODE
}
Copy-Item -Recurse dist ../dist/www
Set-Location ..

# 4. Copy Config
Write-Output "Copying Configuration..."
Copy-Item config.yaml dist/config.yaml

Write-Output "Build Complete! artifacts in /dist"
