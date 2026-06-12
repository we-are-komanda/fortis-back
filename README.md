# Fortis Backend

## Структура проекта

```
cmd/app/              — точка входа (main, router, DI)
internal/             — приватный код
├── config/           — чтение YAML + ENV
├── db/               — GORM + миграции
├── middleware/        — CORS, JWT, Swagger, Prometheus, Content-Type
├── metrics/          — Prometheus-метрики
├── probe/            — Liveness/Readiness/Startup probes
├── rdbms/            — GORM wrapper (executer, tx, pg)
└── modules/          — DDD-модули
    └── platform/     — пример: domain / application / infrastructure / ui
pkg/                  — публичные библиотеки
├── handlers/         — response, валидаторы
config/               — YAML-конфиги по средам (config.local.yml, config.dev.yml и т.д.)
migrations/           — SQL-миграции
tests/                — не-unit тесты
├── integration/      — //go:build integration (с БД)
└── e2e/              — //go:build e2e (HTTP, полный сценарий)
```

## Запуск

### Локально

```bash
make run              # go run ./cmd/app/
```

### Docker Compose

```bash
make up               # docker-compose up -d --build
make down             # docker-compose down --remove-orphans
```

Для локальной разработки можно создать `docker-compose.override.yml` (в .gitignore):
```yaml
services:
  app:
    environment:
      APP_DB_HOST: "postgres"
      APP_DB_PORT: "5432"
      APP_DB_USER: "postgres"
      APP_DB_PASSWORD: "postgres"
      APP_DB_DBNAME: "app"
volumes:
  postgres:
    name: fortis-postgres
```

## Конфигурация

Файлы в `config/`:
- `config.local.yml` — локальная разработка
- `config.dev.yml`, `config.test.yml`, `config.rc.yml`, `config.prod.yml` — среды

Выбор среды через переменную `ENVIRONMENT` (по умолчанию `local`).

Переопределение через переменные с префиксом `APP_`:
```bash
APP_DB_HOST=db APP_DB_PORT=5432 make run
```

Полный список переменных — в `.env.example`.

В DI подконфиги регистрируются отдельно:
- `config.Cors` → `middleware.NewCors`
- `config.Access` → `middleware.NewAccess`
- `config.Postgres` → `db.NewDataBase`

## Тестирование

### Unit-тесты

`*_test.go` в той же папке, что и тестируемый код. Используют `testing` + `testify`.

```bash
make test-unit        # go test -v -cover -tags=unit ./...
```

### Integration-тесты

В `tests/integration/` — с реальной БД (testcontainers или внешний хост).

```bash
go test -tags=integration ./tests/integration/...
```

### E2E-тесты

В `tests/e2e/` — полные HTTP-запросы к запущенному серверу.

```bash
go test -tags=e2e ./tests/e2e/...
```

## Роуты

| Путь | Описание |
|---|---|
| `/_/liveness` | Liveness probe |
| `/_/readiness` | Readiness probe |
| `/_/startup` | Startup probe |
| `/_/metrics` | Prometheus metrics |
| `/_/swagger` | Swagger JSON spec |
| `/_/swaggerui` | Swagger UI |
| `/api/v1/example` | Пример DDD-контроллера |

## Миграции

SQL-миграции в `migrations/`. Формат: `{timestamp}_{name}.up.sql` / `{timestamp}_{name}.down.sql`.

## Линтер

```bash
make lint-install        # установить golangci-lint v1.64.8
golangci-lint run        # запустить
```

Настройка — `.golangci.yml` (19 линтеров).  
Версия линтера зафиксирована: **golangci-lint v1.64.8** (совместимость с Go 1.24.3).  
Pre-commit хук запускает линтер автоматически (`.githooks/pre-commit`).

## Docker

```bash
docker build -f Dockerfile -t fortis-backend .
```

Multi-stage сборка: `golang:1.23.4-alpine` → `node:18-alpine` (swagger → OpenAPI 3) → `alpine:3.21`.
