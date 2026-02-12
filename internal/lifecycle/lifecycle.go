// Package lifecycle управляет жизненным циклом компонентов
package lifecycle

import (
	"context"
	"fmt"
	"sync"

	"service-boilerplate/internal/logger"
	"service-boilerplate/internal/task"
)

// Manager управляет lifecycle компонентов
type Manager struct {
	mu    sync.RWMutex
	tasks []task.Task
	log   *logger.Logger
}

// New создает новый lifecycle менеджер
func New(log *logger.Logger) *Manager {
	return &Manager{
		tasks: make([]task.Task, 0),
		log:   log,
	}
}

// Register регистрирует новую задачу
func (m *Manager) Register(t task.Task) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks = append(m.tasks, t)
	m.log.Info("Task registered", map[string]interface{}{"task": t.Name()})
}

// StartAll запускает все зарегистрированные задачи
func (m *Manager) StartAll(ctx context.Context) error {
	m.mu.RLock()
	tasks := make([]task.Task, len(m.tasks))
	copy(tasks, m.tasks)
	m.mu.RUnlock()

	for _, t := range tasks {
		m.log.Info("Starting task", map[string]interface{}{"task": t.Name()})
		if err := t.AfterStart(ctx); err != nil {
			return fmt.Errorf("failed to start task %s: %w", t.Name(), err)
		}
	}

	return nil
}

// StopAll останавливает все задачи в обратном порядке
func (m *Manager) StopAll(ctx context.Context) error {
	m.mu.RLock()
	tasks := make([]task.Task, len(m.tasks))
	copy(tasks, m.tasks)
	m.mu.RUnlock()

	// Останавливаем в обратном порядке
	for i := len(tasks) - 1; i >= 0; i-- {
		t := tasks[i]
		m.log.Info("Stopping task", map[string]interface{}{"task": t.Name()})
		if err := t.BeforeStop(ctx); err != nil {
			m.log.Error("Error stopping task", map[string]interface{}{
				"task":  t.Name(),
				"error": err.Error(),
			})
		}
	}

	return nil
}
