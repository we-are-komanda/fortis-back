up:
	docker-compose up -d --build

start:
	docker-compose up -d

stop:
	docker-compose stop

down:
	docker-compose down --remove-orphans

run:
	go run ./cmd/app/

# Локальный запуск с подгрузкой .env в окружение (приложение само .env не читает).
run-dev:
	set -a && . ./.env && set +a && go run ./cmd/app/

# Live-reload через Air: ставится `make air-install`, .env подхватывается автоматически.
# Конфиг — .air.toml. Air вызывается из GOPATH/bin, чтобы не зависеть от PATH.
dev:
	set -a && . ./.env && set +a && $(shell go env GOPATH)/bin/air

air-install:
	go install github.com/air-verse/air@latest

down-force:
	docker-compose down --remove-orphans -v

test-unit:
	go test -v -cover -tags=unit ./...

install-hooks:
	git config core.hooksPath .githooks/

lint-install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

swagger:
	swagger generate spec -o swagger.json --scan-models

diff-test-check:
	mkdir -p report
	pip install 'diff_cover==3.0.1'
	go test -coverpkg=$(go list ./... | grep -E "domain|repository" | tr '\n' ',' | sed 's/,$//') -v -tags=unit -coverprofile=coverage.out ./...
	go tool cover -func coverage.out
	go tool cover -html=coverage.out -o report/coverage.html
	gocover-cobertura < coverage.out > report/Cobertura.xml
	sed -i -- "s#filename=\"$(notdir $(shell pwd))/#filename=\"#g" report/Cobertura.xml
	diff-cover report/Cobertura.xml
