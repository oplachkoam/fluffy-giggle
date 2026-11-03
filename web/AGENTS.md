# web

React 19 + TypeScript + Vite. Tailwind 4, shadcn (стиль `radix-nova`), иконки lucide, данные — TanStack Query,
роутер — react-router, тесты — vitest + testing-library.

## Раскладка

```
src/app/          App.tsx, router.tsx, query-client.ts, AppLayout.tsx
src/pages/<area>/ страницы + тесты рядом
src/components/   компоненты проекта; components/ui/ — shadcn, не правим
src/lib/          api.ts, auth.ts, theme.ts, events.ts (SSE), <домен>.ts (типы, маппинг, хуки)
src/test/         render.tsx (renderRoute, stubApi), setup.ts, fixtures
src/index.css     токены темы
```

## Правила

- Весь HTTP — через `api()`; `fetch` в компонентах нет. Бэкенд отдаёт `snake_case`, маппинг — в `lib/<домен>.ts`.
- Креды — `lib/auth.ts`, заголовок `Authorization: Basic`; `401` стирает их, роутер уводит на `/login`.
- Один SSE-поток на вкладку (`useEvents` в `AppLayout`) кладёт события в кэш запросов; страницы читают кэш.
- Только утилиты Tailwind и семантические токены; hex в компонентах нет. Светлая и тёмная темы.
- Интерфейс: минимализм, монохром, цвет только как сигнал. Главное на экране — что произошло с данными и что ушло в облако.
- Тексты на русском, на «вы», без восклицаний.
- Тесты проверяют то, что видит пользователь: текст, роли, поведение после действия.

## Команды

`npm run dev` · `build` · `lint` · `typecheck` · `test` · `format`. `build` пишет в `core/internal/presentation/http/ui/dist`.
