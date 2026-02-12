// Package metrics предоставляет Prometheus метрики
package metrics

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"service-boilerplate/internal/logger"
)

// Server предоставляет HTTP сервер для метрик
type Server struct {
	log       *logger.Logger
	server    *http.Server
	listener  net.Listener
	enabled   bool
	listen    string
	startTime time.Time
	registry  *prometheus.Registry

	// Метрики
	uptimeSeconds *prometheus.CounterVec
	timerRuns     *prometheus.CounterVec
	timerPanics   *prometheus.CounterVec
	activeTimers  prometheus.Gauge
}

// New создает новый metrics сервер
func New(log *logger.Logger, enabled bool, listen string) *Server {
	s := &Server{
		log:       log,
		enabled:   enabled,
		listen:    listen,
		startTime: time.Now(),
	}

	if enabled {
		// Создаем отдельный registry для избежания конфликтов в тестах
		s.registry = prometheus.NewRegistry()

		// Инициализируем метрики
		s.uptimeSeconds = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "service_uptime_seconds",
				Help: "Total service uptime in seconds",
			},
			[]string{},
		)

		s.timerRuns = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "timer_runs_total",
				Help: "Total number of timer executions",
			},
			[]string{"timer"},
		)

		s.timerPanics = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "timer_panics_total",
				Help: "Total number of timer panics",
			},
			[]string{"timer"},
		)

		s.activeTimers = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_timers",
				Help: "Number of active timers",
			},
		)

		// Регистрируем метрики в нашем registry
		s.registry.MustRegister(s.uptimeSeconds)
		s.registry.MustRegister(s.timerRuns)
		s.registry.MustRegister(s.timerPanics)
		s.registry.MustRegister(s.activeTimers)

		// Создаем HTTP сервер с нашим handler
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{}))
		mux.HandleFunc("/health", s.healthHandler)

		s.server = &http.Server{
			Handler: mux,
		}
	}

	return s
}

// GetAddress возвращает адрес сервера (полезно для тестов)
func (s *Server) GetAddress() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.listen
}

// healthHandler обрабатывает запросы /health
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

// Start запускает metrics сервер
func (s *Server) Start(ctx context.Context) error {
	if !s.enabled {
		s.log.Info("Metrics server is disabled")
		return nil
	}

	// Создаем listener чтобы получить реальный адрес (особенно важно для :0)
	listener, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}
	s.listener = listener

	s.log.Info("Starting metrics server", map[string]interface{}{"listen": s.GetAddress()})

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := s.server.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			s.log.Error("Metrics server error", map[string]interface{}{"error": err.Error()})
		}
	}()

	// Обновляем uptime
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.uptimeSeconds.WithLabelValues().Inc()
			}
		}
	}()

	return nil
}

// Stop останавливает metrics сервер
func (s *Server) Stop(ctx context.Context) error {
	if !s.enabled || s.server == nil {
		return nil
	}

	s.log.Info("Stopping metrics server")
	return s.server.Shutdown(ctx)
}

// RecordTimerRun записывает выполнение таймера
func (s *Server) RecordTimerRun(timerName string) {
	if s.enabled && s.timerRuns != nil {
		s.timerRuns.WithLabelValues(timerName).Inc()
	}
}

// RecordTimerPanic записывает panic таймера
func (s *Server) RecordTimerPanic(timerName string) {
	if s.enabled && s.timerPanics != nil {
		s.timerPanics.WithLabelValues(timerName).Inc()
	}
}

// SetActiveTimers устанавливает количество активных таймеров
func (s *Server) SetActiveTimers(count int32) {
	if s.enabled && s.activeTimers != nil {
		s.activeTimers.Set(float64(atomic.LoadInt32(&count)))
	}
}

// IncActiveTimers увеличивает счетчик активных таймеров
func (s *Server) IncActiveTimers() {
	if s.enabled && s.activeTimers != nil {
		s.activeTimers.Inc()
	}
}

// DecActiveTimers уменьшает счетчик активных таймеров
func (s *Server) DecActiveTimers() {
	if s.enabled && s.activeTimers != nil {
		s.activeTimers.Dec()
	}
}
