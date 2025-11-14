.PHONY: build run dev clean

build:
	go build -o bin/workflow-server main.go

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
