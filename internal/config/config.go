package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type ModulesConfig struct {
	Audit     bool `yaml:"audit"`
	Files     bool `yaml:"files"`
	System    bool `yaml:"system"`
	Analytics bool `yaml:"analytics"`
	Reports   bool `yaml:"reports"`
	Stock     bool `yaml:"stock"`
	Users     bool `yaml:"users"`
	Shop      bool `yaml:"shop"`
}

type AuthConfig struct {
	JWTSecret     string `yaml:"jwt_secret"`
	TokenTTLHours int    `yaml:"token_ttl_hours"`
}

type AuditConfig struct {
	BatchSize            int `yaml:"batch_size"`
	FlushIntervalSeconds int `yaml:"flush_interval_seconds"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Modules  ModulesConfig  `yaml:"modules"`
	Auth     AuthConfig     `yaml:"auth"`
	Audit    AuditConfig    `yaml:"audit"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if envHost := os.Getenv("SERVER_HOST"); envHost != "" {
		cfg.Server.Host = envHost
	}

	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		if port, err := strconv.Atoi(envPort); err == nil {
			cfg.Server.Port = port
		}
	}

	if envDBDriver := os.Getenv("DB_DRIVER"); envDBDriver != "" {
		cfg.Database.Driver = envDBDriver
	}
	if envDBDSN := os.Getenv("DB_DSN"); envDBDSN != "" {
		cfg.Database.DSN = envDBDSN
	}

	return &cfg, nil
}
