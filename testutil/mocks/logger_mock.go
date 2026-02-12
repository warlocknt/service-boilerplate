// Package mocks предоставляет моки для тестирования
package mocks

import (
	"sync"
)

// MockLogger мок логгера для тестов
type MockLogger struct {
	mu    sync.RWMutex
	logs  []LogEntry
	level int
}

// LogEntry представляет запись в логе
type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]interface{}
}

// NewMockLogger создает новый мок логгера
func NewMockLogger() *MockLogger {
	return &MockLogger{
		logs:  make([]LogEntry, 0),
		level: 1, // InfoLevel
	}
}

// Debug записывает debug сообщение
func (m *MockLogger) Debug(msg string, fields ...map[string]interface{}) {
	if m.level > 0 {
		return
	}
	m.log("debug", msg, fields)
}

// Info записывает info сообщение
func (m *MockLogger) Info(msg string, fields ...map[string]interface{}) {
	if m.level > 1 {
		return
	}
	m.log("info", msg, fields)
}

// Warn записывает warn сообщение
func (m *MockLogger) Warn(msg string, fields ...map[string]interface{}) {
	if m.level > 2 {
		return
	}
	m.log("warn", msg, fields)
}

// Error записывает error сообщение
func (m *MockLogger) Error(msg string, fields ...map[string]interface{}) {
	if m.level > 3 {
		return
	}
	m.log("error", msg, fields)
}

// Fatal записывает fatal сообщение (не вызывает os.Exit в моке)
func (m *MockLogger) Fatal(msg string, fields ...map[string]interface{}) {
	m.log("fatal", msg, fields)
}

// log внутренний метод для записи
func (m *MockLogger) log(level, msg string, fields []map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}

	m.logs = append(m.logs, LogEntry{
		Level:   level,
		Message: msg,
		Fields:  f,
	})
}

// GetLogs возвращает все записанные логи
func (m *MockLogger) GetLogs() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := make([]LogEntry, len(m.logs))
	copy(logs, m.logs)
	return logs
}

// HasLog проверяет наличие лога с заданным сообщением
func (m *MockLogger) HasLog(message string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, log := range m.logs {
		if log.Message == message {
			return true
		}
	}
	return false
}

// HasLogWithLevel проверяет наличие лога с заданным уровнем и сообщением
func (m *MockLogger) HasLogWithLevel(level, message string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, log := range m.logs {
		if log.Level == level && log.Message == message {
			return true
		}
	}
	return false
}

// SetLevel устанавливает уровень логирования
func (m *MockLogger) SetLevel(level int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.level = level
}

// Flush сбрасывает буферы (в моке ничего не делает)
func (m *MockLogger) Flush() error {
	return nil
}

// Close закрывает логгер (в моке ничего не делает)
func (m *MockLogger) Close() error {
	return nil
}

// Clear очищает все записанные логи
func (m *MockLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = make([]LogEntry, 0)
}
