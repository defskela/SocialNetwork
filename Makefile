include .env
export

.PHONY: build run test lint switch-branch migrate-up migrate-down docker-up docker-down

APP_NAME=social-network
CMD_PATH=cmd/app/main.go

build:
	go build -o bin/$(APP_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH)

test:
	go test -v -race -cover ./...

lint:
	$(shell go env GOPATH)/bin/golangci-lint run

fmt:
	go fmt ./...
	$(shell go env GOPATH)/bin/goimports -w -local github.com/defskela/SocialNetwork .

swagger:
	$(shell go env GOPATH)/bin/swag init -g $(CMD_PATH)

docker-up:
	docker compose up -d

restart:
	docker compose up -d --build

docker-down:
	docker compose down

switch-branch:
	@if [ -z "$(NAME)" ]; then echo "Usage: make switch-branch NAME=<new_branch_name>"; exit 1; fi
	git checkout main
	git pull origin main
	git checkout -b $(NAME)

MIGRATE_CMD=migrate -path migrations -database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSL_MODE)"

migrate-up:
	$(MIGRATE_CMD) up

migrate-down:
	$(MIGRATE_CMD) down

migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=<migration_name>"; exit 1; fi
	migrate create -ext sql -dir migrations -seq $(NAME)
