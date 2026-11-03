# fluffy-giggle

Прототип пайплайна «слабая LLM + сильный adviser» для обращений с персональными данными.
Как устроено — [docs/architecture/overview.md](docs/architecture/overview.md), правила — [AGENTS.md](AGENTS.md).

```
make env      # core/.env из примера, впишите ключи моделей
make run      # бэкенд на :8000
make web      # фронт на :5173
make build    # один бинарь со вшитым фронтом в core/bin/
```
