# Petrochemical Data Platform

Платформа для сбора, хранения и анализа данных в нефтехимической отрасли на Go.

## Архитектура

- **Data Acquisition Layer**: Сбор данных из различных источников (REST API, MQTT, OPC UA, TCP/UDP)
- **Data Storage Layer**: PostgreSQL для метаданных, ClickHouse/TimescaleDB для временных рядов, Redis для кэширования
- **Business Logic Layer**: REST API с Gin, WebSocket для реального времени
- **Simulation Layer**: Генерация синтетических данных для тестирования

## Структура проекта

```
cmd/
├── api/           # Основное API приложение
├── parser/        # Сервис парсинга данных
└── simulator/     # Генератор синтетических данных

internal/
├── config/        # Конфигурация (Viper)
├── domain/        # Бизнес-сущности
├── handler/       # HTTP обработчики
├── service/       # Бизнес-логика
├── repository/    # Работа с БД
└── pkg/
    ├── parser/    # ETL-процессы
    └── simulator/ # Генерация данных

configs/           # Конфигурационные файлы
docker/            # Docker файлы
```

## Запуск

### Требования

- Go 1.21+
- PostgreSQL
- Redis
- MQTT брокер (опционально)

### Установка зависимостей

```bash
go mod tidy
```

### Конфигурация

Скопируйте `configs/config.yaml` и настройте подключения к БД.

### Запуск API сервера

```bash
go run cmd/api/main.go
```

### Запуск симулятора данных

```bash
go run cmd/simulator/main.go
```

## API Endpoints

- `GET /api/v1/assets` - Получение метаданных оборудования
- `GET /api/v1/telemetry/{sensor_id}` - Получение временных рядов
- `POST /api/v1/control` - Отправка управляющих команд
- `GET /ws` - WebSocket для данных в реальном времени

## Docker

```bash
docker-compose up -d
```

## Технологии

- **Go 1.21+**: Основной язык
- **Gin/Echo**: Веб-фреймворк
- **PostgreSQL + pgx**: Реляционная БД
- **ClickHouse**: БД для временных рядов
- **Redis**: Кэширование
- **MQTT**: Промышленные протоколы
- **Viper**: Конфигурация
- **Zap**: Логирование
- **Docker**: Контейнеризация