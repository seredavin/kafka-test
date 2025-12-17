# Changelog

Все значимые изменения в этом проекте будут документированы в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.0.0/),
и этот проект придерживается [Semantic Versioning](https://semver.org/lang/ru/).

## [Unreleased]

## [1.0.3] - 2024-12-17

### Добавлено
- MIT License файл

### Исправлено
- GoReleaser теперь включает LICENSE в релизные архивы

## [1.0.2] - 2024-12-17

### Добавлено
- Поддержка флагов командной строки: `--version` и `--help`
- Конфигурация GoReleaser для автоматизации релизов
- Автоматическая публикация в Homebrew tap
- GoReleaser workflow для GitHub Actions
- Документация по публикации в Homebrew

### Изменено
- Release workflow теперь использует GoReleaser
- Автоматическая генерация Homebrew формулы

## [1.0.1] - 2024-12-17

### Добавлено
- GitHub Actions workflows для CI/CD
- Автоматическая сборка релизов для множества платформ (Linux, macOS, Windows)
- Конфигурация golangci-lint
- Документация по процессу релизов (RELEASE.md)
- Руководство для контрибьюторов (CONTRIBUTING.md)
- Шаблоны для issues и Pull Requests
- Makefile для упрощения разработки
- Dockerfile для контейнеризации
- Dependabot для автоматического обновления зависимостей

### Изменено
- Обновлен README с инструкциями по установке из релизов
- Обновлены зависимости через Dependabot (IBM/sarama, actions, и др.)

### Исправлено
- Исправлены все ошибки golangci-lint (errcheck, goconst, fieldalignment)
- Добавлена поддержка Windows в тестах (HOME/USERPROFILE)
- Исправлена проверка file permissions для Windows
- Использование bash shell в CI для корректной работы на Windows
- Настроена совместимость golangci-lint с Go 1.24
- Добавлены permissions для security scan в GitHub Actions
- Улучшена кроссплатформенность (Linux, macOS, Windows)

## [1.0.0] - 2024-12-17

### Добавлено
- Псевдографический интерфейс (TUI) на базе Bubble Tea
- Поддержка mTLS (Mutual TLS) аутентификации
- История отправленных сообщений
- Валидация и форматирование JSON (F10)
- Сохранение конфигурации в `~/.kafka-producer.json`
- Конфигурация подключения к Kafka brokers
- Отправка сообщений в Kafka топики
- Поддержка message key и value
- Горячие клавиши для быстрой навигации:
  - F2: переключение между экранами
  - F5: подключение к Kafka
  - F9: сохранение конфигурации
  - F10: форматирование JSON
  - Tab/Shift+Tab: навигация по полям
  - Enter: отправка сообщения
  - Esc: выход
- Полное покрытие тестами (84.2%)
- Модульные тесты для всех компонентов

### Технические детали
- Использование IBM/sarama для работы с Kafka
- TUI на базе Bubble Tea framework
- Стилизация с помощью Lipgloss
- Поддержка Go 1.21+

---

## Типы изменений

- `Добавлено` - для новой функциональности
- `Изменено` - для изменений в существующей функциональности
- `Устарело` - для функциональности, которая скоро будет удалена
- `Удалено` - для удаленной функциональности
- `Исправлено` - для исправления ошибок
- `Безопасность` - для изменений, связанных с безопасностью

[Unreleased]: https://github.com/seredavin/kafka-test/compare/v1.0.3...HEAD
[1.0.3]: https://github.com/seredavin/kafka-test/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/seredavin/kafka-test/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/seredavin/kafka-test/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/seredavin/kafka-test/releases/tag/v1.0.0

