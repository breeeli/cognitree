.PHONY: dev-db dev-db-down frontend-install backend frontend dev

dev-db:
	docker compose up -d

dev-db-down:
	docker compose down

frontend-install:
	cd frontend && npm install

backend:
	cd backend && go run cmd/server/main.go

frontend:
	cd frontend && npm run dev

dev: dev-db
	@echo "Database started in WSL. Run 'make backend' and 'make frontend' in separate WSL terminals."
