package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"service-boilerplate/internal/app"
	"service-boilerplate/internal/config"
	"service-boilerplate/internal/logger"
	"service-boilerplate/internal/platform"
)

func main() {
	// Определяем путь к конфигу
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get executable path: %v\n", err)
		os.Exit(1)
	}
	execDir := filepath.Dir(execPath)
	configPath := filepath.Join(execDir, "configs", "config.yaml")

	// Загружаем конфигурацию
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем логгер
	log, err := logger.New(app.ServiceName, cfg.Service.LogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	// Создаем приложение
	application := app.New(cfg, log)

	// Добавляем таймеры согласно ТЗ
	// Таймер 1: каждые 5 секунд
	application.GetScheduler().AddTimer("every_5s", 5*time.Second, func(ctx context.Context) {
		log.Info("Timer executed: every_5s", map[string]interface{}{
			"timer": "every_5s",
		})
	})

	// Таймер 2: каждые 30 секунд
	application.GetScheduler().AddTimer("every_30s", 30*time.Second, func(ctx context.Context) {
		log.Info("Timer executed: every_30s", map[string]interface{}{
			"timer": "every_30s",
		})
	})

	// Таймер 3: каждые 15 минут
	application.GetScheduler().AddTimer("every_15m", 15*time.Minute, func(ctx context.Context) {
		log.Info("Timer executed: every_15m", map[string]interface{}{
			"timer": "every_15m",
		})
	})

	// Таймер 4: каждые 3 часа
	application.GetScheduler().AddTimer("every_3h", 3*time.Hour, func(ctx context.Context) {
		log.Info("Timer executed: every_3h", map[string]interface{}{
			"timer": "every_3h",
		})
	})

	// Определяем команду
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "run":
			// Запуск в консольном режиме
			log.Info("Running in console mode")
			if err := platform.Run(log, application); err != nil {
				log.Fatal("Application error", map[string]interface{}{"error": err.Error()})
			}
		case "install":
			// Установка Windows сервиса
			if err := installService(cfg, execPath); err != nil {
				log.Fatal("Failed to install service", map[string]interface{}{"error": err.Error()})
			}
			log.Info("Service installed successfully")
		case "uninstall":
			// Удаление Windows сервиса
			if err := uninstallService(cfg); err != nil {
				log.Fatal("Failed to uninstall service", map[string]interface{}{"error": err.Error()})
			}
			log.Info("Service uninstalled successfully")
		case "start":
			// Запуск Windows сервиса
			if err := platform.Start(app.ServiceName); err != nil {
				log.Fatal("Failed to start service", map[string]interface{}{"error": err.Error()})
			}
			log.Info("Service started successfully")
		case "stop":
			// Остановка Windows сервиса
			if err := platform.Stop(app.ServiceName); err != nil {
				log.Fatal("Failed to stop service", map[string]interface{}{"error": err.Error()})
			}
			log.Info("Service stopped successfully")
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
			fmt.Fprintf(os.Stderr, "Usage: %s [run|install|uninstall|start|stop]\n", os.Args[0])
			os.Exit(1)
		}
	} else {
		// По умолчанию запускаем как сервис
		if err := platform.Run(log, application); err != nil {
			log.Fatal("Application error", map[string]interface{}{"error": err.Error()})
		}
	}
}

// installService устанавливает Windows сервис
func installService(cfg *config.Config, execPath string) error {
	// Регистрируем источник событий
	if err := logger.RegisterEventSource(app.ServiceName); err != nil {
		return fmt.Errorf("failed to register event source: %w", err)
	}

	// Устанавливаем сервис
	if err := platform.Install(app.ServiceName, app.ServiceDisplayName, app.ServiceDescription, execPath); err != nil {
		logger.UnregisterEventSource(app.ServiceName)
		return err
	}

	return nil
}

// uninstallService удаляет Windows сервис
func uninstallService(cfg *config.Config) error {
	// Удаляем сервис
	if err := platform.Uninstall(app.ServiceName); err != nil {
		return err
	}

	// Удаляем источник событий
	if err := logger.UnregisterEventSource(app.ServiceName); err != nil {
		return fmt.Errorf("failed to unregister event source: %w", err)
	}

	return nil
}
