// Package task предоставляет интерфейс Task для lifecycle
package task

import "context"

// Task определяет интерфейс для компонентов с lifecycle
type Task interface {
	// Name возвращает имя задачи
	Name() string
	// AfterStart вызывается после запуска сервиса
	AfterStart(ctx context.Context) error
	// BeforeStop вызывается перед остановкой сервиса
	BeforeStop(ctx context.Context) error
}
