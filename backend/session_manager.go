package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string
	CreatedAt time.Time
	PTY       *PTY
	JobObject *JobObject
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

func (sm *SessionManager) CreateSession(command string, rows, cols uint16) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 1. Create Job Object for isolation
	jo, err := CreateJobObject("")
	if err != nil {
		return nil, fmt.Errorf("failed to create job object: %v", err)
	}

	// 2. Start PTY
	pty, err := StartPTY(command, rows, cols)
	if err != nil {
		jo.Close()
		return nil, fmt.Errorf("failed to start PTY: %v", err)
	}

	// 3. Assign process to Job Object
	if err := jo.AssignProcess(pty.process.Pid); err != nil {
		pty.Close()
		jo.Close()
		return nil, fmt.Errorf("failed to assign process to job object: %v", err)
	}

	session := &Session{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		PTY:       pty,
		JobObject: jo,
	}

	sm.sessions[session.ID] = session
	return session, nil
}

func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[id]
	return session, ok
}

func (sm *SessionManager) RemoveSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, ok := sm.sessions[id]; ok {
		session.PTY.Close()
		session.JobObject.Close()
		delete(sm.sessions, id)
	}
}

func (sm *SessionManager) ListSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []*Session
	for _, s := range sm.sessions {
		result = append(result, s)
	}
	return result
}
