# Тестовое задание на стажировку от Selectel Career Wave 2026

Позиция: Backend-разработка. Golang

**Васильев Савелий**
**Контакты:** 
- **email:** ust1vasiljev@yandex.ru
- **telegram:** @ustithegoddd

## Реализация
Реализованы все пункты из ТЗ, а так же все бонусные задания:
✅ Реализованы все правила для линтера лог-записей
✅ Соблюдены технические требования:
- Go v1.25.5
- Линтер поддерживается в виде плагина для `golangci-lint`
- Обеспечена поддержка `log/slog` и `go.uber.org/zap`
✅ Проект покрыт тестами, а так же есть возможность протестировать линтер на реальном проекте

Бонусные задания:
✅ Конфигурация правил через `.golangci.yml`
✅ `SuggestedFixes` для всех правил
✅ Кастомные ключевые слова для чувствительных данных в `.golangci.yml`
✅ CI пайплайн в Github Actions

## Инструкция по установке 

Линтер реализован в виде `module plugin` для `golangci-lint`.
Все нужные команды для сборки выполняются через утилиту `Taskfile`.

1. Соберите кастомный бинарник `golangci-lint`:

```bash
task custom-gcl
```

`golangci-lint custom` использует файл `.custom-gcl.yml` в корне проекта.

2. Запустите линтер:

```bash
task lint
```

## Конфигурация

Настройки задаются в `.golangci.yml`:

```yaml
version: "2"
linters:
  default: none
  settings:
    custom:
      logmessagelint:
        type: module
        description: "Log message linter"
        original-url: "github.com/ustithegod/log-linter"
        settings:
          rules:
            lowercase_start: true
            english_only: true
            no_specials: true
            sensitive: true
          sensitive:
            case_insensitive: true
            word_boundary: true
            keywords:
              - password
              - passwd
              - pass
              - token
              - api_key
              - apikey
              - secret
              - private_key
  enable:
    - logmessagelint
```

Конфигурируется:
- то, какие правила будут применяться при линте (rules)
- чувствительность к регистру у чувствительных данных (case_insensitive)
- проверка на целое слово (word_boundary)
- список ключевых слов для линта (keywords)

## Как работает проверка чувствительных данных

Линтер анализирует идентификаторы в выражениях лог‑сообщений (например, при конкатенации строк).
Совпадение ищется по ключевым словам и учитывает `word_boundary` и `case_insensitive`.

Примеры:

```go
token := "t"
passwordHash := "h"
user_token := "u"

log.Info("token: " + token)      // найдено: token
log.Info("password: " + passwordHash) // найдено: password
log.Info("user token: " + user_token) // найдено: token
```

## Разработка

Запуск тестов:

```bash
task test
```

Сборка standalone бинарника:

```bash
task build-linter
```

## Тестирование на реальном проекте

В корне проекта в директории test_project находится проект, который использует slog в качестве logger'а. Внутри логов были допущены ошибки, которые ловит реализованный линтер.

Команда для запуска тестов:

```bash
task lint-test-project
```

## Структура проекта

- `analyzer/` — анализатор и правила.
- `plugin/` — модульный плагин для `golangci-lint`.
- `cmd/loglinter/` — standalone запуск через `singlechecker`.
