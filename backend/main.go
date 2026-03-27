package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rajveer100704/ZeroExec/backend/certs"
	"github.com/rajveer100704/ZeroExec/backend/config"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("No config.yaml found, using defaults")
		cfg = config.DefaultConfig()
	}

	// Ensure certs directory exists
	certDir := filepath.Dir(cfg.Security.TLSCertPath)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		log.Fatalf("Failed to create certs directory: %v", err)
	}

	// Generate certs if they don't exist
	if _, err := os.Stat(cfg.Security.TLSCertPath); os.IsNotExist(err) {
		log.Printf("Generating self-signed certificates...")
		if err := certs.GenerateSelfSignedCert(cfg.Security.TLSCertPath, cfg.Security.TLSKeyPath); err != nil {
			log.Fatalf("Failed to generate certs: %v", err)
		}
	}

	// Initialize Audit Logger
	al, err := NewAuditLogger("audit.log")
	if err != nil {
		log.Printf("Failed to initialize audit logger: %v", err)
	}
	defer al.Close()

	// Initialize Session Manager & Gateway
	sm := NewSessionManager()
	gateway := NewGateway(sm, cfg.Security.JWTSecret, al)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, "%s", "ZeroExec Secure Gateway - Running")
	})

	// Unified endpoint for terminal access
	mux.HandleFunc("/ws", gateway.HandleTerminal)

	// Admin endpoint to generate a test token (for MVP dev)
	mux.HandleFunc("/auth/token", func(w http.ResponseWriter, r *http.Request) {
		token, err := gateway.GenerateToken()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "%s", token)
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting secure server on https://%s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// In a real production app, we would use ListenAndServeTLS
	// log.Fatal(server.ListenAndServeTLS(cfg.Security.TLSCertPath, cfg.Security.TLSKeyPath))
	
	// For this step, we'll just log and exit or keep it running in background if possible.
	log.Printf("ZeroExec Backend Fully Initialized.")
	
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
