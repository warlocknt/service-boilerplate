//go:build !windows
// +build !windows

// Package platform предоставляет кроссплатформенную реализацию сервиса для Linux
package platform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"service-boilerplate/internal/app"
	"service-boilerplate/internal/logger"
)

// Run запускает сервис в Linux режиме
func Run(log *logger.Logger, application *app.App) error {
	log.Info("Starting service in Linux mode")

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Настраиваем обработку сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Запускаем приложение в отдельной горутине
	errChan := make(chan error, 1)
	go func() {
		errChan <- application.Run(ctx)
	}()

	// Ждем сигнала или ошибки
	select {
	case sig := <-sigChan:
		log.Info("Received signal, shutting down gracefully", map[string]interface{}{"signal": sig.String()})
		cancel()
		// Ждем завершения приложения
		if err := <-errChan; err != nil {
			return fmt.Errorf("application error during shutdown: %w", err)
		}
		return nil
	case err := <-errChan:
		return err
	}
}

// Start запускает systemd сервис
func Start(serviceName string) error {
	cmd := exec.Command("systemctl", "start", serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %w (output: %s)", err, string(output))
	}
	return nil
}

// Stop останавливает systemd сервис
func Stop(serviceName string) error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop service: %w (output: %s)", err, string(output))
	}
	return nil
}

// Install устанавливает systemd сервис
func Install(serviceName, displayName, description, execPath string) error {
	return fmt.Errorf("install on Linux: use scripts/install.sh instead")
}

// Uninstall удаляет systemd сервис
func Uninstall(serviceName string) error {
	return fmt.Errorf("uninstall on Linux: use scripts/uninstall.sh instead")
}
