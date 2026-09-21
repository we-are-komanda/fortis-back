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

Если настройки сохранены в локальном `.env`, используйте `make run-dev`:
эта команда загружает `.env` перед запуском. `make run` сам файл не читает.
Для существующего `config.dev.yml` задайте `ENVIRONMENT=dev`,
`APP_WORK_MODE=webapp`, параметры `APP_DB_*` своей локальной базы и
`APP_AUTH_JWTSECRET`. Не добавляйте `.env` с секретом в Git.

Backend должен оставаться запущенным во время работы frontend. Проверка:
`http://localhost:8090/_/readiness`. При старте автоматически применяются
миграции; перед обновлением существующей локальной базы сохраните резервную копию.

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

## Первое предприятие: доверенная команда оператора (FRC03)

Оператор с локальным доступом к процессу и целевой БД создаёт предприятие и
membership для **уже зарегистрированного** пользователя одной командой.
Команда не поднимает HTTP-сервер и не использует пользовательский JWT.
Обычные API создания предприятия и управления membership остаются запрещены.

Из корня backend соберите существующий entrypoint:

```bash
go build -o /tmp/fortis-operator ./cmd/app
```

Выберите штатный `ENVIRONMENT` и передайте настройки **нужной** БД через
`APP_DB_HOST`, `APP_DB_PORT`, `APP_DB_USER`, `APP_DB_PASSWORD`, `APP_DB_DBNAME`
тем же способом, что для webapp. Бинарник сам `.env` не читает. Используются
тот же DI, `config.<environment>.yml` и автоматические миграции при старте;
параллельной конфигурации или отдельной SQL-процедуры нет.

Сохраните один новый UUID предприятия в записи операторского запроса и
используйте его во всех повторах. ID пользователя берётся из существующей
учётной записи. Передайте фактические название, адрес и GIS-центр объекта:

```bash
ENVIRONMENT=dev APP_WORK_MODE=enterprise-provision /tmp/fortis-operator \
  --enterprise-id "$ENTERPRISE_ID" \
  --user-id "$EXISTING_USER_ID" \
  --name 'Название предприятия' \
  --address 'Адрес предприятия' \
  --latitude 55.75 \
  --longitude 37.62
```

Все шесть флагов обязательны. ID — ненулевые UUID; название после trim —
1–512 Unicode-символов (предел поля предприятия в БД), адрес после trim не пуст.
Обе координаты задаются явно: latitude от −90 до 90, longitude от −180 до 180,
включая границы и ноль. `NaN`, бесконечность и пропущенное значение отклоняются.
Не используйте `(0, 0)` вместо неизвестного центра: это допустимая реальная точка,
а не значение по умолчанию. Первоначальный статус — `active`.

При успехе exit code — 0, stdout содержит только JSON
`{"enterpriseId":"…","userId":"…"}`. Название, адрес, координаты и данные
пользователя не попадают в SQL-лог этой операции; ошибки возвращаются без
исходного ввода/ошибок драйвера. Служебные сообщения среды и режима идут в stderr.
После успеха пользователь нажимает «Обновить» в workspace или входит заново:
предприятие доступно через обычный membership-filtered API, затем он создаёт
первую конфигурацию в UI.

Создание предприятия и единственной первоначальной связи выполняется в одной
транзакции. Нет пользователя или не записалась связь — предприятие не остаётся.
Повтор с тем же ID, теми же нормализованными данными и уже существующей связью
возвращает те же ID без изменения строк. Другое содержимое/получатель либо
отсутствующая связь при существующем предприятии — ненулевой exit code/conflict.
Команда не назначает пользователя существующему чужому объекту и не восстанавливает
отозванное membership. Совпадающее текущее состояние считается безопасным повтором;
отдельная история операторских запросов не вводится. Новый UUID означает новый
объект, поэтому при неопределённом результате повторяйте прежний запрос с прежним ID.

Названия пользовательских **конфигураций** имеют отдельное правило: trim и 1–120
Unicode-символов при создании, переименовании и импорте. Legacy schema-one импорт
без отдельного `name` берёт его из `projectName`. Чтение исторического снимка и
сохранение его неизменённого legacy-названия не выполняют очистку старых данных.

Проверка на одноразовом локальном PostgreSQL (без пользовательской БД):

```bash
FORTIS_FRC04_TEST_DSN='host=127.0.0.1 port=55434 user=fortis_test dbname=postgres sslmode=disable' \
  go test ./cmd/app -run 'TestFRC03Postgres' -count=1 -v
```

Эти тесты используют собственные временные схемы. Для полного browser→Next→Go→PG
сценария есть opt-in `TestFRC03BrowserServer`, который привязан только к
одноразовому кластеру `127.0.0.1:55434`, роли `fortis_test`, HTTP `127.0.0.1:8092`:

```bash
FRC03_BROWSER_SERVER=true FRC03_OPERATOR_ENV_FILE=/tmp/frc03-operator-env.json \
  go test ./cmd/app -run '^TestFRC03BrowserServer$' -count=1 -v -timeout=3h
```

Fixture создаёт отдельную БД `frc03_browser_<uuid>` штатными миграциями, две
синтетические учётные записи и ноль предприятий/membership/проектов.
`GET /__fixture` возвращает их ID/email и безопасные параметры тестовой БД;
env-файл содержит только синтетические `APP_DB_*` для того же operator binary.
Тестовые аккаунты используют `fixture-password-only`. `POST /__stop` завершает
fixture и удаляет только её БД; максимальное время работы — два часа.
Эти маршруты существуют только в `_test.go` и не входят в production binary.

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
