export CONFIG_PATH=configs/local.yaml

.PHONY: build run test lint switch-branch fmt swagger restart check

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

restart: swagger
	docker compose up -d --build

check: fmt lint test


switch-branch:
	@if [ -z "$(NAME)" ]; then echo "Usage: make switch-branch NAME=<new_branch_name>"; exit 1; fi
	git checkout main
	git pull origin main
	git checkout -b $(NAME)

