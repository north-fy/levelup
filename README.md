<p align="center">
  <img src="docs/previews/roadmap-graph.svg" alt="Roadmap graph" width="100%" />
</p>

<p align="center">
  <b>English</b> · <a href="README.ru.md">Русский</a>
</p>

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" />
  <img alt="React" src="https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=white" />
  <img alt="TypeScript" src="https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript&logoColor=white" />
  <img alt="Tailwind" src="https://img.shields.io/badge/Tailwind-3-38BDF8?logo=tailwindcss&logoColor=white" />
  <img alt="Docker" src="https://img.shields.io/badge/Docker-compose-2496ED?logo=docker&logoColor=white" />
  <img alt="E2E" src="https://img.shields.io/badge/e2e-green" />
  <img alt="License: MIT" src="https://img.shields.io/badge/license-MIT-green" />
</p>

---

## What is this

LevelUp is an **RPG task tracker**:

-  **Branches & quests** — organize tasks into directions, each quest rewards XP and gold.
-  **Timers** — timed quests track hours and pay proportional rewards.
-  **Roadmap graphs** — nodes with prerequisites form a real directed acyclic graph.
-  **Shop** — spend gold on other players' items or sell your own.
-  **Workshop** — publish roadmaps and install other people's ones.
-  **Statistics** — XP, gold, completed quests and hours in ClickHouse.
-  **Landing + personal account** — light/dark theme, accessible interface (WAI-ARIA).

## Technologies

**Backend** — Go 1.26, Gin, GORM, PostgreSQL, Redis (cache + rate limit), ClickHouse (statistics via outbox), JWT auth + GitHub OAuth, Prometheus metrics, Swagger.

**Frontend** — React 18, TypeScript, Tailwind CSS, Vite.

**Infrastructure** — Docker, Prometheus, Grafana, GitHub Actions, Nginx.

## How to use

### Docker

```bash
# from the repo root
docker compose -f deploy/docker-compose.dev.yml up -d --build
```
### Native development

```bash
$ docker compose -f deploy/docker-compose.dev.yml up -d postgres redis clickhouse

$ cp .env.example .env  

$ make migrate-up
$ make migrate-up-clickhouse
$ make run

$ cd web && npm ci && npm run dev
```

## API documentation

Interactive Swagger UI: **http://localhost:8080/swagger/index.html** (also available at http://localhost/swagger).

The backend exposes `GET /healthz`, `GET /readyz`, `GET /metrics` and the `/api/v1/*` REST API with JWT auth.

## Metrics

- **Prometheus** — http://localhost:9090 (service metrics, request latencies, HTTP codes).
- **Grafana** — http://localhost:3000 (dashboard `levelup.json`, credentials `admin`/`admin`).

## Testing

```bash
make test    
make e2e     
```

## 📄 License

Released under the [MIT License](LICENSE).
