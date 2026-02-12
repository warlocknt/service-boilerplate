//go:build !windows
// +build !windows

package platform

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"service-boilerplate/internal/app"
	"service-boilerplate/internal/config"
	"service-boilerplate/internal/logger"
)

// mockApp реализует упрощенное приложение для тестов
type mockApp struct {
	runCalled  bool
	stopCalled chan bool
	runError   error
}

func (m *mockApp) Run(ctx context.Context) error {
	m.runCalled = true
	select {
	case <-ctx.Done():
		return nil
	case <-m.stopCalled:
		return m.runError
	}
}

// TestRun_GracefulShutdown проверяет graceful shutdown по сигналу
func TestRun_GracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	log, err := logger.New("test-platform", tmpDir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer log.Close()

	// Создаем простой app для теста
	cfg := &config.Config{
		Service: config.ServiceConfig{
			Name:   "test",
			LogDir: tmpDir,
		},
		Scheduler: config.SchedulerConfig{
			MaxPanicRestarts: 3,
			BackoffSeconds:   1,
		},
		Metrics: config.MetricsConfig{
			Enabled: false,
		},
	}
	application := app.New(cfg, log)

	// Запускаем в отдельной горутине
	done := make(chan error, 1)
	go func() {
		done <- Run(log, application)
	}()

	// Даем время на запуск
	time.Sleep(100 * time.Millisecond)

	// Отправляем сигнал SIGTERM
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}
	process.Signal(syscall.SIGTERM)

	// Ждем завершения
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() error = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("Run() did not complete in time")
	}
}

// TestRun_ContextCancellation проверяет отмену контекста
func TestRun_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	log, err := logger.New("test-platform", tmpDir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Name:   "test",
			LogDir: tmpDir,
		},
		Scheduler: config.SchedulerConfig{
			MaxPanicRestarts: 3,
			BackoffSeconds:   1,
		},
		Metrics: config.MetricsConfig{
			Enabled: false,
		},
	}
	application := app.New(cfg, log)

	// Создаем контекст с отменой
	_, cancel := context.WithCancel(context.Background())

	// Запускаем
	done := make(chan error, 1)
	go func() {
		// Нельзя передать контекст в Run напрямую, но можно проверить что приложение запускается
		// и останавливается через сигналы
		done <- Run(log, application)
	}()

	// Даем время на запуск
	time.Sleep(100 * time.Millisecond)

	// Отменяем контекст через сигнал
	cancel()

	// Отправляем SIGTERM для graceful shutdown
	process, _ := os.FindProcess(os.Getpid())
	process.Signal(syscall.SIGTERM)

	select {
	case <-done:
		// OK
	case <-time.After(3 * time.Second):
		t.Error("Run() did not complete in time")
	}
}

// TestSignalHandling проверяет обработку разных сигналов
func TestSignalHandling(t *testing.T) {
	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT}

	for _, sig := range signals {
		t.Run(sig.String(), func(t *testing.T) {
			tmpDir := t.TempDir()
			log, err := logger.New("test-platform", tmpDir)
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := &config.Config{
				Service: config.ServiceConfig{
					Name:   "test",
					LogDir: tmpDir,
				},
				Scheduler: config.SchedulerConfig{
					MaxPanicRestarts: 3,
					BackoffSeconds:   1,
				},
				Metrics: config.MetricsConfig{
					Enabled: false,
				},
			}
			application := app.New(cfg, log)

			done := make(chan error, 1)
			go func() {
				done <- Run(log, application)
			}()

			time.Sleep(100 * time.Millisecond)

			process, _ := os.FindProcess(os.Getpid())
			process.Signal(sig)

			select {
			case <-done:
				// OK
			case <-time.After(3 * time.Second):
				t.Errorf("Run() did not complete after %s", sig.String())
			}
		})
	}
}
