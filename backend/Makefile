.PHONY: run-api run-worker test lint sqlc-generate migrate-up migrate-down test-integration

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

test:
	go test ./...

lint:
	golangci-lint run

sqlc-generate:
	sqlc generate -f sqlc.yaml

migrate-up:
	: "$${DB_DSN:?未设置 DB_DSN，无法执行 migrate-up}"
	goose -dir migrations postgres "$${DB_DSN}" up

migrate-down:
	: "$${DB_DSN:?未设置 DB_DSN，无法执行 migrate-down}"
	goose -dir migrations postgres "$${DB_DSN}" down

test-integration:
	go test ./test/integration
