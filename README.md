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
