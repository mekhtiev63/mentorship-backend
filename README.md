# Mentorship Backend (Go API)

REST `/api/v1`, PostgreSQL, Redis, JWT. Исходники **в корне этого репозитория** (`cmd/`, `internal/`, …).

**Соседние репо:** [раскладка](../MENTORSHIP.md) — `mentorship-frontend-master`, `mentorship-frontend-admin`.

---

## Структура

```text
mentorship-backend/
├── cmd/api, cmd/migrate
├── internal/, migrations/, seeds/, pkg/
├── Dockerfile, docker-compose.yml
├── Makefile, docs/, .env.example
└── README.md
```

---

## Локально

```bash
export DATABASE_URL='postgres://mentorship:mentorship@127.0.0.1:5432/mentorship?sslmode=disable'
export REDIS_ADDR='127.0.0.1:6379'
export HTTP_PORT=8081
export JWT_SECRET='dev-local-secret'

make migrate-up
make run
```

Сиды: `psql ... -f seeds/001_dev_user.sql -f seeds/002_student_buddy.sql`

---

## Docker

| Команда | Что поднимает |
|---------|----------------|
| `make compose-api-up` | postgres, redis, migrate, backend |
| `docker-compose up -d --build` | + frontends из `../mentorship-frontend-master` и `../mentorship-frontend-admin` |

---

## Тестовые аккаунты

Пароль **`changeme`**: `student@example.com`, `buddy@example.com`, `admin@example.com`.

---

## Документация

- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- [.env.example](.env.example)
