# Go Mentorship Platform

Платформа менторства для Go-разработчиков: дорожная карта обучения, прогресс по блокам, 1:1 с бадди, собеседования, геймификация и администрирование программы.

Монолитный backend (REST API), два SPA (студент/бадди и админка), PostgreSQL и Redis. Подходит для локальной разработки и деплоя через Docker Compose.

---

## Содержание

- [Возможности](#возможности)
- [Архитектура](#архитектура)
- [Стек технологий](#стек-технологий)
- [Структура репозитория](#структура-репозитория)
- [Требования](#требования)
- [Быстрый старт (Docker)](#быстрый-старт-docker)
- [Локальная разработка](#локальная-разработка)
- [Роли пользователей](#роли-пользователей)
- [Тестовые аккаунты](#тестовые-аккаунты)
- [API и health](#api-и-health)
- [Переменные окружения](#переменные-окружения)
- [Миграции и сиды](#миграции-и-сиды)
- [Production](#production)
- [Разработка и тесты](#разработка-и-тесты)

---

## Возможности

| Область | Описание |
|--------|----------|
| **Identity** | JWT (access + refresh), вход, выбор активной роли |
| **Профиль** | Профиль пользователя, видимость, Telegram |
| **Roadmap & Progress** | Блоки дорожной карты, материалы, статусы прогресса, согласование бадди |
| **1:1** | Заявки и встречи студент ↔ бадди |
| **Calendar** | События календаря |
| **Interviews** | Пробные и реальные собеседования, feedback |
| **Achievements & Bonus** | Достижения, бонусные баллы |
| **Final check** | Финальная проверка (tech / roast) |
| **Admin** | Управление пользователями, ролями, назначениями бадди, справочниками |
| **Notifications & Activity** | In-app уведомления, журнал активности |

---

## Архитектура

**Стиль:** модульный монолит — один deployable API и отдельные frontend-приложения.

**Backend:** Clean Architecture внутри bounded contexts (`domain` → `application` → `adapter`).

**Frontends:** Feature-Sliced Design (слои `app`, `pages`, `widgets`, `features`, `entities`, `shared`).

```mermaid
flowchart TB
  subgraph clients [Clients]
    SPA[frontend Student/Buddy]
    ADM[frontend-admin]
  end

  subgraph runtime [Runtime]
    API[Go API cmd/api]
    PG[(PostgreSQL)]
    RD[(Redis)]
  end

  SPA -->|"/api/v1 REST"| API
  ADM -->|"/api/v1 REST"| API
  API --> PG
  API --> RD
```

**Docker (production-like):** nginx в контейнерах frontend проксирует `/api/` на сервис `backend:8080`. Миграции выполняются one-shot контейнером `migrate` перед стартом API.

**Модули backend (`internal/`):**

| Модуль | Назначение |
|--------|------------|
| `identity` | Аутентификация, JWT, активная роль |
| `user` | Учётные записи, роли, назначения buddy |
| `profile` | Профили |
| `roadmap` | Дорожная карта и материалы |
| `progress` | Прогресс студента |
| `oneonone` | 1:1 встречи |
| `calendar` | Календарь |
| `interview` | Собеседования |
| `achievement`, `bonus` | Геймификация |
| `finalcheck` | Финальная проверка |
| `admin` | Admin API |
| `notification`, `activity` | Уведомления и активность |
| `platform` | Config, HTTP server, Postgres, Redis, wiring |

---

## Стек технологий

| Слой | Технологии |
|------|------------|
| **Backend** | Go 1.25, [chi](https://github.com/go-chi/chi), pgx, golang-migrate, Redis, JWT, bcrypt |
| **Student / Buddy UI** | React 19, TypeScript, Vite, MUI, TanStack Query, Zustand, React Router |
| **Admin UI** | React 19, TypeScript, Vite, MUI, TanStack Query, Zustand |
| **Data** | PostgreSQL 16, Redis 7 |
| **Deploy** | Docker, Docker Compose, nginx (static + reverse proxy) |

---

## Структура репозитория

```text
go-mentorship-platform/
├── cmd/
│   ├── api/                 # Точка входа HTTP API
│   └── migrate/             # CLI миграций
├── internal/                # Bounded contexts (Go)
├── migrations/              # SQL-миграции (версионированные)
├── seeds/                   # Dev-данные (SQL, вручную)
├── backend/
│   └── Dockerfile           # Образ API + migrate (context: корень репо)
├── frontend/                # SPA студент / бадди
│   ├── Dockerfile
│   ├── nginx.conf
│   └── src/                 # FSD: app, pages, widgets, features, entities, shared
├── frontend-admin/          # SPA администратора
│   ├── Dockerfile
│   ├── nginx.conf
│   └── src/
├── docker-compose.yml       # Полный стек
├── .env.example             # Шаблон для Compose и документация env
├── Makefile                 # build, test, migrate-up, compose-up
├── go.mod
└── README.md
```

Legacy: `deployments/docker-compose.yml` — устаревший вариант; используйте корневой `docker-compose.yml`.

---

## Требования

**Локальная разработка:**

- Go 1.25+
- Node.js 22+ и npm
- PostgreSQL 16+
- Redis 7+

**Docker:**

- Docker Engine
- Docker Compose (`docker-compose` или `docker compose`)

---

## Быстрый старт (Docker)

Поднимает PostgreSQL, Redis, миграции, backend, frontend и frontend-admin.

```bash
git clone <repository-url>
cd go-mentorship-platform

# опционально: переопределить порты и секреты
cp .env.example .env

docker-compose up -d --build
# или: make compose-up
```

**URL после старта:**

| Сервис | URL |
|--------|-----|
| Student / Buddy UI | http://localhost:5173 |
| Admin UI | http://localhost:5174 |
| API (напрямую) | http://localhost:8080 |
| Health | http://localhost:8080/health/live |
| Readiness | http://localhost:8080/health/ready |

API из браузера также доступен через nginx: `http://localhost:5173/api/v1/...`.

**Проверка:**

```bash
docker-compose ps
curl -s http://localhost:8080/health/ready
curl -s http://localhost:8080/api/v1/ping
```

**Сиды (пользователи для входа):** см. [Миграции и сиды](#миграции-и-сиды).

**Заметки:**

- Если порты `8080` / `5173` заняты локальным API или Vite, задайте в `.env` другие `HTTP_PORT` / `FRONTEND_PORT` или остановите dev-процессы.
- На `docker-compose` 1.x при ошибке `'ContainerConfig'` при пересоздании контейнера: `docker-compose rm -sf migrate && docker-compose up -d`.

---

## Локальная разработка

Корневой `.env` **не подхватывается** `go run` и Vite автоматически — используйте его только для Docker Compose.

### 1. Инфраструктура

Вариант A — только БД и Redis в Docker:

```bash
docker-compose up -d postgres redis
```

Вариант B — локальные PostgreSQL и Redis на `127.0.0.1`.

### 2. Миграции

```bash
export DATABASE_URL='postgres://mentorship:mentorship@127.0.0.1:5432/mentorship?sslmode=disable'
go run ./cmd/migrate -direction up
# или: make migrate-up
```

### 3. Сиды (dev-пользователи)

```bash
PGPASSWORD=mentorship psql -h 127.0.0.1 -U mentorship -d mentorship \
  -f seeds/001_dev_user.sql \
  -f seeds/002_student_buddy.sql
```

### 4. Backend

```bash
export DATABASE_URL='postgres://mentorship:mentorship@127.0.0.1:5432/mentorship?sslmode=disable'
export REDIS_ADDR='127.0.0.1:6379'
export HTTP_PORT=8081
export JWT_SECRET='dev-local-secret'
go run ./cmd/api
```

Порт **8081** согласован с proxy в `frontend/vite.config.ts`.

### 5. Frontends

```bash
# frontend/.env — VITE_API_BASE_URL=/api/v1
cd frontend && npm install && npm run dev      # http://localhost:5173

cd frontend-admin && npm install && npm run dev  # http://localhost:5174
```

---

## Роли пользователей

| Роль | Код | Клиент | Возможности (кратко) |
|------|-----|--------|----------------------|
| **Студент** | `student` | `frontend` | Дорожная карта, прогресс, 1:1, календарь, собеседования, профиль, бонусы |
| **Бадди** | `buddy` | `frontend` | Студенты на сопровождении, согласование прогресса, 1:1, mock-интервью, final check |
| **Администратор** | `admin` | `frontend-admin` (и admin API) | Пользователи, роли, назначения buddy, справочники, модерация |

**Активная роль:** пользователь с несколькими ролями выбирает контекст после входа (`/select-role`). JWT содержит `active_role`.

**Портал `frontend`:** только роли `student` и `buddy`. Учётка с одной ролью `admin` получит сообщение, что нужна отдельная админка.

**Назначение buddy:** связь студент ↔ бадди хранится в `buddy_assignments` (назначает администратор).

---

## Тестовые аккаунты

Применяются файлами из `seeds/` (не часть автоматического `docker-compose up`).

**Пароль для всех dev-аккаунтов:** `changeme`

| Email | Роли | Профиль | Примечание |
|-------|------|---------|------------|
| `student@example.com` | student | Иван Студентов | Вход в `frontend`, сразу роль студента |
| `buddy@example.com` | buddy | Мария Бадди | Вход в `frontend`, раздел бадди |
| `admin@example.com` | admin, student | Admin | В `frontend` выбрать **Студент**; админ-функции — в `frontend-admin` |

Студент `student@example.com` привязан к бадди `buddy@example.com` (сид `002_student_buddy.sql`).

---

## API и health

- **Base path:** `/api/v1`
- **Пример:** `GET /api/v1/ping` → `{"data":{"message":"pong"}}`
- **Auth:** `POST /api/v1/auth/login`, `GET /api/v1/auth/me`, `PUT /api/v1/auth/active-role`
- **Live:** `GET /health/live`
- **Ready:** `GET /health/ready` (PostgreSQL + Redis)

---

## Переменные окружения

Полный список и комментарии — в [`.env.example`](.env.example):

- Docker: хосты `postgres`, `redis` внутри compose (URL для backend собирается из `POSTGRES_*` в `docker-compose.yml`)
- Local: `127.0.0.1`, `HTTP_PORT=8081` для API
- Production: `APP_ENV=production`, сильный `JWT_SECRET`, `CORS_ALLOWED_ORIGINS`

---

## Миграции и сиды

| Команда | Назначение |
|---------|------------|
| `go run ./cmd/migrate -direction up` | Применить миграции |
| `go run ./cmd/migrate -direction down` | Откат (осторожно) |
| `psql ... -f seeds/*.sql` | Dev-пользователи и demo-профили |

В Docker миграции запускает сервис `migrate` один раз перед `backend`.

---

## Production

Перед выкладкой:

1. Задайте `APP_ENV=production`, уникальный `JWT_SECRET`, надёжный пароль PostgreSQL.
2. Ограничьте `CORS_ALLOWED_ORIGINS` реальными origin фронтендов.
3. Не публикуйте порты PostgreSQL/Redis в интернет (уберите или ограничьте `ports` в compose).
4. Соберите образы с `VITE_USE_MOCK=false` для admin (дефолт в compose).
5. Добавьте TLS-терминацию (reverse proxy) перед nginx-контейнерами.

Подробнее — блок **Production deploy** в `.env.example`.

---

## Разработка и тесты

```bash
make build          # bin/api, bin/migrate
make test           # go test ./...
make migrate-up
make compose-up     # docker-compose up -d --build
make compose-down
```

---

## Лицензия

Уточните лицензию у maintainers репозитория при публикации open-source.
