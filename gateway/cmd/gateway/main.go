package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rajveer100704/ZeroExec/agent"
	"github.com/rajveer100704/ZeroExec/gateway"
)

type AgentAdapter struct {
	sm *agent.SessionManager
}

func (a *AgentAdapter) StartSession(command string, rows, cols uint16) (string, error) {
	s, err := a.sm.StartSession(command, rows, cols)
	if err != nil {
		return "", err
	}
	return s.ID, nil
}

func (a *AgentAdapter) WriteInput(id string, data []byte) error {
	s, ok := a.sm.GetSession(id)
	if !ok {
		return fmt.Errorf("session not found")
	}
	_, err := s.PTY.Write(data)
	return err
}

func (a *AgentAdapter) ReadOutput(id string) ([]byte, error) {
	s, ok := a.sm.GetSession(id)
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	buf := make([]byte, 8192)
	n, err := s.PTY.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (a *AgentAdapter) StopSession(id string) error {
	return a.sm.StopSession(id)
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	// 1. Load Configuration
	cfg, err := gateway.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("💎 AI Key Loaded: %v\n", cfg.AI.APIKey != "")

	// 2. Initialize Observability & Audit
	obs := gateway.NewObserver()
	audit, err := gateway.NewAuditManager(cfg.Audit.LogPath)
	if err != nil {
		log.Fatalf("Failed to init audit: %v", err)
	}
	defer audit.Close()

	// 3. Initialize Gateway Core
	mw, _ := gateway.NewMiddleware(cfg.Security.JWTSecret, audit)
	agentSm := agent.NewSessionManager()
	adapter := &AgentAdapter{sm: agentSm}
	sm := gateway.NewSessionManager(adapter, mw, obs, cfg)

	// 3b. Optionally start Cloudflare Tunnel
	if cfg.Tunnel.Enabled {
		tm := gateway.NewTunnelManager()
		obs.SetTunnel(tm)
		go tm.Start(cfg.Tunnel.CloudflaredPath, cfg.Server.Port)
		defer tm.Stop()
	}

	// 4. Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", sm.HandleWS)
	mux.HandleFunc("/health", obs.HandleHealth)
	mux.HandleFunc("/metrics", obs.HandleMetrics)
	mux.HandleFunc("/admin/sessions", sm.HandleListSessions)
	mux.HandleFunc("/admin/terminate", sm.HandleTerminateSession)
	mux.HandleFunc("/admin/recordings", sm.HandleListRecordings)
	mux.HandleFunc("/admin/recording", sm.HandleGetRecording)

	// AI Assistant
	aiHandler := gateway.NewAIHandler(cfg, mw)
	mux.HandleFunc("/ai/chat", aiHandler.HandleChat)

	// Dev endpoint to get token
	mux.HandleFunc("/auth/token", func(w http.ResponseWriter, r *http.Request) {
		token, _ := mw.GenerateDevToken()
		fmt.Fprintf(w, "%s", token)
	})

	// 5. Server Setup & TLS
	os.MkdirAll("certs", 0755)
	if err := gateway.EnsureCerts(cfg.Server.CertPath, cfg.Server.KeyPath); err != nil {
		log.Fatalf("Failed to ensure certs: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Addr, cfg.Server.Port)
	
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	server := &http.Server{
		Addr:    addr,
		Handler: corsMiddleware(mux),
	}

	// 6. Graceful Shutdown Logic
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		obs.Info("Gateway shutting down...", "", "")
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown: %v", err)
		}
		close(done)
	}()

	obs.Info(fmt.Sprintf("ZeroExec Gateway starting on http://%s", addr), "", "")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v", addr, err)
	}

	<-done
	obs.Info("Gateway stopped", "", "")
}
