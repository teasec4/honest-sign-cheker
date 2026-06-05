# onestsignt

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Node.js](https://img.shields.io/badge/Node.js-20+-339933?logo=node.js&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript&logoColor=white)
![Svelte](https://img.shields.io/badge/SvelteKit-2-FF3E00?logo=svelte&logoColor=white)
![Bun](https://img.shields.io/badge/Bun-1.x-000000?logo=bun&logoColor=white)

Локальный инструмент для сверки кодов маркировки.

Сценарий работы простой: мы выдаем фабрике коды в CSV или Excel, фабрика возвращает размеченный файл после нанесения, а приложение проверяет возврат:

- считает количество кодов;
- показывает точные совпадения с выданными кодами;
- находит похожие поврежденные коды для ручного восстановления;
- показывает коды, которые не удалось сопоставить;
- ловит дубликаты в возврате или в восстановленном файле.

Коды, которые были выданы, но не вернулись от фабрики, не считаются ошибкой: коды выдаются с запасом.

## Стек

- `web/` - SvelteKit интерфейс.
- `node-api/` - TypeScript/Node.js API.
- `cmd/`, `internal/` - Go-версия логики и CLI-команды.

Основной рабочий вариант сейчас: `SvelteKit + Node API`.

## Форматы файлов

Поддерживаются:

- `.csv`
- `.txt`
- `.xlsx`
- `.xlsm`

Старый Excel-формат `.xls` не поддерживается.

Для Excel, если в первой строке есть колонка `DM CODE`, приложение читает именно эту колонку и пропускает строку заголовка.

## Быстрый запуск

Из корня проекта:

```bash
bun install --cwd web
bun install --cwd node-api
bun run dev
```

После запуска:

- API: `http://127.0.0.1:8080`
- Web: адрес покажет Vite в терминале, обычно `http://127.0.0.1:5173`

## Варианты запуска

Запустить все вместе:

```bash
bun run dev
```

Запустить только Node API:

```bash
bun run dev:api
```

Запустить только web:

```bash
bun run dev:web
```

Запустить собранный Node API:

```bash
bun run build:api
bun run --cwd node-api start
```

Запустить Go API вместо Node API:

```bash
go run ./cmd/api
```

Важно: Node API и Go API по умолчанию используют один порт `127.0.0.1:8080`, поэтому одновременно их запускать не нужно.

## Проверки

Проверить TypeScript API:

```bash
bun run check:api
bun run build:api
```

Проверить SvelteKit:

```bash
bun run check:web
bun run build:web
```

Проверить Go-код:

```bash
go test ./...
go vet ./...
```

## CLI на Go

Первичная сверка:

```bash
go run ./cmd/primary-check -issued path/to/issued.csv -returned path/to/returned.xlsx -out reports/manual
```

Проверка дубликатов в восстановленном файле:

```bash
go run ./cmd/duplicate-check -input path/to/restored.csv -out reports/manual
```

## API

Основные эндпоинты:

- `GET /api/health`
- `POST /api/primary-check`
- `POST /api/duplicate-check`

Фронт отправляет файлы как `FormData`, пути к файлам вручную указывать не нужно.
