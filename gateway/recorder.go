package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type RecordEntry struct {
	Time float64 `json:"t"`
	Data string  `json:"d"`
}

type Recorder struct {
	file      *os.File
	startTime time.Time
	mu        sync.Mutex
}

func NewRecorder(dir string, sessionID string) (*Recorder, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	path := filepath.Join(dir, sessionID+".vtr") // ZeroExec Recording
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &Recorder{
		file:      f,
		startTime: time.Now(),
	}, nil
}

func (r *Recorder) Write(data string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry := RecordEntry{
		Time: time.Since(r.startTime).Seconds(),
		Data: data,
	}

	bytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = r.file.Write(append(bytes, '\n'))
	return err
}

func (r *Recorder) Close() {
	if r.file != nil {
		r.file.Close()
	}
}
