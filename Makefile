include .envrc
MIGRATIONS_PATH=./cmd/migrate/migrations

.PHONY: migrate-create
migrate-create:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path $(MIGRATIONS_PATH) -database $(DB_ADDR) up

.PHONY: migrate-down
migrate-down:
	@migrate -path $(MIGRATIONS_PATH) -database $(DB_ADDR) down

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: swagger
swagger:
	@swag init -g ./main.go -d ./cmd/api,./internal/store,./internal/utils -o ./docs && swag fmt

.PHONY: test
test:
	@go test -v ./...
