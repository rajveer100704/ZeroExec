package main

import (
	"os"
	"time"
)

type AuditLogger struct {
	file *os.File
}

func NewAuditLogger(path string) (*AuditLogger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &AuditLogger{file: f}, nil
}

func (al *AuditLogger) LogIO(sessionID string, direction string, data []byte) {
	timestamp := time.Now().Format(time.RFC3339)
	entry := timestamp + " [" + sessionID + "] " + direction + ": " + string(data) + "\n"
	al.file.WriteString(entry)
}

func (al *AuditLogger) Close() error {
	return al.file.Close()
}
