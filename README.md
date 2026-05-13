# hh-autoresponder

AI-powered job application automation system.

## Stack
- **Backend**: Go 1.22, chi router, WebSocket
- **Frontend**: React 18, TypeScript, Vite, TailwindCSS
- **Proxy**: Nginx
- **Container**: Docker + Docker Compose

## Quick Start

```bash
# 1. Clone
git clone ...
cd hh-autoresponder

# 2. Setup env
cp deploy/.env.example deploy/.env
# Edit deploy/.env — add HH_CLIENT_ID, HH_CLIENT_SECRET, OPENAI_API_KEY

# 3. Start (production)
cd deploy
make up

# 4. Open browser
open http://localhost
```

## Development

```bash
cd deploy
make dev        # start with hot reload
make dev-down   # stop
```

Frontend runs at: http://localhost:5173
Backend runs at: http://localhost:8080

## Make Commands

| Command | Description |
|---|---|
| `make dev` | Start dev environment |
| `make dev-build` | Rebuild dev images |
| `make dev-down` | Stop dev |
| `make up` | Start production |
| `make down` | Stop production |
| `make restart` | Restart production |
| `make rebuild` | Rebuild & restart |
| `make logs` | Tail all logs |
| `make backend-logs` | Tail backend |
| `make frontend-logs` | Tail frontend |
| `make ps` | Show containers |
| `make health` | Health check |
| `make clean` | Prune stopped containers |
| `make prune` | Full cleanup |
| `make backend-shell` | Shell into backend |
| `make fmt` | Format Go code |
| `make lint` | Lint Go code |

## Architecture

```text
Browser → Nginx (:80)
              ├── /api/*  → Backend (:8080)
              ├── /ws     → Backend (:8080) [WebSocket]
              └── /*      → Frontend (:80)  [SPA]
```

## Environment Variables

See `deploy/.env.example` for all variables.

## VPS Deployment

```bash
# On VPS
git clone ...
cd hh-autoresponder/deploy
cp .env.example .env
nano .env   # fill secrets
make up
```

## Troubleshooting

**Backend не стартует**: проверь `make backend-logs`, убедись что `ACCOUNTS_FILE` доступен.
**Frontend не билдится**: проверь `make frontend-logs`, убедись что `npm ci` отработал.
**WebSocket не работает**: проверь nginx конфиг, должен быть `/ws` location.
**Rebuild**: `make rebuild` пересобирает все контейнеры с нуля.
