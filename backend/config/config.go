package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Security struct {
		JWTSecret   string `yaml:"jwt_secret"`
		TLSCertPath string `yaml:"tls_cert_path"`
		TLSKeyPath  string `yaml:"tls_key_path"`
	} `yaml:"security"`
	Session struct {
		TimeoutMinutes int `yaml:"timeout_minutes"`
		MaxSessions    int `yaml:"max_sessions"`
	} `yaml:"session"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func DefaultConfig() *Config {
	cfg := &Config{}
	cfg.Server.Port = 8080
	cfg.Server.Host = "127.0.0.1"
	cfg.Security.TLSCertPath = "certs/cert.pem"
	cfg.Security.TLSKeyPath = "certs/key.pem"
	cfg.Session.TimeoutMinutes = 30
	cfg.Session.MaxSessions = 10
	return cfg
}
