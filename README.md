# AIagent

Учебный pet-проект: AI-агент на **Go**, реализующий паттерны ReAct и Plan-and-Execute с вызовом инструментов (function calling) и подключением к PostgreSQL.

Большинство AI-агентов пишут на Python (LangChain, LangGraph и аналоги). Здесь — эксперимент с Go: типизация, простой деплой одним бинарником и знакомство с LLM-интеграцией без тяжёлого фреймворка.

> **Статус:** ранняя разработка. Архитектура и API будут меняться по мере доработки.

## Возможности

| Компонент | Описание |
|-----------|----------|
| **ReAct** | Цикл *think → act → observe*: модель рассуждает, предлагает ответ или вызывает инструмент, получает наблюдение и итеративно уточняет результат |
| **Plan-and-Execute** | Планировщик разбивает задачу на шаги, исполнитель выполняет их по одному |
| **Function calling** | Модель может вызвать зарегистрированные инструменты (сейчас — `GetContributors`) |
| **PostgreSQL** | Хранение данных для инструментов агента |
| **Gemini API** | Запросы через OpenAI-совместимый endpoint Google AI |

## Стек

- **Go** 1.25
- **PostgreSQL** 15 (Docker)
- [pgx](https://github.com/jackc/pgx) — драйвер БД
- [godotenv](https://github.com/joho/godotenv) — переменные окружения
- **Google Gemini** (`gemini-3.5-flash`) через `generativelanguage.googleapis.com`

## Быстрый старт

### Требования

- Go 1.25+
- Docker и Docker Compose
- API-ключ [Google AI Studio](https://aistudio.google.com/)

### 1. Клонирование

```bash
git clone git@github.com:SSSmilar/AiAgent.git
cd AiAgent
```

### 2. Переменные окружения

Создайте файл `.env` в корне проекта:

```env
OPENAI_API_KEY=ваш_ключ_google_ai_studio
DATABASE_URL=postgres://agent_user:agent_password@localhost:5432/agent_db?sslmode=disable
```

> Ключ называется `OPENAI_API_KEY`, потому что используется OpenAI-совместимый API Gemini.

### 3. База данных

```bash
docker compose up -d
```

При первом запуске выполнится `init.sql` — создаётся таблица `gitRepo` с тестовыми данными.

### 4. Запуск

```bash
go run .
```

Агент последовательно выполнит планирование (`Plan`) и ReAct-цикл (`ReAct`) для заданной задачи.

## Архитектура

```
┌─────────────┐     HTTP (OpenAI API)     ┌──────────────────┐
│   main.go   │ ─────────────────────────► │  Gemini API      │
│  plan.go    │                            └──────────────────┘
│  reAct.go   │
└──────┬──────┘
       │ SQL
       ▼
┌─────────────┐
│ PostgreSQL  │
│  (gitRepo)  │
└─────────────┘
```

### Файлы

| Файл | Назначение |
|------|------------|
| `main.go` | Точка входа, типы сообщений/инструментов, HTTP-клиент к LLM, `GetContributors` |
| `plan.go` | Паттерн Plan-and-Execute: планировщик + пошаговый исполнитель |
| `reAct.go` | ReAct-цикл с поддержкой tool calls и ограничением итераций |
| `init.sql` | Схема и начальные данные БД |
| `docker-compose.yml` | PostgreSQL для локальной разработки |

### ReAct-цикл

На каждой итерации модель выводит блок:

```
Thought: <рассуждение>
Action: PROPOSE: <вариант ответа>
   — или —
Action: FINAL: <финальный ответ>
```

После `PROPOSE` агент отправляет `Observation` с просьбой перепроверить. Цикл завершается при `FINAL` или после 5 итераций.

### Инструменты

Сейчас зарегистрирован один инструмент:

- **`GetContributors`** — возвращает список контрибьюторов из таблицы `gitRepo` (`login`, `commit_count`)

## Разработка

### Линтер

```bash
# через Task
task lint

# или напрямую
golangci-lint run ./...
```

В CI на GitHub Actions запускается `golangci-lint` при push и pull request в `main`.

## Дорожная карта

Проект в активной разработке. Планируется:

- [ ] Рефакторинг: разделение на пакеты (`agent`, `tools`, `llm`, `storage`)
- [ ] Расширяемый реестр инструментов
- [ ] Конфигурация модели и лимитов через env / CLI-флаги
- [ ] Новые инструменты (веб-поиск, файловая система и др.)
- [ ] Тесты для ключевых компонентов
- [ ] Документация по добавлению своих инструментов

## Автор

**Andrey** — [github.com/SSSmilar](https://github.com/SSSmilar)

## Лицензия

Все права защищены. Использование, копирование и распространение — только с разрешения автора.
