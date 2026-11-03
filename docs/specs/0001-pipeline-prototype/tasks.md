# 0001 — задачи

`[P]` — можно делать параллельно с соседними.

## Где остановились

Всё сделано, включая прогон на реальных моделях: результат в спеке. Коммитов нет.

## Задачи

- [x] T-01 Скелет core: config, platform (logger, http, sqlite, validation), bootstrap, health
- [x] T-02 Домен: run, step, vault, фикстура; порты и моки
- [x] T-03 [P] Гейт 1: детекторы и тесты (FR-6)
- [x] T-04 [P] LLM-клиент на openai-go, три модели, тест на httptest (FR-5, FR-15)
- [x] T-05 Пайплайн: gate2, draft, egress, advise, finalize, judge, reidentify, канарейки; тесты на моках (FR-7…FR-16)
- [x] T-06 Хранилище sqlite: миграции, sqlc, репозитории (FR-4, FR-18)
- [x] T-07 HTTP: auth, runs, review, events SSE, cases, benchmarks; openapi (FR-1…FR-3, FR-17…FR-22)
- [x] T-08 Бенчмарк: прогон, сводка (FR-20)
- [x] T-09 [P] Фикстура: 20 писем, генератор ПДн, 5 PNG (FR-3, FR-21)
- [x] T-10 Web: каркас, вход, новый запуск, запуск с трассой, запуски, бенчмарк; тесты (FR-23…FR-28)
- [x] T-11 Makefile, Dockerfile, GitHub Actions, AGENTS.md, карта системы, ADR
- [x] T-12 Сквозной прогон с ключами оператора
