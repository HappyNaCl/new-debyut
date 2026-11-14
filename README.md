# new-debyut

- `frontend/` — Svelte 5 SPA (TypeScript, Vite, Bun)
- `backend/` — Go gRPC microservices behind a Gin REST gateway

## Run

```sh
docker compose up --build
cd backend && make migrate
cd frontend && bun install && bun run dev
```

API: http://localhost:8080/api · Swagger: http://localhost:8080/swagger/index.html · App: http://localhost:5173

## Sample data

`cd backend && make seed` migrates, then adds categories, an admin
(`admin@debyut.id`), 20 random users, and six organizations with sample events
(12 each by default: upcoming, past and draft, across cities and formats; pass
`-events N` to `run-migrations -seed` for more). Organizer logins are
`organizer@debyut.id` and `organizer.{nusantara,ruang,tawa,lari,tekno}@debyut.id`.
Seeding is safe to rerun: existing rows are kept, and organizations that already
have events get no more. Published sample events are queued for search indexing.

## Search

Event search runs on Meilisearch (set `MEILISEARCH_MASTER_KEY` in `.env`). The
event service keeps the index in step through a transactional outbox
(`search_outbox`) and fills an empty index on its own. To rebuild it anyway,
run `cd backend && make reindex-search`.
