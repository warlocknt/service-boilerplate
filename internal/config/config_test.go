package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad_Success проверяет успешную загрузку конфигурации из YAML
func TestLoad_Success(t *testing.T) {
	// Создаем временную директорию для теста
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Создаем тестовый конфиг
	configContent := `
service:
  name: test-service
  display_name: Test Service
  description: Test description
  log_dir: ./testlogs

scheduler:
  max_panic_restarts: 10
  backoff_seconds: 3

metrics:
  enabled: false
  listen: ":8080"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Загружаем конфиг
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Проверяем значения
	if cfg.Service.Name != "test-service" {
		t.Errorf("Service.Name = %v, want test-service", cfg.Service.Name)
	}
	if cfg.Service.DisplayName != "Test Service" {
		t.Errorf("Service.DisplayName = %v, want Test Service", cfg.Service.DisplayName)
	}
	if cfg.Service.Description != "Test description" {
		t.Errorf("Service.Description = %v, want Test description", cfg.Service.Description)
	}
	if cfg.Service.LogDir != "./testlogs" {
		t.Errorf("Service.LogDir = %v, want ./testlogs", cfg.Service.LogDir)
	}
	if cfg.Scheduler.MaxPanicRestarts != 10 {
		t.Errorf("Scheduler.MaxPanicRestarts = %v, want 10", cfg.Scheduler.MaxPanicRestarts)
	}
	if cfg.Scheduler.BackoffSeconds != 3 {
		t.Errorf("Scheduler.BackoffSeconds = %v, want 3", cfg.Scheduler.BackoffSeconds)
	}
	if cfg.Metrics.Enabled != false {
		t.Errorf("Metrics.Enabled = %v, want false", cfg.Metrics.Enabled)
	}
	if cfg.Metrics.Listen != ":8080" {
		t.Errorf("Metrics.Listen = %v, want :8080", cfg.Metrics.Listen)
	}
}

// TestLoad_DefaultValues проверяет установку значений по умолчанию
func TestLoad_DefaultValues(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Создаем минимальный конфиг
	configContent := `service: {}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Проверяем значения по умолчанию
	if cfg.Service.Name != "service-boilerplate" {
		t.Errorf("Service.Name default = %v, want service-boilerplate", cfg.Service.Name)
	}
	if cfg.Service.DisplayName != "Service Boilerplate" {
		t.Errorf("Service.DisplayName default = %v, want Service Boilerplate", cfg.Service.DisplayName)
	}
	if cfg.Service.Description != "Cross-platform service boilerplate" {
		t.Errorf("Service.Description default = %v, want Cross-platform service boilerplate", cfg.Service.Description)
	}
	if cfg.Service.LogDir != "./logs" {
		t.Errorf("Service.LogDir default = %v, want ./logs", cfg.Service.LogDir)
	}
	if cfg.Scheduler.MaxPanicRestarts != 5 {
		t.Errorf("Scheduler.MaxPanicRestarts default = %v, want 5", cfg.Scheduler.MaxPanicRestarts)
	}
	if cfg.Scheduler.BackoffSeconds != 5 {
		t.Errorf("Scheduler.BackoffSeconds default = %v, want 5", cfg.Scheduler.BackoffSeconds)
	}
	if cfg.Metrics.Listen != ":9090" {
		t.Errorf("Metrics.Listen default = %v, want :9090", cfg.Metrics.Listen)
	}
}

// TestLoad_FileNotFound проверяет ошибку при отсутствии файла
func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() expected error for non-existent file, got nil")
	}
}

// TestLoad_InvalidYAML проверяет ошибку при невалидном YAML
func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Создаем невалидный YAML
	configContent := `
service:
  name: [invalid yaml structure
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
}

// TestLoad_EmptyFile проверяет загрузку пустого файла
func TestLoad_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Создаем пустой файл
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Проверяем что применились дефолтные значения
	if cfg.Service.Name != "service-boilerplate" {
		t.Errorf("expected default name, got %v", cfg.Service.Name)
	}
}

// TestLoad_NegativeMaxRestarts проверяет обработку отрицательного значения
func TestLoad_NegativeMaxRestarts(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
scheduler:
  max_panic_restarts: -1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Отрицательное значение должно быть заменено на дефолтное
	if cfg.Scheduler.MaxPanicRestarts != 5 {
		t.Errorf("MaxPanicRestarts with negative value = %v, want 5", cfg.Scheduler.MaxPanicRestarts)
	}
}

// TestLoad_ZeroBackoff проверяет обработку нулевого backoff
func TestLoad_ZeroBackoff(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
scheduler:
  backoff_seconds: 0
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Нулевое значение должно быть заменено на дефолтное
	if cfg.Scheduler.BackoffSeconds != 5 {
		t.Errorf("BackoffSeconds with zero = %v, want 5", cfg.Scheduler.BackoffSeconds)
	}
}
