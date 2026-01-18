package config

import (
	"os"

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
	Files     bool `yaml:"files"`
	Analytics bool `yaml:"analytics"`
	Reports   bool `yaml:"reports"`
	Stock     bool `yaml:"stock"`
	Users     bool `yaml:"users"`
	Shop      bool `yaml:"shop"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Modules  ModulesConfig  `yaml:"modules"`
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
	return &cfg, nil
}
