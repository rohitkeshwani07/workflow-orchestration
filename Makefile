.PHONY: help install dev build clean docker-build docker-up docker-down docker-logs

help:
	@echo "Workflow Orchestration - Available Commands:"
	@echo ""
	@echo "  make install       - Install all dependencies"
	@echo "  make dev          - Start development servers"
	@echo "  make build        - Build for production"
	@echo "  make clean        - Clean build artifacts"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build - Build Docker images"
	@echo "  make docker-up    - Start containers"
	@echo "  make docker-down  - Stop containers"
	@echo "  make docker-logs  - View container logs"
	@echo ""

install:
	@echo "Installing Go dependencies..."
	cd backend && go mod download
	@echo "Installing frontend dependencies..."
	npm install
	@echo "Done! Run 'make dev' to start development servers."

dev:
	npm run dev

build:
	npm run build

clean:
	rm -rf backend/bin/
	rm -rf backend/data/
	rm -rf packages/frontend/dist/
	rm -rf packages/frontend/node_modules/
	rm -rf node_modules/

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d
	@echo ""
	@echo "✅ Containers started!"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend:  http://localhost:3001"

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-clean:
	docker-compose down -v
	docker system prune -f
