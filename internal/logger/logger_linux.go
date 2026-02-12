//go:build !windows
// +build !windows

// Package logger предоставляет кроссплатформенное логирование для Linux
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level представляет уровень логирования
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// Logger представляет структурированный JSON логгер
type Logger struct {
	mu      sync.RWMutex
	level   Level
	file    *os.File
	writer  io.Writer
	logDir  string
	service string
}

// LogEntry представляет одну запись в логе
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// New создает новый логгер
func New(serviceName, logDir string) (*Logger, error) {
	// Создаем директорию для логов
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Открываем файл для логирования
	logFile := filepath.Join(logDir, serviceName+".log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Создаем multiwriter для записи и в файл, и в stdout (для journald)
	writer := io.MultiWriter(file, os.Stdout)

	return &Logger{
		level:   InfoLevel,
		file:    file,
		writer:  writer,
		logDir:  logDir,
		service: serviceName,
	}, nil
}

// SetLevel устанавливает уровень логирования
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// log записывает сообщение в лог
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	writer := l.writer
	service := l.service
	l.mu.RUnlock()

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Service:   service,
		Message:   msg,
		Fields:    fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("failed to marshal log entry: %v", err)
		return
	}

	fmt.Fprintln(writer, string(data))
}

// Debug записывает debug сообщение
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DebugLevel, msg, f)
}

// Info записывает info сообщение
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(InfoLevel, msg, f)
}

// Warn записывает warn сообщение
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WarnLevel, msg, f)
}

// Error записывает error сообщение
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ErrorLevel, msg, f)
}

// Fatal записывает fatal сообщение и завершает программу
func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(FatalLevel, msg, f)
	l.Flush()
	os.Exit(1)
}

// Flush сбрасывает буферы логирования
func (l *Logger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Sync()
	}
	return nil
}

// Close закрывает логгер
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// RegisterEventSource регистрирует источник событий (только для Windows, на Linux no-op)
func RegisterEventSource(serviceName string) error {
	// На Linux не используется Windows Event Log
	return nil
}

// UnregisterEventSource удаляет источник событий (только для Windows, на Linux no-op)
func UnregisterEventSource(serviceName string) error {
	// На Linux не используется Windows Event Log
	return nil
}
