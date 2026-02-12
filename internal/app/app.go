// Package app объединяет все компоненты сервиса
package app

import (
	"context"
	"fmt"
	"time"

	"service-boilerplate/internal/config"
	"service-boilerplate/internal/lifecycle"
	"service-boilerplate/internal/logger"
	"service-boilerplate/internal/metrics"
	"service-boilerplate/internal/scheduler"
	"service-boilerplate/internal/task"
)

// App представляет основное приложение
type App struct {
	config    *config.Config
	log       *logger.Logger
	lifecycle *lifecycle.Manager
	scheduler *scheduler.Scheduler
	metrics   *metrics.Server
}

// New создает новое приложение
func New(cfg *config.Config, log *logger.Logger) *App {
	// Создаем сервер метрик
	metricsServer := metrics.New(log, cfg.Metrics.Enabled, cfg.Metrics.Listen)

	// Создаем планировщик
	sched := scheduler.New(log, metricsServer, cfg.Scheduler.MaxPanicRestarts, cfg.Scheduler.BackoffSeconds)

	// Создаем lifecycle менеджер
	lc := lifecycle.New(log)

	return &App{
		config:    cfg,
		log:       log,
		lifecycle: lc,
		scheduler: sched,
		metrics:   metricsServer,
	}
}

// GetScheduler возвращает планировщик для добавления таймеров
func (a *App) GetScheduler() *scheduler.Scheduler {
	return a.scheduler
}

// RegisterTask регистрирует задачу в lifecycle
func (a *App) RegisterTask(t task.Task) {
	a.lifecycle.Register(t)
}

// Run запускает приложение
func (a *App) Run(ctx context.Context) error {
	a.log.Info("Application starting", map[string]interface{}{
		"service": a.config.Service.Name,
		"version": "1.0.0",
	})

	// Запускаем все lifecycle задачи
	if err := a.lifecycle.StartAll(ctx); err != nil {
		return fmt.Errorf("failed to start lifecycle tasks: %w", err)
	}

	// Запускаем metrics сервер
	if err := a.metrics.Start(ctx); err != nil {
		return fmt.Errorf("failed to start metrics server: %w", err)
	}

	// Запускаем планировщик
	if err := a.scheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	a.log.Info("Application started successfully")

	// Ждем отмены контекста
	<-ctx.Done()

	a.log.Info("Application shutting down...")

	// Создаем контекст для graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем планировщик
	if err := a.scheduler.Stop(shutdownCtx); err != nil {
		a.log.Error("Error stopping scheduler", map[string]interface{}{"error": err.Error()})
	}

	// Останавливаем lifecycle задачи
	if err := a.lifecycle.StopAll(shutdownCtx); err != nil {
		a.log.Error("Error stopping lifecycle tasks", map[string]interface{}{"error": err.Error()})
	}

	// Останавливаем metrics сервер
	if err := a.metrics.Stop(shutdownCtx); err != nil {
		a.log.Error("Error stopping metrics server", map[string]interface{}{"error": err.Error()})
	}

	// Flush логов
	a.log.Info("Application stopped gracefully")
	a.log.Flush()

	return nil
}
