package agent

import (
	"fmt"
	"sync"
	"time"
)

type Session struct {
	ID        string
	PTY       *PTY
	JobObject *JobObject
	CreatedAt time.Time
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) StartSession(command string, rows, cols uint16) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 1. Create Job Object
	jo, err := CreateJobObject("")
	if err != nil {
		return nil, err
	}

	// 2. Start PTY
	pty, err := StartPTY(command, rows, cols)
	if err != nil {
		jo.Close()
		return nil, err
	}

	// 3. Assign Process to Job Object
	if err := jo.AssignProcess(pty.process); err != nil {
		pty.Close()
		jo.Close()
		return nil, err
	}

	id := fmt.Sprintf("session-%d", time.Now().UnixNano())
	session := &Session{
		ID:        id,
		PTY:       pty,
		JobObject: jo,
		CreatedAt: time.Now(),
	}

	sm.sessions[id] = session
	return session, nil
}

func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.sessions[id]
	return s, ok
}

func (sm *SessionManager) StopSession(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[id]
	if !ok {
		return fmt.Errorf("session not found")
	}

	session.PTY.Close()
	session.JobObject.Close()
	delete(sm.sessions, id)
	return nil
}

func (sm *SessionManager) Cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for id, s := range sm.sessions {
		s.PTY.Close()
		s.JobObject.Close()
		delete(sm.sessions, id)
	}
}
