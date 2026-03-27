package gateway

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sync"
)

type TunnelStatus string

const (
	TunnelPending TunnelStatus = "pending"
	TunnelActive  TunnelStatus = "active"
	TunnelError   TunnelStatus = "error"
	TunnelOff     TunnelStatus = "off"
)

// TunnelManager spawns cloudflared and extracts the assigned public URL.
type TunnelManager struct {
	mu     sync.RWMutex
	url    string
	status TunnelStatus
	cmd    *exec.Cmd
}

func NewTunnelManager() *TunnelManager {
	return &TunnelManager{status: TunnelOff}
}

// Start launches cloudflared and blocks until the URL is discovered or the process exits.
func (tm *TunnelManager) Start(cloudflaredPath string, port int) {
	tm.mu.Lock()
	tm.status = TunnelPending
	tm.mu.Unlock()

	portStr := fmt.Sprintf("http://localhost:%d", port)
	tm.cmd = exec.Command(cloudflaredPath, "tunnel", "--url", portStr)

	stderr, err := tm.cmd.StderrPipe()
	if err != nil {
		tm.setError()
		return
	}

	if err := tm.cmd.Start(); err != nil {
		log.Printf("[TUNNEL] Failed to start cloudflared: %v", err)
		tm.setError()
		return
	}

	// Parse cloudflared output for the tunnel URL
	urlRe := regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if match := urlRe.FindString(line); match != "" {
			tm.mu.Lock()
			tm.url = match
			tm.status = TunnelActive
			tm.mu.Unlock()
			log.Printf("[TUNNEL] Active: %s", match)
			break
		}
	}

	// If we exit the loop without finding a URL, it's an error
	tm.mu.RLock()
	isActive := tm.status == TunnelActive
	tm.mu.RUnlock()
	if !isActive {
		tm.setError()
	}
}

func (tm *TunnelManager) Stop() {
	if tm.cmd != nil && tm.cmd.Process != nil {
		tm.cmd.Process.Kill()
	}
}

func (tm *TunnelManager) TunnelURL() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.url
}

func (tm *TunnelManager) TunnelStatus() TunnelStatus {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.status
}

func (tm *TunnelManager) setError() {
	tm.mu.Lock()
	tm.status = TunnelError
	tm.mu.Unlock()
}
