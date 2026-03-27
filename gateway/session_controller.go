package gateway

import (
	"log"
	"sync"
	"time"
)

const (
	MaxSessionDuration = 1 * time.Hour
	IdleTimeout        = 10 * time.Minute
)

type SessionController struct {
	sessions map[string]*TerminalSession
	mu       sync.RWMutex
	sm       *SessionManager
}

func NewSessionController(sm *SessionManager) *SessionController {
	sc := &SessionController{
		sessions: make(map[string]*TerminalSession),
		sm:       sm,
	}
	go sc.monitorLoop()
	return sc
}

func (sc *SessionController) Register(ts *TerminalSession) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.sessions[ts.ID] = ts
	log.Printf("[SC] Registered Session: %s for User: %s", ts.ID, ts.UserID)
}

func (sc *SessionController) Unregister(id string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.sessions, id)
}

func (sc *SessionController) monitorLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sc.mu.Lock() // Write lock as we might update Status
		var toKill []string
		now := time.Now()

		for id, ts := range sc.sessions {
			age := now.Sub(ts.StartTime)
			idle := now.Sub(ts.LastActivity)

			// Update Status
			if idle > 2*time.Minute {
				ts.Status = "idle"
			} else {
				ts.Status = "active"
			}

			if age > MaxSessionDuration || idle > IdleTimeout {
				toKill = append(toKill, id)
			}
		}
		sc.mu.Unlock()

		for _, id := range toKill {
			log.Printf("[SC] Session Expired/Idle: %s", id)
			sc.mu.RLock()
			ts, ok := sc.sessions[id]
			sc.mu.RUnlock()
			if ok {
				ts.Conn.Close()
			}
		}
	}
}

func (sc *SessionController) ListSessions() []*TerminalSession {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	var list []*TerminalSession
	for _, s := range sc.sessions {
		list = append(list, s)
	}
	return list
}
