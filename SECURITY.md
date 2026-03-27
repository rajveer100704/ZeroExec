# 🛡️ ZeroExec Security Manifest

ZeroExec implements a multi-layered security architecture designed for high-integrity remote execution.

## 🧩 Defense Layers

### 1. Kernel Layer (Isolation)
ZeroExec does not rely on application-level process management. It uses **Windows Job Objects** to bind the lifecycle of all terminal sessions to the Agent process.
- **Guarantee**: If the parent process terminates (gracefully or via crash), the Windows kernel reaps every descendant process instantly.

### 2. Network Layer (Encryption)
- **TLS 1.2+**: All traffic is encrypted using modern cipher suites.
- **WSS**: WebSocket communication is upgraded from HTTPS, preserving origin and header security.

### 3. Identity Layer (RBAC)
ZeroExec uses RS256 signed JSON Web Tokens (JWT) with embedded role claims.
- **Enforcement**: RBAC checks are performed at the Gateway boundary for every single WebSocket message (Input, Resize, Admin actions).

### 4. Traffic Layer (Hardening)
- **Throttling**: Per-connection token-bucket rate limiting prevents automation-based abuse.
- **Validation**: Strict JSON schema enforcement ensures that malformed payloads are dropped before processing.

## 🔬 Threat Model & Mitigations

| Category | Threat | ZeroExec Mitigation |
| :--- | :--- | :--- |
| **Persistence** | Malicious long-running shell | Max Session Duration (1h) + Job Object |
| **Spoofing** | Forged identity | JWT Signature Validation |
| **DoS** | Buffer flooding | Rate Limiting + 8KB Payload Limit |
| **Leakage** | Unauthorized data access | Viewer/Operator Role Separation |
| **Resource** | Orphaned processes | `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` |

## ⚠️ Security Assumptions
- The `config.yaml` file is protected by OS-level file permissions.
- The machine running the Gateway has a trusted environment (no internal root-level adversaries).
- Users are responsible for rotated CA certificates in the `certs/` directory.
