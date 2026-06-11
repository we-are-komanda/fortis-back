# Golang Template
## Конфигурация

1. Замените `golang-template` на название вашего проекта во всех файлах проекта.

2. Настройте конфигурацию в [var/config](var/config) для каждой среды.

3. Если используется база данных, можно воспользоваться пакетом [db/db.go](db/db.go). Если база данных не используется, удалите этот файл.

4. Опишите ваш проект в [README.md](README.md) и обновите заголовок.

5. Удалите эту памятку.

6. Удалить примеры реализации и описание из [README.md](README.md)

## Конфигурация приложения через переменные окружения
 
Для этого необходимо прокинуть в контейнер параметры. Например что бы заполнить конфиг [config.prod.yml](var/config/config.prod.yml) нужно прокинуть переменные
```yml
APP_DB_USER=localhost
APP_DB_PASSWORD=postgres
APP_DB_HOST=db
APP_DB_PORT=5432
```

## Регистрация зависимостей

В файле [dependencies.go](./dependencies.go) вы можете зарегистрировать сервисы, репозитории и прочие объекты 
созданные в ваших пакетах.
```bash
err := app.container.Provide(api.NewExampleController)
```

## Маршрутизация

В файле [router.go](./router.go) вы можете зарегистрировать urls по которым будут доступны интерфейсы

```bash
/api/v1/example
```

## Запуск проекта

### Локально

Для запуска проекта локально используйте следующую команду:

```bash
go build golang-template
```

### В контейнере

Для запуска проекта в контейнере выполните следующую команду:

```bash
make up
```

## Тестирование
Тесты размещаются в файлах с суффиксом `_test.go`, которые хранятся там же где и файл с кодом, а не в отдельном каталоге. 
Используют функции из пакета `testing` и `testify`.

Для запуска выполните команду `go test ./... -cover -v -tags=unit`. Флаг -cover покажет процент покрытия тестами.

Для более детального отчета о покрытии используйте
`go test -coverpkg=$(go list ./... | grep -E "domain|repository" | tr '\n' ',' | sed 's/,$//') -v -tags=unit -coverprofile=coverage.out ./...`. 
Затем визуализируйте информацию о покрытие в браузере `go tool cover -html=coverage.out`
Обратите внимание, что покрытие мы считаем только по коду располагаемому в основном (доменном) слое и слое инраструктурном. 

## Анализ качества кода
Для использования `golangci-lint` линтера необходимо установить его локально
`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`. Запустить можно с помощью команды `golangci-lint run`,
команда проверяет не только форматирование кода, но и запускает статический анализ кода, а так же многое другое. 
Полный список проверок можно посмотреть в [.golangci.yml](.golangci.yml)

## Примеры реализации

### UI Приложение
В примере [examples/web_app](examples/web_app) можно увидеть где располагается статика (css, js) и 
каким образом организован основной код HTML

### API Приложение
В примере [examples/api](examples/web_app) можно увидеть реализацию кодовой базы, 
разделенной по слоям в парадигме DDD. Более подробно про DDD можно прочитать на [habr.com](https://habr.com/ru/companies/dododev/articles/489352/)

## Миграции
Файлы с миграциями распологаются в [migrations](migrations). Название файла стандартизированы и должны соответствовать 
маске `000001_migration_name.(up|down).sql`

По умолчанию они выключены и если необходимо включить нужно в пакете [db](db/db.go) раскомментировать вызов метода `db.MigrationsUp()`

Для более подробного изучения можно почитать [MIGRATIONS.md](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md)
## Swagger
Поддержка Swagger реализована с использованием пакета [go-swagger](https://github.com/go-swagger/go-swagger).

Генерация `swagger.json` выполняется в процессе сборки в [Dockerfile](./Dockerfile).
Добавьте аннотации в комментариях к методам, чтобы описать Swagger-спецификации. [Документация](https://goswagger.io/use/spec.html#annotation-syntax).

Для локальной генерации вы можете использовать Docker Compose или консольную утилиту `go-swagger`. Для macOS, вы можете установить с помощью brew:

```bash
brew tap go-swagger/go-swagger
brew install go-swagger
```

После запуска вашего приложения, документация Swagger будет доступна по следующим URL:

- `/_/swagger` - JSON-файл
- `/_/swaggerui` - интерфейс Swagger

## Healthcheck (Probes)

Шаблон поддерживает 4 типа проб:

- Readiness - `/_/readiness`
- Liveness - `/_/liveness`
- Startup - `/_/startup`
- Status - `/_/status` (обрабатывается как Liveness и сохранена для обратной совместимости с текущими настройками путей до проб на уровне инфраструктуры)

По умолчанию, все типы проб в Kubernetes обращаются к `/_/status`.

Чтобы добавить свою собственную проверку, опишите структуру, которая реализует интерфейс `probe.CheckInterface`. Затем зарегистрируйте проверку в соответствующей пробе с использованием внедрения зависимостей.

[app.go](app.go)

```code
app.container.Provide(func() *probe.Controller {
    return probe.NewProbeController(
        // startup
        *probe.NewCompositeCheckService(),
        // liveness
        *probe.NewCompositeCheckService(),
        // readiness
        *probe.NewCompositeCheckService(),
    )
})
```

## OpenMetrics (для Prometheus)

Поддерживается отдача метрик в формате OpenMetrics по URL `/_/metrics`.

Чтобы добавить свои метрики в приложение, зарегистрируйте их в файле [metrics/prometheus.go](metrics/prometheus.go).

## CORS (Cross-Origin Resource Sharing)

Поддерживается работа с CORS. По умолчанию включена.

## Access

Посредник, который предоставляет централизованную аутентификацию. пользовательских запросов.
Код располагается в [middleware/access.go](middleware/access.go). Для настройки правил доступа к страницам которые не требуют
авторизации, можно воспользоваться регулярными выражениями или просто перечислить их в [config/config.go](config/config.go)