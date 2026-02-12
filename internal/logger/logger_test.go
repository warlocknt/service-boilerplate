package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestNew_CreatesLogDir проверяет создание директории для логов
func TestNew_CreatesLogDir(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs", "subdir")

	logger, err := New("test-service", logDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// Проверяем что директория создана
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("Log directory was not created")
	}
}

// TestLogLevels проверяет разные уровни логирования
func TestLogLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")

	logger, err := New("test-service", logDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// Тестируем все уровни
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Flush логов
	logger.Flush()

	// Читаем лог файл
	logFile := filepath.Join(logDir, "test-service.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Debug не должен быть записан (по умолчанию InfoLevel)
	if strings.Contains(string(content), "debug message") {
		t.Error("Debug message should not be logged at InfoLevel")
	}

	// Остальные уровни должны быть записаны
	if !strings.Contains(string(content), "info message") {
		t.Error("Info message not found in log")
	}
	if !strings.Contains(string(content), "warn message") {
		t.Error("Warn message not found in log")
	}
	if !strings.Contains(string(content), "error message") {
		t.Error("Error message not found in log")
	}
}

// TestJSONFormat проверяет формат JSON в логах
func TestJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")

	logger, err := New("test-service", logDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// Пишем лог с полями
	logger.Info("test message", map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	})
	logger.Flush()

	// Читаем и парсим JSON
	logFile := filepath.Join(logDir, "test-service.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 {
		t.Fatal("No log lines found")
	}

	var entry LogEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	// Проверяем структуру
	if entry.Timestamp == "" {
		t.Error("Timestamp is empty")
	}
	if entry.Level != "info" {
		t.Errorf("Level = %v, want info", entry.Level)
	}
	if entry.Service != "test-service" {
		t.Errorf("Service = %v, want test-service", entry.Service)
	}
	if entry.Message != "test message" {
		t.Errorf("Message = %v, want test message", entry.Message)
	}
	if entry.Fields == nil {
		t.Error("Fields is nil")
	} else {
		if entry.Fields["key1"] != "value1" {
			t.Errorf("Fields[key1] = %v, want value1", entry.Fields["key1"])
		}
		if entry.Fields["key2"] != float64(123) {
			t.Errorf("Fields[key2] = %v, want 123", entry.Fields["key2"])
		}
	}
}

// TestConcurrentLogging проверяет потокобезопасность логгера
func TestConcurrentLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")

	logger, err := New("test-service", logDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// Пишем из нескольких горутин
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				logger.Info("concurrent log", map[string]interface{}{
					"goroutine": id,
					"iteration": j,
				})
			}
		}(i)
	}
	wg.Wait()
	logger.Flush()

	// Проверяем что все записи на месте
	logFile := filepath.Join(logDir, "test-service.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	expectedLines := 10 * 100
	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}
}

// TestSetLevel проверяет изменение уровня логирования
func TestSetLevel(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")

	logger, err := New("test-service", logDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// Меняем уровень на Debug
	logger.SetLevel(DebugLevel)
	logger.Debug("debug after set level")
	logger.Flush()

	// Проверяем что debug теперь пишется
	logFile := filepath.Join(logDir, "test-service.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "debug after set level") {
		t.Error("Debug message not found after SetLevel(DebugLevel)")
	}
}

// TestTimestampFormat проверяет формат timestamp
func TestTimestampFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")

	logger, err := New("test-service", logDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	before := time.Now().UTC()
	logger.Info("timestamp test")
	logger.Flush()
	after := time.Now().UTC()

	logFile := filepath.Join(logDir, "test-service.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Парсим timestamp
	logTime, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}

	// Проверяем что время в правильном диапазоне
	if logTime.Before(before) || logTime.After(after) {
		t.Error("Timestamp is out of expected range")
	}
}

// TestFlush проверяет сброс буферов
func TestFlush(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")

	logger, err := New("test-service", logDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	logger.Info("flush test")

	// Flush должен завершиться без ошибок
	if err := logger.Flush(); err != nil {
		t.Errorf("Flush() error = %v", err)
	}

	// Проверяем что данные записаны
	logFile := filepath.Join(logDir, "test-service.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "flush test") {
		t.Error("Log message not found after Flush()")
	}
}
