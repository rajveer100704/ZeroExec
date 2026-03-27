package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Localhost only enforcement should be at the listener level
	},
}

type Gateway struct {
	sessionManager *SessionManager
	jwtSecret      []byte
	auditLogger    *AuditLogger
}

func NewGateway(sm *SessionManager, secret string, al *AuditLogger) *Gateway {
	return &Gateway{
		sessionManager: sm,
		jwtSecret:      []byte(secret),
		auditLogger:    al,
	}
}

func (g *Gateway) HandleTerminal(w http.ResponseWriter, r *http.Request) {
	// 1. Validate JWT from query param or header
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return g.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// 2. Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS Upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// 3. Create or get session (simplification: one session per token/connection for now)
	session, err := g.sessionManager.CreateSession("powershell.exe", 24, 80)
	if err != nil {
		log.Printf("Session creation error: %v", err)
		return
	}
	defer g.sessionManager.RemoveSession(session.ID)

	log.Printf("Terminal session started: %s", session.ID)

	// 4. Bridge WS <-> PTY
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// PTY -> WS
	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := session.PTY.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("PTY read error: %v", err)
				}
				cancel()
				return
			}
			if g.auditLogger != nil {
				g.auditLogger.LogIO(session.ID, "OUT", buf[:n])
			}
			if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				log.Printf("WS write error: %v", err)
				cancel()
				return
			}
		}
	}()

	// WS -> PTY
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if g.auditLogger != nil {
				g.auditLogger.LogIO(session.ID, "IN ", msg)
			}
			if _, err := session.PTY.Write(msg); err != nil {
				return
			}
		}
	}
}

func (g *Gateway) GenerateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"exp":        time.Now().Add(time.Hour * 1).Unix(),
	})
	return token.SignedString(g.jwtSecret)
}
