// Package config предоставляет загрузку конфигурации из YAML
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config представляет конфигурацию сервиса
type Config struct {
	Service   ServiceConfig   `yaml:"service"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
	Metrics   MetricsConfig   `yaml:"metrics"`
}

// ServiceConfig содержит настройки сервиса
type ServiceConfig struct {
	LogDir string `yaml:"log_dir"`
}

// SchedulerConfig содержит настройки планировщика
type SchedulerConfig struct {
	MaxPanicRestarts int `yaml:"max_panic_restarts"`
	BackoffSeconds   int `yaml:"backoff_seconds"`
}

// MetricsConfig содержит настройки метрик
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

// Load загружает конфигурацию из YAML файла
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Устанавливаем значения по умолчанию
	if cfg.Service.LogDir == "" {
		cfg.Service.LogDir = "./logs"
	}
	if cfg.Scheduler.MaxPanicRestarts <= 0 {
		cfg.Scheduler.MaxPanicRestarts = 5
	}
	if cfg.Scheduler.BackoffSeconds <= 0 {
		cfg.Scheduler.BackoffSeconds = 5
	}
	if cfg.Metrics.Listen == "" {
		cfg.Metrics.Listen = ":9090"
	}

	return &cfg, nil
}
