# Карта системы

Что с чем связано. Правила — в `AGENTS.md`, решения — в [adr/](../adr/), требования — в [specs/0001](../specs/0001-pipeline-prototype/spec.md).

## Развёртывание

Один бинарь: Go-сервис со вшитым `web/` и SQLite в одном файле. Образ собирает GitHub Actions и пушит в Yandex Container Registry; деплой вне этого репозитория.

```mermaid
flowchart LR
  browser["Браузер"] --> core["fluffy-giggle · :8000"]
  core --> db[("SQLite · /data/app.db")]
  core -->|"vision · gate 2 · draft · finalize"| local["«локальные» модели<br/>OpenAI-совместимый API"]
  core -->|"advise · judge<br/>только после egress"| cloud["«облачная» модель<br/>OpenAI-совместимый API"]
```

Все три модели — внешние endpoint'ы ([ADR 0002](../adr/0002-logical-boundary.md)). «Локальная» сторона — те, кому можно видеть сырые данные; «облачная» получает только обезличенную карточку и черновик.

## Пайплайн

```mermaid
flowchart TD
  in["вход: текст или скан"] --> vision["vision · local<br/>транскрипция"]
  in --> g1
  vision --> g1["gate 1 · system<br/>регулярки, контрольные суммы, ФИО"]
  g1 --> g2["gate 2 · local, до 3 проходов<br/>модель отдаёт спаны, код заменяет"]
  g2 -->|"пусто"| draft["draft · local<br/>карточка + черновик"]
  g2 -->|"всё ещё находит"| review["needs_review<br/>оператор решает"]
  review -->|"continue"| draft
  review -->|"finish_without_advisor"| fin
  draft --> egress["egress · system<br/>vault, цифры, детекторы"]
  egress -->|"чисто"| advise["advise · cloud<br/>до 5 правок"]
  egress -->|"находка"| blocked["blocked · совета нет"]
  advise --> fin["finalize · local"]
  blocked --> fin
  fin --> judge["egress → judge · cloud<br/>оценка 1–5"]
  judge --> re["reidentify · system<br/>плейсхолдеры → значения"]
```

Режимы adviser: `on`, `off` (совета нет), `placebo` (три общих замечания без обращения к облаку). Судья работает во всех режимах, поэтому бенчмарк сравнивает их на одной шкале. Канарейки фикстуры считаются по тому, что реально ушло в cloud-шаги.

## Путь запроса

```mermaid
flowchart TD
  req["HTTP"] --> mw["Recoverer · Logging"]
  mw --> health["GET /health"]
  mw --> v1["/v1 · Basic auth"]
  mw --> ui["/* → index.html"]
  v1 --> handler["handler · DTO"]
  handler --> usecase["usecase"]
  usecase --> port["port"]
  port --> storage["storage · sqlc · SQLite"]
  usecase --> queue["worker · очередь запусков"]
  queue --> pipeline["runpipeline"]
  pipeline --> bus["eventbus"] --> sse["GET /v1/events · SSE"]
```

`POST /v1/runs` только создаёт запуск; воркеры (`PIPELINE_CONCURRENCY`) исполняют его в фоне и публикуют каждый шаг в шину, которую читает SSE. Страница запуска живёт на этих событиях.

## Границы

| Кто | Что видит |
|---|---|
| `platform` | ничего о прикладном коде |
| `domain` | стандартную библиотеку и себя |
| `usecase` | `domain` |
| `presentation`, `storage`, `llm`, `anonymizer`, `fixture` | `usecase`, `domain`, `platform` |
