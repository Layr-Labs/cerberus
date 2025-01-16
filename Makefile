APP_NAME=cerberus
GO_LINES_IGNORED_DIRS=
GO_PACKAGES=./internal/... ./cmd/...
GO_FOLDERS=$(shell echo ${GO_PACKAGES} | sed -e "s/\.\///g" | sed -e "s/\/\.\.\.//g")

.PHONY: build
build:
	@echo "Building..."
	go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go
	@echo "Done"

.PHONY: start
start:
	make build
	./bin/$(APP_NAME) --log-level=debug

.PHONY: fmt
fmt: ## formats all go files
	go fmt ./...
	make format-lines

.PHONY: format-lines
format-lines: ## formats all go files with golines
	go install github.com/segmentio/golines@latest
	golines -w -m 100 --ignore-generated --shorten-comments --ignored-dirs=${GO_LINES_IGNORED_DIRS} ${GO_FOLDERS}

.PHONY: lint
lint: ## runs all linters
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

.PHONY: tests
tests: ## runs all tests
	go test ./... -covermode=atomic

.PHONY: docker
docker: ## runs docker build
	docker build -t $(APP_NAME):latest .

.PHONY: migrate
migrate: ## runs database migrations
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	migrate -path internal/database/migrations/ -database "postgres://user:password@localhost:5432/cerberus?sslmode=disable" --verbose up

.PHONY: create-migration
create-migration: ## creates a new database migration
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	migrate create -dir internal/database/migrations/ -ext sql $(name)
