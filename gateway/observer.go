package gateway

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type Observer struct {
	logger         *log.Logger
	activeSessions int64
	totalMessages  int64
	totalErrors    int64
	tunnel         *TunnelManager
}

func NewObserver() *Observer {
	return &Observer{
		logger: log.New(os.Stdout, "", 0),
	}
}

func (o *Observer) SetTunnel(tm *TunnelManager) {
	o.tunnel = tm
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

func (o *Observer) Info(msg string, sessionID, userID string) {
	o.log("INFO", msg, sessionID, userID)
}

func (o *Observer) Error(err error, sessionID, userID string) {
	atomic.AddInt64(&o.totalErrors, 1)
	o.log("ERROR", err.Error(), sessionID, userID)
}

func (o *Observer) log(level, msg, sessionID, userID string) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   msg,
		SessionID: sessionID,
		UserID:    userID,
	}
	blob, _ := json.Marshal(entry)
	o.logger.Println(string(blob))
}

func (o *Observer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	health := map[string]interface{}{
		"status": "UP",
		"tunnel_status": string(TunnelOff),
		"tunnel_url": "",
	}
	if o.tunnel != nil {
		health["tunnel_status"] = string(o.tunnel.TunnelStatus())
		health["tunnel_url"] = o.tunnel.TunnelURL()
	}
	json.NewEncoder(w).Encode(health)
}

func (o *Observer) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_sessions": atomic.LoadInt64(&o.activeSessions),
		"total_messages":  atomic.LoadInt64(&o.totalMessages),
		"total_errors":    atomic.LoadInt64(&o.totalErrors),
	})
}

func (o *Observer) IncSessions() { atomic.AddInt64(&o.activeSessions, 1) }
func (o *Observer) DecSessions() { atomic.AddInt64(&o.activeSessions, -1) }
func (o *Observer) IncMessages() { atomic.AddInt64(&o.totalMessages, 1) }
