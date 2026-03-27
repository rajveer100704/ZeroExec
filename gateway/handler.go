package gateway

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Security at the listener level
	},
}

type SessionManager struct {
	agent      AgentInterface
	middleware *Middleware
	sessions   map[string]*TerminalSession
	mu         sync.RWMutex
	Controller *SessionController
	Obs        *Observer
	Cfg        *Config
}

type AgentInterface interface {
	StartSession(command string, rows, cols uint16) (string, error)
	WriteInput(sessionID string, data []byte) error
	ReadOutput(sessionID string) ([]byte, error)
	StopSession(sessionID string) error
}

type TerminalSession struct {
	ID            string
	Conn          *websocket.Conn
	OutputMsg     chan TerminalMessage
	Quit          chan struct{}
	UserID        string
	Role          string
	StartTime     time.Time
	LastActivity  time.Time
	Status        string // "active" or "idle"
	Recorder      *Recorder
	CommandBuffer string // Tracks the current line being typed
}

func NewSessionManager(agent AgentInterface, middleware *Middleware, obs *Observer, cfg *Config) *SessionManager {
	sm := &SessionManager{
		agent:      agent,
		middleware: middleware,
		sessions:   make(map[string]*TerminalSession),
		Obs:        obs,
		Cfg:        cfg,
	}
	sm.Controller = NewSessionController(sm)
	return sm
}

func (sm *SessionManager) HandleWS(w http.ResponseWriter, r *http.Request) {
	sm.Obs.IncSessions()
	defer sm.Obs.DecSessions()
	// 1. Auth & Identity
	user, err := sm.middleware.ValidateToken(r.URL.Query().Get("token"))
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// 2. Upgrade to WS
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 3. Rate Limiter & RBAC Engine
	rl := NewRateLimiter(20, 50) // 20 msg/sec
	rbac := &RBAC{}

	// 3. Create Session via Agent
	id, err := sm.agent.StartSession("powershell.exe", 24, 80)
	if err != nil {
		log.Printf("[WS] Session setup failed: %v", err)
		conn.WriteJSON(NewErrorMessage("Failed to start agent session"))
		return
	}
	defer sm.agent.StopSession(id)

	rec, _ := NewRecorder(sm.Cfg.Audit.RecordDir, id)
	ts := &TerminalSession{
		ID:           id,
		Conn:         conn,
		OutputMsg:    make(chan TerminalMessage, 1024),
		Quit:         make(chan struct{}),
		UserID:       user.ID,
		Role:         string(user.Role),
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		Status:       "active",
		Recorder:     rec,
	}

	sm.Controller.Register(ts)
	defer sm.Controller.Unregister(id)

	sm.mu.Lock()
	sm.sessions[id] = ts
	sm.mu.Unlock()

	defer func() {
		sm.mu.Lock()
		delete(sm.sessions, id)
		sm.mu.Unlock()
		if ts.Recorder != nil {
			ts.Recorder.Close()
		}
		close(ts.Quit)
	}()

	// 4. Heartbeat (Ping/Pong)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
					return
				}
			case <-ts.Quit:
				return
			}
		}
	}()

	// 5. Concurrency: Bridge
	var wg sync.WaitGroup
	wg.Add(2)

	// Agent -> Browser (Output)
	go func() {
		defer wg.Done()
		for {
			data, err := sm.agent.ReadOutput(id)
			if err != nil {
				return
			}
			msg := NewOutputMessage(string(data))
			msg.Version = ProtocolVersion
			sm.middleware.LogIO(id, "OUT", string(data))
			if ts.Recorder != nil {
				ts.Recorder.Write(string(data))
			}
			
			select {
			case ts.OutputMsg <- msg:
			case <-ts.Quit:
				return
			default:
			}
		}
	}()

	// Channel -> Browser
	go func() {
		defer wg.Done()
		for {
			select {
			case msg := <-ts.OutputMsg:
				if err := conn.WriteJSON(msg); err != nil {
					return
				}
			case <-ts.Quit:
				return
			}
		}
	}()

	// Browser -> Agent (Input) with Command RBAC Interception
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Security: Rate Limit
		if !rl.Allow() {
			conn.WriteJSON(NewErrorMessage("Rate limit exceeded"))
			continue
		}

		var msg TerminalMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Printf("[DEBUG] JSON Unmarshal failed for payload %q: %v", string(payload), err)
			continue
		}

		// Update activity
		ts.LastActivity = time.Now()
		ts.Status = "active"

		if msg.Type == MessageTypeInput {
			// Feature-level RBAC: Viewers cannot write at all
			if !rbac.CanPerform(user.Role, PermTerminalWrite) {
				conn.WriteJSON(NewErrorMessage("\r\n\x1b[91m🚫 Access Denied: Read-only session (Viewer)\x1b[0m\r\n"))
				continue
			}

			// Command-level RBAC: Buffer input and validate on Enter
			for _, ch := range msg.Data {
				switch ch {
				case '\r', '\n':
					// Enter pressed: validate the buffered command
					cmd := strings.TrimSpace(ts.CommandBuffer)
					ts.CommandBuffer = ""

					if cmd != "" && !rbac.IsCommandAllowed(user.Role, cmd) {
						// Send Ctrl+C to cancel the pending line in the PTY
						sm.agent.WriteInput(id, []byte("\x03"))
						// Notify the user
						deniedMsg := NewErrorMessage(
							"\r\n\x1b[91m🚫 Command Blocked: '" + cmd + "' is not permitted for role [" + string(user.Role) + "]\x1b[0m\r\n")
						select {
						case ts.OutputMsg <- deniedMsg:
						default:
						}
						sm.Obs.Info("RBAC_DENIED: "+cmd, id, user.ID)
						continue
					}

					// Command allowed: forward Enter
					sm.agent.WriteInput(id, []byte(string(ch)))

				case '\x7f', '\x08':
					// Backspace: trim buffer
					if len(ts.CommandBuffer) > 0 {
						ts.CommandBuffer = ts.CommandBuffer[:len(ts.CommandBuffer)-1]
					}
					sm.agent.WriteInput(id, []byte(string(ch)))

				case '\x03':
					// Ctrl+C: clear buffer and forward
					ts.CommandBuffer = ""
					sm.agent.WriteInput(id, []byte(string(ch)))

				case '\x15':
					// Ctrl+U: clear line buffer
					ts.CommandBuffer = ""
					sm.agent.WriteInput(id, []byte(string(ch)))

				default:
					// Regular character: buffer and forward
					ts.CommandBuffer += string(ch)
					log.Printf("[DEBUG] Writing to agent: %q", string(ch))
					sm.agent.WriteInput(id, []byte(string(ch)))
				}
			}

			sm.middleware.LogIO(id, "IN ", msg.Data)
		}
	}

	wg.Wait()
}
