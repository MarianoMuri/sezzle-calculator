.PHONY: install test coverage backend frontend docker

install:            ## Install frontend dependencies
	cd frontend && npm ci

test:               ## Run all unit tests
	cd backend && go test -race ./...
	cd frontend && npm test

coverage:           ## Run tests with coverage for both layers
	cd backend && go test -race -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out && go tool cover -html=coverage.out -o coverage.html
	cd frontend && npm run coverage

backend:            ## Run the API on :8080
	cd backend && go run ./cmd/server

frontend:           ## Run the React dev server on :5173 (proxies /api to :8080)
	cd frontend && npm run dev

docker:             ## Build and run the full stack on :3000
	docker compose up --build
