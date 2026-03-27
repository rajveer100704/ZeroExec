package gateway

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port     int    `yaml:"port"`
		Addr     string `yaml:"addr"`
		CertPath string `yaml:"cert_path"`
		KeyPath  string `yaml:"key_path"`
	} `yaml:"server"`
	Security struct {
		JWTSecret   string        `yaml:"jwt_secret"`
		TokenExpiry time.Duration `yaml:"token_expiry"`
		MaxPayload  int           `yaml:"max_payload"`
	} `yaml:"security"`
	Session struct {
		IdleTimeout  time.Duration `yaml:"idle_timeout"`
		MaxDuration  time.Duration `yaml:"max_duration"`
		HeartbeatInt time.Duration `yaml:"heartbeat_interval"`
	} `yaml:"session"`
	RateLimit struct {
		Rate  float64 `yaml:"messages_per_sec"`
		Burst int     `yaml:"burst"`
	} `yaml:"rate_limit"`
	Audit struct {
		LogPath   string `yaml:"log_path"`
		RecordDir string `yaml:"record_dir"`
	} `yaml:"audit"`
	Tunnel struct {
		Enabled         bool   `yaml:"enabled"`
		CloudflaredPath string `yaml:"cloudflared_path"`
	} `yaml:"tunnel"`
	AI struct {
		APIKey string `yaml:"api_key"`
		Model  string `yaml:"model"`
	} `yaml:"ai"`
}

func LoadConfig(path string) (*Config, error) {
	// Attempt to load .env file; ignore error if it doesn't exist.
	_ = godotenv.Load(".env", "../.env")

	cfg := &Config{}
	
	// Set Defaults
	cfg.Server.Port = 8081
	cfg.Server.Addr = "127.0.0.1"
	cfg.Server.CertPath = "certs/cert.pem"
	cfg.Server.KeyPath = "certs/key.pem"
	cfg.Security.TokenExpiry = 15 * time.Minute
	cfg.Security.MaxPayload = 8192
	cfg.Session.IdleTimeout = 10 * time.Minute
	cfg.Session.MaxDuration = 1 * time.Hour
	cfg.Session.HeartbeatInt = 30 * time.Second
	cfg.RateLimit.Rate = 20
	cfg.RateLimit.Burst = 50
	cfg.Audit.LogPath = "gateway_audit.log"
	cfg.Audit.RecordDir = "recordings"
	cfg.Tunnel.Enabled = false
	cfg.Tunnel.CloudflaredPath = "cloudflared"
	cfg.AI.Model = "gemini-2.0-flash"

	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
			return nil, err
		}
	}

	// Environment Overrides
	if secret := os.Getenv("VT_JWT_SECRET"); secret != "" {
		cfg.Security.JWTSecret = secret
	}
	if aiKey := os.Getenv("VT_AI_KEY"); aiKey != "" {
		cfg.AI.APIKey = aiKey
	}
	if port := os.Getenv("VT_PORT"); port != "" {
		// Simplified for now
	}

	return cfg, nil
}
