.PHONY: build run dev clean migrate-up migrate-down migrate-create migrate-version

build:
	go build -o bin/workflow-server main.go
	go build -o bin/migrate cmd/migrate/main.go

run: build
	./bin/workflow-server

dev:
	go run main.go

clean:
	rm -rf bin/
	rm -rf data/

deps:
	go mod download
	go mod tidy

test:
	go test ./...

# Migration commands
migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

migrate-version:
	go run cmd/migrate/main.go version

migrate-create:
	@echo "Usage: make migrate-create NAME=migration_name"
	@echo "Then manually create the migration files in backend/migrations/"

migrate-force:
	@echo "Usage: go run cmd/migrate/main.go force [version]"
