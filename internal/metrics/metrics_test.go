package metrics

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"service-boilerplate/internal/logger"
)

// setupTestMetrics создает тестовый metrics server
func setupTestMetrics(t *testing.T, enabled bool) (*Server, *logger.Logger) {
	tmpDir := t.TempDir()
	log, err := logger.New("test-metrics", tmpDir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	server := New(log, enabled, "127.0.0.1:0") // :0 для случайного порта
	return server, log
}

// waitForServer ожидает готовности сервера
func waitForServer(t *testing.T, addr string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://" + addr + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("Server did not become ready in time")
}

// TestNew_Disabled проверяет создание отключенного сервера
func TestNew_Disabled(t *testing.T) {
	server, log := setupTestMetrics(t, false)
	defer log.Close()

	if server.enabled {
		t.Error("Expected metrics to be disabled")
	}

	if server.server != nil {
		t.Error("Server should be nil when disabled")
	}
}

// TestNew_Enabled проверяет создание включенного сервера
func TestNew_Enabled(t *testing.T) {
	server, log := setupTestMetrics(t, true)
	defer log.Close()

	if !server.enabled {
		t.Error("Expected metrics to be enabled")
	}

	if server.server == nil {
		t.Error("Server should not be nil when enabled")
	}
}

// TestStartStop_Disabled проверяет запуск/остановку отключенного сервера
func TestStartStop_Disabled(t *testing.T) {
	server, log := setupTestMetrics(t, false)
	defer log.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Не должно быть ошибок при disabled
	if err := server.Start(ctx); err != nil {
		t.Errorf("Start() with disabled error = %v", err)
	}

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Stop() with disabled error = %v", err)
	}
}

// TestHealthHandler проверяет endpoint /health
func TestHealthHandler(t *testing.T) {
	server, log := setupTestMetrics(t, true)
	defer log.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Ждем готовности сервера
	waitForServer(t, server.GetAddress(), 2*time.Second)

	// Делаем запрос к /health
	resp, err := http.Get("http://" + server.GetAddress() + "/health")
	if err != nil {
		t.Fatalf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health check status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	if !strings.Contains(string(body), "healthy") {
		t.Errorf("Health response doesn't contain 'healthy': %s", string(body))
	}

	server.Stop(ctx)
}

// TestMetricsEndpoint проверяет endpoint /metrics
func TestMetricsEndpoint(t *testing.T) {
	server, log := setupTestMetrics(t, true)
	defer log.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Ждем готовности сервера
	waitForServer(t, server.GetAddress(), 2*time.Second)

	// Делаем запрос к /metrics
	resp, err := http.Get("http://" + server.GetAddress() + "/metrics")
	if err != nil {
		t.Fatalf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Metrics endpoint status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	// Проверяем что есть Prometheus метрики
	if !strings.Contains(string(body), "# HELP") {
		t.Error("Metrics response doesn't contain Prometheus HELP")
	}

	server.Stop(ctx)
}

// TestRecordTimerRun проверяет запись выполнения таймера
func TestRecordTimerRun(t *testing.T) {
	server, log := setupTestMetrics(t, true)
	defer log.Close()

	// Записываем несколько выполнений
	server.RecordTimerRun("timer1")
	server.RecordTimerRun("timer1")
	server.RecordTimerRun("timer2")

	// Метрики должны быть записаны (проверка без ошибок)
}

// TestRecordTimerPanic проверяет запись panic таймера
func TestRecordTimerPanic(t *testing.T) {
	server, log := setupTestMetrics(t, true)
	defer log.Close()

	server.RecordTimerPanic("panic-timer")
	server.RecordTimerPanic("panic-timer")

	// Метрики должны быть записаны
}

// TestIncDecActiveTimers проверяет изменение счетчика активных таймеров
func TestIncDecActiveTimers(t *testing.T) {
	server, log := setupTestMetrics(t, true)
	defer log.Close()

	server.IncActiveTimers()
	server.IncActiveTimers()
	server.DecActiveTimers()

	// Счетчики должны измениться без ошибок
}

// TestSetActiveTimers проверяет установку количества таймеров
func TestSetActiveTimers(t *testing.T) {
	server, log := setupTestMetrics(t, true)
	defer log.Close()

	server.SetActiveTimers(5)
	server.SetActiveTimers(0)
	server.SetActiveTimers(100)

	// Значения должны быть установлены без ошибок
}

// TestRecordTimerRun_Disabled проверяет работу при disabled
func TestRecordTimerRun_Disabled(t *testing.T) {
	server, log := setupTestMetrics(t, false)
	defer log.Close()

	// Не должно быть panic при disabled
	server.RecordTimerRun("timer")
	server.RecordTimerPanic("timer")
	server.IncActiveTimers()
	server.DecActiveTimers()
	server.SetActiveTimers(5)
}

// TestUptimeMetric проверяет метрику uptime
func TestUptimeMetric(t *testing.T) {
	server, log := setupTestMetrics(t, true)
	defer log.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Ждем готовности сервера
	waitForServer(t, server.GetAddress(), 2*time.Second)

	// Ждем немного для накопления uptime
	time.Sleep(1500 * time.Millisecond)

	// Делаем запрос к /metrics
	resp, err := http.Get("http://" + server.GetAddress() + "/metrics")
	if err != nil {
		t.Fatalf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	// Проверяем наличие метрики uptime
	if !strings.Contains(string(body), "service_uptime_seconds") {
		t.Error("Uptime metric not found")
	}

	server.Stop(ctx)
}

// TestGracefulShutdown проверяет graceful shutdown
func TestGracefulShutdown(t *testing.T) {
	server, log := setupTestMetrics(t, true)
	defer log.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Ждем готовности сервера
	waitForServer(t, server.GetAddress(), 2*time.Second)

	// Проверяем что сервер работает
	resp, err := http.Get("http://" + server.GetAddress() + "/health")
	if err != nil {
		t.Fatalf("Server not running: %v", err)
	}
	resp.Body.Close()

	// Останавливаем сервер
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}
