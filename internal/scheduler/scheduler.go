// Package scheduler предоставляет enterprise планировщик таймеров
package scheduler

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"service-boilerplate/internal/logger"
	"service-boilerplate/internal/metrics"
)

// Handler функция-обработчик таймера
type Handler func(ctx context.Context)

// Timer представляет один таймер
type Timer struct {
	name           string
	interval       time.Duration
	handler        Handler
	panicCount     int32
	maxRestarts    int
	backoffSeconds int
	running        int32
}

// Scheduler управляет таймерами
type Scheduler struct {
	mu             sync.RWMutex
	timers         map[string]*Timer
	log            *logger.Logger
	metrics        *metrics.Server
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	maxRestarts    int
	backoffSeconds int
	activeTimers   int32
}

// New создает новый планировщик
func New(log *logger.Logger, metricsServer *metrics.Server, maxRestarts, backoffSeconds int) *Scheduler {
	return &Scheduler{
		timers:         make(map[string]*Timer),
		log:            log,
		metrics:        metricsServer,
		maxRestarts:    maxRestarts,
		backoffSeconds: backoffSeconds,
	}
}

// AddTimer добавляет новый таймер
func (s *Scheduler) AddTimer(name string, interval time.Duration, handler Handler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.timers[name]; exists {
		return fmt.Errorf("timer %s already exists", name)
	}

	timer := &Timer{
		name:           name,
		interval:       interval,
		handler:        handler,
		maxRestarts:    s.maxRestarts,
		backoffSeconds: s.backoffSeconds,
	}

	s.timers[name] = timer
	s.log.Info("Timer added", map[string]interface{}{
		"name":     name,
		"interval": interval.String(),
	})

	return nil
}

// Start запускает все таймеры
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ctx != nil {
		return fmt.Errorf("scheduler already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)

	// Если нет таймеров, просто ждем отмены контекста
	if len(s.timers) == 0 {
		s.log.Info("No timers configured, scheduler running idle")
		return nil
	}

	// Запускаем каждый таймер в отдельной горутине
	for name, timer := range s.timers {
		s.wg.Add(1)
		atomic.AddInt32(&s.activeTimers, 1)
		if s.metrics != nil {
			s.metrics.IncActiveTimers()
		}
		go s.runTimer(name, timer)
	}

	s.log.Info("Scheduler started", map[string]interface{}{
		"timers_count": len(s.timers),
	})

	return nil
}

// runTimer выполняет таймер с защитой от panic
func (s *Scheduler) runTimer(name string, timer *Timer) {
	defer s.wg.Done()
	defer func() {
		atomic.AddInt32(&s.activeTimers, -1)
		if s.metrics != nil {
			s.metrics.DecActiveTimers()
		}
	}()

	s.log.Info("Timer started", map[string]interface{}{"timer": name})

	ticker := time.NewTicker(timer.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.log.Info("Timer stopped", map[string]interface{}{"timer": name})
			return
		case <-ticker.C:
			s.executeTimerWithRecovery(name, timer)
		}
	}
}

// executeTimerWithRecovery выполняет таймер с восстановлением после panic
func (s *Scheduler) executeTimerWithRecovery(name string, timer *Timer) {
	// Проверяем лимит перезапусков
	if timer.maxRestarts > 0 {
		panicCount := atomic.LoadInt32(&timer.panicCount)
		if int(panicCount) > timer.maxRestarts {
			s.log.Error("Timer exceeded max panic restarts, disabling", map[string]interface{}{
				"timer":        name,
				"panic_count":  panicCount,
				"max_restarts": timer.maxRestarts,
			})
			// Останавливаем этот таймер
			return
		}
	}

	// Выполняем с защитой от panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Увеличиваем счетчик panic
				newCount := atomic.AddInt32(&timer.panicCount, 1)

				// Логируем подробную информацию
				s.log.Error("Timer panic recovered", map[string]interface{}{
					"timer":       name,
					"panic":       r,
					"panic_count": newCount,
					"stacktrace":  string(debug.Stack()),
				})

				// Записываем метрику
				if s.metrics != nil {
					s.metrics.RecordTimerPanic(name)
				}

				// Backoff перед следующей попыткой
				if timer.backoffSeconds > 0 {
					time.Sleep(time.Duration(timer.backoffSeconds) * time.Second)
				}
			}
		}()

		// Записываем метрику выполнения
		if s.metrics != nil {
			s.metrics.RecordTimerRun(name)
		}

		// Выполняем обработчик
		timer.handler(s.ctx)
	}()
}

// Stop останавливает все таймеры
func (s *Scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	s.log.Info("Stopping scheduler...")

	// Ждем завершения всех таймеров с таймаутом
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.log.Info("All timers stopped gracefully")
	case <-ctx.Done():
		s.log.Warn("Timeout waiting for timers to stop")
	}

	return nil
}

// GetTimerCount возвращает количество таймеров
func (s *Scheduler) GetTimerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.timers)
}

// GetActiveTimerCount возвращает количество активных таймеров
func (s *Scheduler) GetActiveTimerCount() int32 {
	return atomic.LoadInt32(&s.activeTimers)
}
