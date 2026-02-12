# Service Boilerplate

Cross-platform production-ready service boilerplate на Go с поддержкой Windows Service и systemd.

## Возможности

- **Кроссплатформенность**: Windows Service + Linux systemd
- **Enterprise Scheduler**: Планировщик с таймерами, panic recovery, backoff
- **Метрики Prometheus**: `/metrics` и `/health` endpoints
- **Структурированное логирование**: JSON логи + Windows Event Log
- **Graceful shutdown**: Корректная остановка всех компонентов

## Требования

- Go 1.25+
- Без CGO

## Сборка и тестирование

### Быстрая сборка (автоматически запускает тесты)

**Windows (CMD/Batch):**
```cmd
# Полный pipeline: тесты + сборка
scripts\build.bat

# Доступные команды:
scripts\build.bat test         # Запустить тесты
scripts\build.bat test-fast    # Быстрые тесты
scripts\build.bat build        # Сборка (только после тестов)
scripts\build.bat build-only   # Сборка без тестов
scripts\build.bat build-all    # Сборка для всех платформ
scripts\build.bat build-win    # Сборка для Windows
scripts\build.bat build-linux  # Сборка для Linux
scripts\build.bat coverage     # Отчет о покрытии
scripts\build.bat check        # Проверка форматирования
scripts\build.bat clean        # Очистка
scripts\build.bat deps         # Загрузка зависимостей
scripts\build.bat ci           # Полный CI pipeline
```

**Linux/Mac:**
```bash
# Полный pipeline: тесты + сборка
make
# или
./scripts/build.sh

# Доступные команды:
make test          # Запустить тесты с race detector
make test-fast     # Быстрые тесты
make build         # Сборка (только после прохождения тестов)
make build-only    # Сборка без тестов
make build-all     # Сборка для всех платформ
make coverage      # Отчет о покрытии
make check         # Проверка форматирования и vet
make clean         # Очистка
make ci            # Полный CI pipeline
```

### Ручная сборка

```bash
# Загрузка зависимостей
go mod tidy

# Запуск тестов
go test -v -race -timeout 120s ./...

# Сборка (только если тесты прошли)
go build -o service-boilerplate ./cmd/service-boilerplate

# Сборка для Linux (с Windows)
GOOS=linux GOARCH=amd64 go build -o service-boilerplate-linux ./cmd/service-boilerplate

# Сборка для Windows (с Linux)
GOOS=windows GOARCH=amd64 go build -o service-boilerplate.exe ./cmd/service-boilerplate
```

## Конфигурация

Файл `configs/config.yaml`:

```yaml
service:
  name: service-boilerplate
  display_name: Service Boilerplate
  description: Cross-platform service boilerplate
  log_dir: ./logs

scheduler:
  max_panic_restarts: 5      # Максимум перезапусков после panic (0 = unlimited)
  backoff_seconds: 5         # Задержка перед перезапуском

metrics:
  enabled: true
  listen: ":9090"           # Адрес HTTP сервера метрик
```

## Windows

### Установка службы

Откройте **Командную строку от имени администратора** (cmd.exe):

```cmd
:: Установка службы
service-boilerplate.exe install

:: Запуск службы
service-boilerplate.exe start

:: Проверка статуса
sc query service-boilerplate
```

### Управление

```cmd
:: Остановка
service-boilerplate.exe stop

:: Удаление службы
service-boilerplate.exe uninstall

:: Запуск в консольном режиме (для отладки)
service-boilerplate.exe run
```

## Linux

### Установка systemd сервиса

```bash
# От имени root
sudo ./scripts/install.sh

# Или вручную:
sudo cp service-boilerplate /opt/service-boilerplate/
sudo cp configs/config.yaml /etc/service-boilerplate/configs/
sudo cp scripts/service.service /etc/systemd/system/service-boilerplate.service
sudo systemctl daemon-reload
```

### Управление

```bash
# Запуск
sudo systemctl start service-boilerplate

# Остановка
sudo systemctl stop service-boilerplate

# Автозапуск при загрузке
sudo systemctl enable service-boilerplate

# Просмотр логов
sudo journalctl -u service-boilerplate -f

# Статус
sudo systemctl status service-boilerplate
```

### Удаление

```bash
sudo ./scripts/uninstall.sh
```

## Метрики

При включенных метриках доступны endpoints:

- `http://localhost:9090/metrics` - Prometheus метрики
- `http://localhost:9090/health` - Health check

### Доступные метрики

- `service_uptime_seconds` - Время работы сервиса
- `timer_runs_total{timer="name"}` - Количество выполнений таймера
- `timer_panics_total{timer="name"}` - Количество panic в таймере
- `active_timers` - Количество активных таймеров

## Добавление таймера

В `cmd/service-boilerplate/main.go`:

```go
application.GetScheduler().AddTimer("my_timer", 1*time.Minute, func(ctx context.Context) {
    log.Info("My timer executed")
    // Ваша логика здесь
})
```

Таймеры уже добавлены:
- `every_5s` - каждые 5 секунд
- `every_30s` - каждые 30 секунд
- `every_15m` - каждые 15 минут
- `every_3h` - каждые 3 часа

## Добавление Task

Создайте структуру, реализующую интерфейс `task.Task`:

```go
type MyTask struct {
    log *logger.Logger
}

func (t *MyTask) Name() string {
    return "MyTask"
}

func (t *MyTask) AfterStart(ctx context.Context) error {
    t.log.Info("MyTask started")
    // Инициализация
    return nil
}

func (t *MyTask) BeforeStop(ctx context.Context) error {
    t.log.Info("MyTask stopping")
    // Очистка ресурсов
    return nil
}
```

Зарегистрируйте в `main.go`:

```go
myTask := &MyTask{log: log}
application.RegisterTask(myTask)
```

## Структура проекта

```
service-boilerplate/
├── cmd/service-boilerplate/
│   └── main.go              # Точка входа
├── internal/
│   ├── app/
│   │   └── app.go          # Основное приложение
│   ├── config/
│   │   └── config.go       # Загрузка конфигурации
│   ├── lifecycle/
│   │   └── lifecycle.go    # Управление lifecycle
│   ├── scheduler/
│   │   └── scheduler.go    # Планировщик таймеров
│   ├── logger/
│   │   ├── logger_linux.go # Логгер для Linux
│   │   └── logger_windows.go # Логгер для Windows
│   ├── metrics/
│   │   └── metrics.go      # Prometheus метрики
│   ├── platform/
│   │   ├── service_linux.go  # Linux сервис
│   │   └── service_windows.go # Windows сервис
│   └── task/
│       └── task.go         # Интерфейс Task
├── configs/
│   └── config.yaml         # Конфигурация
├── scripts/
│   ├── install.sh          # Установка Linux
│   ├── uninstall.sh        # Удаление Linux
│   └── service.service     # Systemd unit
├── go.mod
└── README.md
```

## Лицензия

MIT
