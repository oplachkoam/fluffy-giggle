# core

Go 1.27: `uuid` из стандартной библиотеки (`uuid.NewV7()`), `errors.AsType[T]`, `slog.DiscardHandler` в тестах.
DI — uber-fx, HTTP — chi, SQLite — `modernc.org/sqlite` без CGO, запросы — sqlc, моки — mockgen.

## Раскладка

```
cmd/app/main.go        только bootstrap.Run()
internal/bootstrap/    граф fx: config → platform → прикладной код → http.Server
internal/config/       единый Config, cleanenv, префиксы env по подсистемам
internal/platform/     database/sqlite, http (server, router, render, middleware), logger, validation
internal/domain/       entity/, vo/, port/ (+ port/mock), errs/
internal/usecase/<op>/ usecase.go (+ model.go); runpipeline/ — сам пайплайн и воркер
internal/anonymizer/   гейт 1: детерминированные детекторы
internal/llm/          клиент OpenAI-совместимого API, один на три роли
internal/fixture/      20 синтетических писем, генератор ПДн, 5 сканов
internal/storage/      queries/*.sql → sqlite/sqlc (генерируется), sqlite/<агрегат>/, memory/eventbus
internal/presentation/http/  auth/, handler/<op>/, routes/, dto/, ui/ (вшитый SPA)
api/openapi.yml        контракт
migrations/            NNN_name.sql, применяются на старте бинаря
```

Зависимости смотрят внутрь: `presentation → usecase → port ← storage`. Домен не импортирует
`platform`, `presentation`, `storage`. Usecase зависит только от `port.*`. `platform` не знает о прикладном коде.

## Правила

- **Сущности** — приватные поля, геттеры, `New*`/`Restore*` через `*State`, переходы состояния — методы с проверкой.
- **Usecase** — оркестрация: валидация → `errs.NewValidationError`; отказ по правилу — `errs.Err*`;
  инфраструктурная ошибка — `fmt.Errorf("%s: %w", msg, err)`; смена состояния — `slog.InfoContext` с идентификаторами.
- **Handler** — тонкий: DTO → usecase → `switch errors.Is` → `httprender.*`.
- **Storage** — реализует `port.*`; «не найдено» — `(nil, nil)`.
- **Сильная модель** вызывается только из `runpipeline.callCloud` после `Egress`. Другого пути нет и не будет.
- Ключи, пароль и тела запросов к моделям в логи не попадают; в трассу шагов попадают намеренно — это и есть наблюдаемость.

## Ошибки и HTTP

| Ошибка | Статус |
|---|---|
| `ErrValidationError` | 422, с именем поля |
| `ErrUnauthenticated` | 401, без `WWW-Authenticate` |
| `ErrRunNotFound`, `ErrRunHasNoImage`, `ErrCaseNotFound`, `ErrBenchmarkNotFound` | 404 |
| `ErrRunNotAwaitingReview`, `ErrBenchmarkRunning` | 409 |
| прочее | 500, без деталей |

## Тесты

Unit рядом с кодом, `go test -race ./...`. Пайплайн — на фейках моделей и хранилищ, граф — `fx.ValidateApp`.
Интеграционных тестов нет. Покрытие исключает `sqlc/` и `mock/`.

## Стиль

golangci-lint v2 (`.golangci.yml`), `golangci-lint fmt`, строка до 110, импорты тремя группами, local-prefix `fluffy-giggle/core`.
