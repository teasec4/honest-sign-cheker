# onestsignt-node-api

TypeScript/Node.js версия текущего Go API.

```bash
bun install
bun run dev
```

По умолчанию сервер слушает `127.0.0.1:8080`, поэтому существующий SvelteKit proxy может ходить в него без изменений.

Эндпоинты:

- `GET /api/health`
- `POST /api/primary-check` multipart: `issued`, `returned`, `minPercent`
- `POST /api/duplicate-check` multipart: `restored`

Поддерживаемые форматы такие же, как в Go: `.csv`, `.txt`, `.xlsx`, `.xlsm`.
