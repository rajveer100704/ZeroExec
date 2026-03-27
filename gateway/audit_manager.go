package gateway

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type AuditManager struct {
	logger *log.Logger
	file   *os.File
	mu     sync.Mutex
	path   string
}

func NewAuditManager(path string) (*AuditManager, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &AuditManager{
		logger: log.New(f, "", 0),
		file:   f,
		path:   path,
	}, nil
}

func (a *AuditManager) Log(sessionID, direction, data string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Structured Audit Log
	entry := fmt.Sprintf("[%s] %s | %s | %s", 
		os.Getenv("HOSTNAME"), sessionID, direction, data)
	a.logger.Println(entry)
}

func (a *AuditManager) Close() {
	if a.file != nil {
		a.file.Close()
	}
}

func (a *AuditManager) ExportJSON(w io.Writer) error {
	// Simple export: read file and wrap in JSON
	// In production, this would parse the structured format
	f, err := os.Open(a.path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}
