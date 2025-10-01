package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AccountingPolicy   string `yaml:"accounting_policy"`
	AllowNegativeStock bool   `yaml:"allow_negative_stock"`
}

func LoadStockConfig(path string) (*Config, error) {
	cfg := &Config{
		AccountingPolicy:   "fifo",
		AllowNegativeStock: false,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("err while reading file")
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
