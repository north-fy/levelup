<p align="center">
  <img src="docs/previews/roadmap-graph.svg" alt="Граф роадмапы" width="100%" />
</p>

<p align="center">
  <a href="README.md">English</a> · <b>Русский</b>
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

## ✨ Что это

LevelUp — это **RPG-трекер задач**:

- 🌿 **Ветки и квесты** — задачи по направлениям, каждый квест даёт XP и золото.
- ⏱️ **Таймеры** — timed-квесты засекают часы и платят пропорционально времени.
- 🗺️ **Роадмапы-графы** — узлы с пререквизитами образуют настоящий направленный ациклический граф.
- 🛒 **Магазин** — тратьте золото на товары других игроков или продавайте свои.
- 🧰 **Воркшоп** — публикуйте роадмапы и устанавливайте чужие.
- 📊 **Статистика** — XP, золото, выполненные квесты и часы в ClickHouse.
- 🎨 **Лендинг + личный кабинет** — светлая/тёмная тема, доступный интерфейс (WAI-ARIA).

## 🛠️ Технологии

**Бэкенд** — Go 1.26, Gin, GORM, PostgreSQL, Redis (кэш + rate limit), ClickHouse (статистика через outbox), JWT-авторизация + GitHub OAuth, метрики Prometheus, Swagger.

**Фронтенд** — React 18, TypeScript, Tailwind CSS, Vite.

**Инфраструктура** — Docker, Prometheus, Grafana, Github Actions, Nginx.

## Как использовать

### Docker

```bash
# из корня репозитория
$ docker compose -f deploy/docker-compose.dev.yml up -d --build
```
### Нативная разработка

```bash
$ docker compose -f deploy/docker-compose.dev.yml up -d postgres redis clickhouse

$ cp .env.example .env  

$ make migrate-up
$ make migrate-up-clickhouse
$ make run

$ cd web && npm ci && npm run dev
```

## 📚 Документация API

Интерактивный Swagger UI: **http://localhost:8080/swagger/index.html** (также доступен на http://localhost/swagger).

Бэкенд отдаёт `GET /healthz`, `GET /readyz`, `GET /metrics` и REST API `/api/v1/*` с JWT-авторизацией.

## 📈 Метрики

- **Prometheus** — http://localhost:9090 (метрики сервиса, задержки запросов, HTTP-коды).
- **Grafana** — http://localhost:3000 (дашборд `levelup.json`, логин `admin`/`admin`).

## 🧪 Тестирование

```bash
$ make test    
$ make e2e     
```

## 📄 Лицензия

Проект распространяется под [лицензией MIT](LICENSE). 