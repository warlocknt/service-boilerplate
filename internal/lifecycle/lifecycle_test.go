package lifecycle

import (
	"context"
	"errors"
	"testing"
	"time"

	"service-boilerplate/internal/logger"
)

// mockTask реализует task.Task для тестов
type mockTask struct {
	name        string
	startError  error
	stopError   error
	started     bool
	stopped     bool
	startOrder  int
	stopOrder   int
	globalOrder *int
}

func (m *mockTask) Name() string {
	return m.name
}

func (m *mockTask) AfterStart(ctx context.Context) error {
	if m.globalOrder != nil {
		m.startOrder = *m.globalOrder
		*m.globalOrder++
	}
	m.started = true
	return m.startError
}

func (m *mockTask) BeforeStop(ctx context.Context) error {
	if m.globalOrder != nil {
		m.stopOrder = *m.globalOrder
		*m.globalOrder++
	}
	m.stopped = true
	return m.stopError
}

// setupTestManager создает тестовый lifecycle manager
func setupTestManager(t *testing.T) (*Manager, *logger.Logger) {
	tmpDir := t.TempDir()
	log, err := logger.New("test-lifecycle", tmpDir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	return New(log), log
}

// TestRegister проверяет регистрацию задачи
func TestRegister(t *testing.T) {
	manager, log := setupTestManager(t)
	defer log.Close()

	task1 := &mockTask{name: "task1"}
	manager.Register(task1)

	// Задача должна быть зарегистрирована
	// (внутренняя проверка через выполнение StartAll)
	ctx := context.Background()
	if err := manager.StartAll(ctx); err != nil {
		t.Errorf("StartAll() error = %v", err)
	}

	if !task1.started {
		t.Error("Task was not started after registration")
	}
}

// TestStartAll_Success проверяет успешный запуск всех задач
func TestStartAll_Success(t *testing.T) {
	manager, log := setupTestManager(t)
	defer log.Close()

	tasks := []*mockTask{
		{name: "task1"},
		{name: "task2"},
		{name: "task3"},
	}

	for _, task := range tasks {
		manager.Register(task)
	}

	ctx := context.Background()
	if err := manager.StartAll(ctx); err != nil {
		t.Fatalf("StartAll() error = %v", err)
	}

	// Все задачи должны быть запущены
	for _, task := range tasks {
		if !task.started {
			t.Errorf("Task %s was not started", task.name)
		}
	}
}

// TestStartAll_Error проверяет ошибку при запуске
func TestStartAll_Error(t *testing.T) {
	manager, log := setupTestManager(t)
	defer log.Close()

	task1 := &mockTask{name: "task1"}
	task2 := &mockTask{name: "task2", startError: errors.New("start failed")}
	task3 := &mockTask{name: "task3"}

	manager.Register(task1)
	manager.Register(task2)
	manager.Register(task3)

	ctx := context.Background()
	err := manager.StartAll(ctx)
	if err == nil {
		t.Error("StartAll() expected error, got nil")
	}

	// Первая задача должна быть запущена
	if !task1.started {
		t.Error("First task should be started")
	}

	// Третья задача не должна быть запущена (остановка после ошибки)
	if task3.started {
		t.Error("Third task should not be started after error")
	}
}

// TestStopAll_ReverseOrder проверяет остановку в обратном порядке
func TestStopAll_ReverseOrder(t *testing.T) {
	manager, log := setupTestManager(t)
	defer log.Close()

	order := 0
	tasks := []*mockTask{
		{name: "task1", globalOrder: &order},
		{name: "task2", globalOrder: &order},
		{name: "task3", globalOrder: &order},
	}

	for _, task := range tasks {
		manager.Register(task)
	}

	ctx := context.Background()
	if err := manager.StartAll(ctx); err != nil {
		t.Fatalf("StartAll() error = %v", err)
	}

	// Сбрасываем счетчик для проверки порядка остановки
	order = 0
	if err := manager.StopAll(ctx); err != nil {
		t.Errorf("StopAll() error = %v", err)
	}

	// Остановка должна быть в обратном порядке: task3, task2, task1
	if tasks[2].stopOrder != 0 {
		t.Errorf("Task3 stop order = %d, want 0", tasks[2].stopOrder)
	}
	if tasks[1].stopOrder != 1 {
		t.Errorf("Task2 stop order = %d, want 1", tasks[1].stopOrder)
	}
	if tasks[0].stopOrder != 2 {
		t.Errorf("Task1 stop order = %d, want 2", tasks[0].stopOrder)
	}
}

// TestStopAll_ContinuesOnError проверяет продолжение остановки при ошибке
func TestStopAll_ContinuesOnError(t *testing.T) {
	manager, log := setupTestManager(t)
	defer log.Close()

	task1 := &mockTask{name: "task1"}
	task2 := &mockTask{name: "task2", stopError: errors.New("stop failed")}
	task3 := &mockTask{name: "task3"}

	manager.Register(task1)
	manager.Register(task2)
	manager.Register(task3)

	ctx := context.Background()
	if err := manager.StartAll(ctx); err != nil {
		t.Fatalf("StartAll() error = %v", err)
	}

	if err := manager.StopAll(ctx); err != nil {
		t.Errorf("StopAll() error = %v", err)
	}

	// Все задачи должны быть остановлены (даже с ошибкой)
	if !task1.stopped {
		t.Error("Task1 was not stopped")
	}
	if !task2.stopped {
		t.Error("Task2 was not stopped")
	}
	if !task3.stopped {
		t.Error("Task3 was not stopped")
	}
}

// TestMultipleRegistrations проверяет регистрацию нескольких задач
func TestMultipleRegistrations(t *testing.T) {
	manager, log := setupTestManager(t)
	defer log.Close()

	// Регистрируем 10 задач
	for i := 0; i < 10; i++ {
		task := &mockTask{name: "task-" + string(rune('0'+i))}
		manager.Register(task)
	}

	ctx := context.Background()
	if err := manager.StartAll(ctx); err != nil {
		t.Fatalf("StartAll() error = %v", err)
	}

	if err := manager.StopAll(ctx); err != nil {
		t.Errorf("StopAll() error = %v", err)
	}
}

// TestContextCancellation проверяет передачу контекста
func TestContextCancellation(t *testing.T) {
	manager, log := setupTestManager(t)
	defer log.Close()

	taskWithTimeout := &mockTask{name: "timeout-task"}
	manager.Register(taskWithTimeout)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := manager.StartAll(ctx); err != nil {
		t.Fatalf("StartAll() error = %v", err)
	}

	if !taskWithTimeout.started {
		t.Error("Task was not started")
	}

	if err := manager.StopAll(ctx); err != nil {
		t.Errorf("StopAll() error = %v", err)
	}

	if !taskWithTimeout.stopped {
		t.Error("Task was not stopped")
	}
}

// TestEmptyManager проверяет работу с пустым менеджером
func TestEmptyManager(t *testing.T) {
	manager, log := setupTestManager(t)
	defer log.Close()

	ctx := context.Background()

	// StartAll и StopAll должны работать без задач
	if err := manager.StartAll(ctx); err != nil {
		t.Errorf("StartAll() with no tasks error = %v", err)
	}

	if err := manager.StopAll(ctx); err != nil {
		t.Errorf("StopAll() with no tasks error = %v", err)
	}
}

// TestConcurrentAccess проверяет потокобезопасность
func TestConcurrentAccess(t *testing.T) {
	manager, log := setupTestManager(t)
	defer log.Close()

	// Одновременно регистрируем задачи
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func(idx int) {
			task := &mockTask{name: "concurrent-task-" + string(rune('0'+idx))}
			manager.Register(task)
			done <- true
		}(i)
	}

	// Ждем завершения регистрации
	for i := 0; i < 3; i++ {
		<-done
	}

	ctx := context.Background()
	if err := manager.StartAll(ctx); err != nil {
		t.Fatalf("StartAll() error = %v", err)
	}

	if err := manager.StopAll(ctx); err != nil {
		t.Errorf("StopAll() error = %v", err)
	}
}
