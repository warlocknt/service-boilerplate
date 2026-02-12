package app

import (
	"context"
	"testing"
	"time"

	"service-boilerplate/internal/config"
	"service-boilerplate/internal/logger"
	"service-boilerplate/internal/task"
)

// mockTask реализует task.Task для тестов
type mockTask struct {
	name       string
	startError error
	stopError  error
	started    bool
	stopped    bool
}

func (m *mockTask) Name() string {
	return m.name
}

func (m *mockTask) AfterStart(ctx context.Context) error {
	m.started = true
	return m.startError
}

func (m *mockTask) BeforeStop(ctx context.Context) error {
	m.stopped = true
	return m.stopError
}

// setupTestApp создает тестовое приложение
func setupTestApp(t *testing.T) (*App, *config.Config, *logger.Logger) {
	tmpDir := t.TempDir()
	log, err := logger.New("test-app", tmpDir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Name:        "test-service",
			DisplayName: "Test Service",
			Description: "Test Description",
			LogDir:      tmpDir,
		},
		Scheduler: config.SchedulerConfig{
			MaxPanicRestarts: 3,
			BackoffSeconds:   1,
		},
		Metrics: config.MetricsConfig{
			Enabled: false,
			Listen:  ":9090",
		},
	}

	return New(cfg, log), cfg, log
}

// TestNew создает новое приложение
func TestNew(t *testing.T) {
	app, _, log := setupTestApp(t)
	defer log.Close()

	if app == nil {
		t.Fatal("New() returned nil")
	}

	if app.config == nil {
		t.Error("app.config is nil")
	}

	if app.log == nil {
		t.Error("app.log is nil")
	}

	if app.lifecycle == nil {
		t.Error("app.lifecycle is nil")
	}

	if app.scheduler == nil {
		t.Error("app.scheduler is nil")
	}

	if app.metrics == nil {
		t.Error("app.metrics is nil")
	}
}

// TestGetScheduler возвращает планировщик
func TestGetScheduler(t *testing.T) {
	app, _, log := setupTestApp(t)
	defer log.Close()

	sched := app.GetScheduler()
	if sched == nil {
		t.Error("GetScheduler() returned nil")
	}
}

// TestRegisterTask регистрирует задачу
func TestRegisterTask(t *testing.T) {
	app, _, log := setupTestApp(t)
	defer log.Close()

	task1 := &mockTask{name: "test-task"}
	app.RegisterTask(task1)

	// Задача будет использована при запуске
}

// TestRun_StartsComponents запускает все компоненты
func TestRun_StartsComponents(t *testing.T) {
	app, _, log := setupTestApp(t)
	defer log.Close()

	// Регистрируем задачу для проверки lifecycle
	task1 := &mockTask{name: "lifecycle-task"}
	app.RegisterTask(task1)

	// Запускаем с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	// Ждем немного для запуска
	time.Sleep(200 * time.Millisecond)

	// Отменяем контекст для остановки
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Run() did not complete in time")
	}

	// Проверяем что задача была запущена и остановлена
	if !task1.started {
		t.Error("Task was not started")
	}
	if !task1.stopped {
		t.Error("Task was not stopped")
	}
}

// TestRun_WithTimer запускает с таймером
func TestRun_WithTimer(t *testing.T) {
	app, _, log := setupTestApp(t)
	defer log.Close()

	// Добавляем таймер
	executed := make(chan bool, 1)
	app.GetScheduler().AddTimer("test-timer", 100*time.Millisecond, func(ctx context.Context) {
		select {
		case executed <- true:
		default:
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	// Ждем выполнения таймера
	select {
	case <-executed:
		// Таймер выполнился
	case <-time.After(250 * time.Millisecond):
		t.Error("Timer was not executed")
	}

	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Error("Run() did not complete in time")
	}
}

// TestRun_MultipleTasks с несколькими задачами
func TestRun_MultipleTasks(t *testing.T) {
	app, _, log := setupTestApp(t)
	defer log.Close()

	tasks := []*mockTask{
		{name: "task1"},
		{name: "task2"},
		{name: "task3"},
	}

	for _, task := range tasks {
		app.RegisterTask(task)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Error("Run() did not complete in time")
	}

	// Проверяем что все задачи были запущены и остановлены
	for _, task := range tasks {
		if !task.started {
			t.Errorf("Task %s was not started", task.name)
		}
		if !task.stopped {
			t.Errorf("Task %s was not stopped", task.name)
		}
	}
}

// TestGracefulShutdown проверяет graceful shutdown
func TestGracefulShutdown(t *testing.T) {
	app, _, log := setupTestApp(t)
	defer log.Close()

	task1 := &mockTask{name: "graceful-task"}
	app.RegisterTask(task1)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	// Даем время на запуск
	time.Sleep(100 * time.Millisecond)

	// Graceful shutdown
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() error during shutdown = %v", err)
		}
		if !task1.stopped {
			t.Error("Task was not stopped during graceful shutdown")
		}
	case <-time.After(5 * time.Second):
		t.Error("Run() did not complete graceful shutdown in time")
	}
}

// TestRun_WithMetricsEnabled запуск с включенными метриками
func TestRun_WithMetricsEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	log, err := logger.New("test-app-metrics", tmpDir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Name:   "test-service",
			LogDir: tmpDir,
		},
		Scheduler: config.SchedulerConfig{
			MaxPanicRestarts: 3,
			BackoffSeconds:   1,
		},
		Metrics: config.MetricsConfig{
			Enabled: true,
			Listen:  ":0", // Случайный порт
		},
	}

	application := New(cfg, log)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- application.Run(ctx)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Error("Run() did not complete in time")
	}
}

// TestApp_ImplementsTaskInterface проверяет интерфейс (compile-time check)
func TestApp_ImplementsTaskInterface(t *testing.T) {
	// Этот тест проверяет что наши моки реализуют интерфейс
	var _ task.Task = &mockTask{}
}
