# Production-Grade Gateway: Architecture & Protocol

The ZeroExec Gateway is the secure entry point for all client interactions. It bridges browser-based xterm.js instances to the Windows-native terminal agent.

## 🏗️ High-Level Design

### Modular Separation
- **Agent**: Low-level Windows OS interaction (ConPTY, Job Objects).
- **Gateway**: High-level networking, security, and session orchestration.
- **Interface**: The Gateway interacts with the Agent via a clean API, ensuring no tight coupling between the transport layer and the OS guts.

### Concurrency & Streaming
- **Goroutine per Stream**: Each session utilizes two dedicated goroutines:
  - **Input Handler**: Reads JSON from WebSocket, validates/sanitizes, and writes to PTY Stdin.
  - **Output Streamer**: Reads from PTY Stdout, wraps in JSON, and sends via WebSocket.
- **Backpressure**: Uses buffered channels (1024 messages) for output. If the buffer fills, the oldest messages are dropped to prevent memory exhaustion (prioritizing system stability and real-time state).

## 📡 Message Protocol (JSON)

All communication between the Browser and Gateway uses structured JSON to ensure extensibility and error handling.

### Client → Gateway (Input)
```json
{
  "type": "input",
  "data": "H4sIAAA...", // Base64 encoded or raw string
  "metadata": {
    "resize": { "cols": 80, "rows": 24 }
  }
}
```

### Gateway → Client (Output / Events)
```json
{
  "type": "output",
  "data": "Result of command...",
  "timestamp": "2026-03-22T16:00:00Z"
}
```
```json
{
  "type": "error",
  "message": "Session timeout",
  "code": 4001
}
```

## 🔐 Security Reinforcements
- **Handshake Validation**: JWT is checked *before* WebSocket upgrade.
- **WSS Enforcement**: Rejects any non-TLS connections.
- **Strict Binding**: Gateway listens only on `127.0.0.1`.
- **Fail-Safe**: If the WebSocket closes, the Gateway immediately signals the Agent to terminate the specific shell process via its Job Object.
