//go:build windows
// +build windows

// Package platform предоставляет реализацию Windows Service
package platform

import (
	"context"
	"fmt"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/mgr"

	"service-boilerplate/internal/app"
	"service-boilerplate/internal/logger"
)

// windowsService реализует интерфейс svc.Service
type windowsService struct {
	log     *logger.Logger
	app     *app.App
	ctx     context.Context
	cancel  context.CancelFunc
	errChan chan error
}

// Execute запускается Windows Service Control Manager
func (s *windowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// Создаем контекст для приложения
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.errChan = make(chan error, 1)

	// Запускаем приложение
	go func() {
		s.errChan <- s.app.Run(s.ctx)
	}()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	s.log.Info("Windows service started")

	// Основной цикл обработки команд
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s.log.Info("Received stop/shutdown command")
				changes <- svc.Status{State: svc.StopPending}
				s.cancel()
				// Ждем завершения приложения
				<-s.errChan
				changes <- svc.Status{State: svc.Stopped}
				return
			default:
				s.log.Error("Unexpected control request", map[string]interface{}{"cmd": c.Cmd})
			}
		case err := <-s.errChan:
			if err != nil {
				s.log.Error("Application error", map[string]interface{}{"error": err.Error()})
			}
			changes <- svc.Status{State: svc.Stopped}
			return
		}
	}
}

// Run запускает сервис как обычное приложение (для тестирования)
func Run(log *logger.Logger, application *app.App) error {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return fmt.Errorf("failed to determine if running as service: %w", err)
	}

	if isService {
		// Запускаем как Windows Service
		s := &windowsService{
			log: log,
			app: application,
		}
		return svc.Run("service-boilerplate", s)
	}

	// Запускаем как обычное приложение
	log.Info("Running in console mode")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return application.Run(ctx)
}

// Install устанавливает сервис в Windows
func Install(serviceName, displayName, description string, execPath string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}

	s, err = m.CreateService(serviceName, execPath, mgr.Config{
		DisplayName: displayName,
		Description: description,
		StartType:   mgr.StartAutomatic,
	})
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	return nil
}

// Uninstall удаляет сервис из Windows
func Uninstall(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s does not exist", serviceName)
	}
	defer s.Close()

	// Останавливаем сервис если он запущен
	status, err := s.Query()
	if err == nil && status.State != svc.Stopped {
		s.Control(svc.Stop)
	}

	return s.Delete()
}

// Start запускает установленный сервис
func Start(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s does not exist", serviceName)
	}
	defer s.Close()

	return s.Start()
}

// Stop останавливает запущенный сервис
func Stop(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s does not exist", serviceName)
	}
	defer s.Close()

	_, err = s.Control(svc.Stop)
	return err
}

// RunAsService запускает сервис через SCM (Service Control Manager)
func RunAsService(log *logger.Logger, application *app.App) error {
	s := &windowsService{
		log: log,
		app: application,
	}
	return svc.Run("service-boilerplate", s)
}

// Log сообщения для debug
var elog debug.Log
