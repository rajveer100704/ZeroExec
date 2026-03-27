package gateway

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (sm *SessionManager) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	user, err := sm.middleware.ValidateToken(r.URL.Query().Get("token"))
	if err != nil || user.Role != RoleAdmin {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	type SessionInfo struct {
		ID           string `json:"id"`
		UserID       string `json:"user_id"`
		Role         string `json:"role"`
		Status       string `json:"status"`
		StartTime    string `json:"start_time"`
		LastActivity string `json:"last_activity"`
		Duration     int    `json:"duration_sec"`
	}

	list := make([]SessionInfo, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		list = append(list, SessionInfo{
			ID:           s.ID,
			UserID:       s.UserID,
			Role:         s.Role,
			Status:       s.Status,
			StartTime:    s.StartTime.Format(time.RFC3339),
			LastActivity: s.LastActivity.Format(time.RFC3339),
			Duration:     int(time.Since(s.StartTime).Seconds()),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (sm *SessionManager) HandleTerminateSession(w http.ResponseWriter, r *http.Request) {
	user, err := sm.middleware.ValidateToken(r.URL.Query().Get("token"))
	if err != nil || user.Role != RoleAdmin {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing session id", http.StatusBadRequest)
		return
	}

	sm.mu.RLock()
	ts, ok := sm.sessions[id]
	sm.mu.RUnlock()

	if ok {
		ts.Conn.Close() // Trigger cleanup
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Session terminated"))
	} else {
		http.Error(w, "Session not found", http.StatusNotFound)
	}
}

func (sm *SessionManager) HandleListRecordings(w http.ResponseWriter, r *http.Request) {
	user, err := sm.middleware.ValidateToken(r.URL.Query().Get("token"))
	if err != nil || user.Role != RoleAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	entries, err := os.ReadDir(sm.Cfg.Audit.RecordDir)
	if err != nil {
		http.Error(w, "Failed to read recordings", http.StatusInternalServerError)
		return
	}

	recordings := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".vtr" {
			recordings = append(recordings, entry.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recordings)
}

func (sm *SessionManager) HandleGetRecording(w http.ResponseWriter, r *http.Request) {
	user, err := sm.middleware.ValidateToken(r.URL.Query().Get("token"))
	if err != nil || user.Role != RoleAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	path := filepath.Join(sm.Cfg.Audit.RecordDir, id)
	http.ServeFile(w, r, path)
}
