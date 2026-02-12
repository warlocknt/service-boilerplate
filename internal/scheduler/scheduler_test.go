package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"service-boilerplate/internal/logger"
	"service-boilerplate/internal/metrics"
)

// setupTestScheduler создает тестовый scheduler
func setupTestScheduler(t *testing.T) (*Scheduler, *logger.Logger) {
	tmpDir := t.TempDir()
	log, err := logger.New("test-scheduler", tmpDir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	metricsServer := metrics.New(log, false, "")
	sched := New(log, metricsServer, 3, 0) // 3 max restarts, 0 backoff для скорости

	return sched, log
}

// TestAddTimer_Success проверяет успешное добавление таймера
func TestAddTimer_Success(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	err := sched.AddTimer("test-timer", time.Second, func(ctx context.Context) {})

	if err != nil {
		t.Errorf("AddTimer() error = %v", err)
	}

	if sched.GetTimerCount() != 1 {
		t.Errorf("Timer count = %d, want 1", sched.GetTimerCount())
	}
}

// TestAddTimer_DuplicateName проверяет ошибку при дублировании имени
func TestAddTimer_DuplicateName(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	err := sched.AddTimer("duplicate", time.Second, func(ctx context.Context) {})
	if err != nil {
		t.Fatalf("First AddTimer() error = %v", err)
	}

	err = sched.AddTimer("duplicate", time.Second, func(ctx context.Context) {})
	if err == nil {
		t.Error("AddTimer() expected error for duplicate name, got nil")
	}
}

// TestTimerExecution проверяет выполнение таймера
func TestTimerExecution(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	var counter int32
	err := sched.AddTimer("exec-timer", 50*time.Millisecond, func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	})
	if err != nil {
		t.Fatalf("AddTimer() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Ждем несколько выполнений
	time.Sleep(180 * time.Millisecond)

	sched.Stop(ctx)

	// Должно быть минимум 2 выполнения (50ms интервал за 180ms)
	count := atomic.LoadInt32(&counter)
	if count < 2 {
		t.Errorf("Timer executed %d times, expected at least 2", count)
	}
}

// TestPanicRecovery проверяет восстановление после panic
func TestPanicRecovery(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	// Ограничиваем количество restarts
	sched.maxRestarts = 2
	sched.backoffSeconds = 0

	var panicCount int32
	err := sched.AddTimer("panic-timer", 50*time.Millisecond, func(ctx context.Context) {
		atomic.AddInt32(&panicCount, 1)
		panic("test panic")
	})
	if err != nil {
		t.Fatalf("AddTimer() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Ждем несколько panic
	time.Sleep(250 * time.Millisecond)

	sched.Stop(ctx)

	// Должно быть несколько panic (restars + 1)
	count := atomic.LoadInt32(&panicCount)
	if count < 2 {
		t.Errorf("Panic count = %d, expected at least 2", count)
	}
}

// TestMaxRestartsExceeded проверяет отключение таймера после превышения лимита
func TestMaxRestartsExceeded(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	// Устанавливаем лимит в 2 restarts
	sched.maxRestarts = 2
	sched.backoffSeconds = 0

	var execCount int32
	err := sched.AddTimer("limited-timer", 50*time.Millisecond, func(ctx context.Context) {
		atomic.AddInt32(&execCount, 1)
		panic("test panic")
	})
	if err != nil {
		t.Fatalf("AddTimer() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Ждем достаточно времени (3 выполнения: первое + 2 restarts)
	time.Sleep(300 * time.Millisecond)

	sched.Stop(ctx)

	count := atomic.LoadInt32(&execCount)
	// Должно быть минимум 3 выполнения: первое + 2 restarts
	if count < 3 {
		t.Errorf("Execution count = %d, expected at least 3 (1 + 2 restarts)", count)
	}
}

// TestBackoff проверяет задержку перед перезапуском
func TestBackoff(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	sched.maxRestarts = 3
	sched.backoffSeconds = 1 // 1 секунда backoff

	startTimes := make([]time.Time, 0)
	var mu sync.Mutex

	err := sched.AddTimer("backoff-timer", 50*time.Millisecond, func(ctx context.Context) {
		mu.Lock()
		startTimes = append(startTimes, time.Now())
		mu.Unlock()
		panic("test panic")
	})
	if err != nil {
		t.Fatalf("AddTimer() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Ждем достаточно времени для выполнений с backoff
	time.Sleep(3500 * time.Millisecond)
	sched.Stop(ctx)

	mu.Lock()
	times := make([]time.Time, len(startTimes))
	copy(times, startTimes)
	mu.Unlock()

	// Должно быть минимум 2 выполнения для проверки backoff
	if len(times) < 2 {
		t.Fatalf("Expected at least 2 executions, got %d", len(times))
	}

	// Проверяем что была задержка между выполнениями (backoff + какое-то время)
	diff := times[1].Sub(times[0])
	if diff < 1*time.Second {
		t.Errorf("Backoff time = %v, expected at least 1s", diff)
	}

	// Проверяем общее время
	if time.Since(start) < 2*time.Second {
		t.Error("Total execution time too short, backoff not working")
	}
}

// TestGracefulStop проверяет graceful остановку
func TestGracefulStop(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	var running int32
	err := sched.AddTimer("stop-timer", 50*time.Millisecond, func(ctx context.Context) {
		atomic.AddInt32(&running, 1)
		time.Sleep(100 * time.Millisecond) // Долгая операция
		atomic.AddInt32(&running, -1)
	})
	if err != nil {
		t.Fatalf("AddTimer() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Даем таймеру запуститься
	time.Sleep(70 * time.Millisecond)

	// Останавливаем
	cancel()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer stopCancel()

	if err := sched.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Проверяем что все горутины остановились
	if atomic.LoadInt32(&running) != 0 {
		t.Error("Timer still running after Stop()")
	}
}

// TestConcurrentTimerExecution проверяет параллельное выполнение нескольких таймеров
func TestConcurrentTimerExecution(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	var counters [3]int32

	for i := 0; i < 3; i++ {
		idx := i
		err := sched.AddTimer(
			"timer-"+string(rune('A'+idx)),
			time.Duration(50+idx*20)*time.Millisecond,
			func(ctx context.Context) {
				atomic.AddInt32(&counters[idx], 1)
			},
		)
		if err != nil {
			t.Fatalf("AddTimer() error = %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	time.Sleep(250 * time.Millisecond)
	sched.Stop(ctx)

	// Все таймеры должны были выполниться несколько раз
	for i := 0; i < 3; i++ {
		count := atomic.LoadInt32(&counters[i])
		if count < 1 {
			t.Errorf("Timer %d executed %d times, expected at least 1", i, count)
		}
	}
}

// TestGetTimerCount проверяет получение количества таймеров
func TestGetTimerCount(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	if sched.GetTimerCount() != 0 {
		t.Errorf("Initial timer count = %d, want 0", sched.GetTimerCount())
	}

	for i := 0; i < 5; i++ {
		err := sched.AddTimer("timer-"+string(rune('0'+i)), time.Second, func(ctx context.Context) {})
		if err != nil {
			t.Fatalf("AddTimer() error = %v", err)
		}
	}

	if sched.GetTimerCount() != 5 {
		t.Errorf("Timer count after adding 5 = %d, want 5", sched.GetTimerCount())
	}
}

// TestGetActiveTimerCount проверяет получение количества активных таймеров
func TestGetActiveTimerCount(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	if sched.GetActiveTimerCount() != 0 {
		t.Errorf("Initial active timer count = %d, want 0", sched.GetActiveTimerCount())
	}

	for i := 0; i < 3; i++ {
		err := sched.AddTimer("active-timer-"+string(rune('0'+i)), time.Second, func(ctx context.Context) {})
		if err != nil {
			t.Fatalf("AddTimer() error = %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Даем время на запуск
	time.Sleep(50 * time.Millisecond)

	if sched.GetActiveTimerCount() != 3 {
		t.Errorf("Active timer count = %d, want 3", sched.GetActiveTimerCount())
	}

	sched.Stop(ctx)
}

// TestStart_AlreadyRunning проверяет ошибку при повторном запуске
func TestStart_AlreadyRunning(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	err := sched.AddTimer("test", time.Second, func(ctx context.Context) {})
	if err != nil {
		t.Fatalf("AddTimer() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("First Start() error = %v", err)
	}

	// Пытаемся запустить снова
	if err := sched.Start(ctx); err == nil {
		t.Error("Second Start() expected error, got nil")
	}

	sched.Stop(ctx)
}

// TestNoTimers проверяет работу без таймеров
func TestNoTimers(t *testing.T) {
	sched, log := setupTestScheduler(t)
	defer log.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() with no timers error = %v", err)
	}

	// Должен работать без таймеров
	time.Sleep(50 * time.Millisecond)

	if err := sched.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}
