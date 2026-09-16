package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	App    AppConfig    `yaml:"app"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type AppConfig struct {
	WorkerCount int `yaml:"workers"`
	TopN        int `yaml:"top_n"`
	MaxFiles    int `yaml:"max_files"`
	MaxRetries  int `yaml:"max_retries"`
	RetryDelay  int `yaml:"retry_delay"`
}

func Load(path string) (*Config, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if v := os.Getenv("SERV_HOST"); v != "" {
		cfg.Server.Host = v
	}

	if v := os.Getenv("SERV_PORT"); v != "" {
		cfg.Server.Port = v
	}

	return &cfg, nil
}
