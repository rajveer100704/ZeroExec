# 📦 ZeroExec Deployment Guide

This guide covers the production setup and configuration of the ZeroExec platform on Windows systems.

## 📋 Prerequisites
- Go 1.21+
- Node.js 18+
- Windows 10/11 or Windows Server 2019+ (Direct ConPTY support required)

## 🚀 Build & Installation

### 1. Automated Setup
The simplest way to deploy ZeroExec is using the provided PowerShell scripts:

```powershell
# 1. Compile all components and bundle frontend
.\scripts\build.ps1

# 2. Deploy to C:\ZeroExec and generate unique secrets
.\scripts\install.ps1
```

### 2. Physical Layout
After installation, your deployment directory will look like this:
- `zeroexec.exe`: Primary Gateway bridge.
- `zeroexec-agent.exe`: Native Windows Agent.
- `config.yaml`: Centralized configuration.
- `www/`: Static frontend assets.
- `logs/`: Application and Audit logs.
- `certs/`: TLS certificates (Auto-generated on first run).

## ⚙️ Configuration

ZeroExec uses a hierarchical configuration system (`config.yaml` < Environment Variables).

### Key Parameters
- `server.port` (VT_PORT): Port for the Gateway (Default: 8081).
- `jwt_secret` (VT_JWT_SECRET): Secret for token signing. **Must be changed in production.**
- `session.idle_timeout`: Session reaping time for inactive users (e.g., `10m`).
- `rate_limit.messages_per_sec`: Max messages from browser to terminal per second.

## 🔐 TLS Management
ZeroExec is **Secure-by-Default**.
- On first execution, the Gateway generates a self-signed RSA certificate in `certs/`.
- For production, replace `certs/cert.pem` and `certs/key.pem` with your organization's CA-signed certificates.

## 📊 Monitoring
Monitor system health and metrics via simple HTTP GET requests:
- Health: `https://localhost:8081/health` (includes `tunnel_status` and `tunnel_url`)
- Metrics: `https://localhost:8081/metrics`

## 🌐 Zero-Trust Remote Access (Feature 3)

Expose ZeroExec publicly with **zero firewall changes** using Cloudflare Tunnel.

### Setup
1. Download `cloudflared`: https://github.com/cloudflare/cloudflared/releases
2. Place `cloudflared.exe` in your `PATH` or set the full path in `config.yaml`.
3. Enable in `config.yaml`:

```yaml
tunnel:
  enabled: true
  cloudflared_path: "cloudflared"   # or full path like C:\tools\cloudflared.exe
```

4. Restart the Gateway. A `*.trycloudflare.com` URL will appear in the logs and in the **Remote Access** pill in the browser header.

### Security Guarantee
The tunnel is a **transparent pipe only** — JWT authentication and RBAC are still enforced at the Gateway layer. No credentials are stored with Cloudflare.

> **Note**: `trycloudflare.com` URLs are ephemeral (reset on restart). For persistent domains, configure a named Cloudflare Tunnel with `cloudflared tunnel login`.
