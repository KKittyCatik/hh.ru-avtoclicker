.PHONY: dev dev-build dev-down up down restart rebuild ps logs backend-logs frontend-logs nginx-logs clean prune backend-shell frontend-shell health fmt lint

COMPOSE_PROD = docker compose -f docker-compose.yml
COMPOSE_DEV  = docker compose -f docker-compose.dev.yml

# ── DEVELOPMENT ──────────────────────────────────────────────

dev: ## Start dev environment with hot reload
$(COMPOSE_DEV) up -d

dev-build: ## Rebuild dev images
$(COMPOSE_DEV) build --no-cache

dev-down: ## Stop dev environment
$(COMPOSE_DEV) down

# ── PRODUCTION ───────────────────────────────────────────────

up: ## Start production environment
$(COMPOSE_PROD) up -d

down: ## Stop production environment
$(COMPOSE_PROD) down

restart: down up ## Restart production

rebuild: ## Rebuild and restart production
$(COMPOSE_PROD) build --no-cache
$(COMPOSE_PROD) up -d

# ── LOGS ─────────────────────────────────────────────────────

logs: ## Tail all logs
$(COMPOSE_PROD) logs -f

backend-logs: ## Tail backend logs
$(COMPOSE_PROD) logs -f backend

frontend-logs: ## Tail frontend logs
$(COMPOSE_PROD) logs -f frontend

nginx-logs: ## Tail nginx logs
$(COMPOSE_PROD) logs -f nginx

# ── UTILS ────────────────────────────────────────────────────

ps: ## Show running containers
$(COMPOSE_PROD) ps

backend-shell: ## Open shell in backend container
$(COMPOSE_PROD) exec backend sh

frontend-shell: ## Open shell in frontend container
$(COMPOSE_PROD) exec frontend sh

health: ## Check health of all services
@echo "=== Backend ===" && curl -sf http://localhost:80/api/health || echo "UNHEALTHY"
@echo "=== Nginx ===" && curl -sf http://localhost:80/nginx-health || echo "UNHEALTHY"

clean: ## Remove stopped containers and unused images
docker system prune -f

prune: ## Full cleanup including volumes
$(COMPOSE_PROD) down -v
docker system prune -af

fmt: ## Format Go code
cd .. && gofmt -w .

lint: ## Lint Go code
cd .. && go vet ./...

help: ## Show this help
@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
