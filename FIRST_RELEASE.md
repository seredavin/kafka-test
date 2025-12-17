# Руководство по первому релизу

Это краткое руководство поможет вам создать первый релиз проекта.

## ✅ Что уже настроено

### GitHub Actions Workflows

✅ **CI Workflow** (`.github/workflows/ci.yml`)
- Автоматическое тестирование на Linux, macOS, Windows
- Проверка линтером (golangci-lint)
- Security scanning (gosec)
- Тестирование на Go 1.21, 1.22, 1.23
- Автоматическая загрузка артефактов сборки

✅ **Release Workflow** (`.github/workflows/release.yml`)
- Автоматическая сборка при создании тега
- Сборка для 6 платформ:
  - Linux (AMD64, ARM64)
  - macOS (AMD64, ARM64)
  - Windows (AMD64, ARM64)
- Создание архивов и контрольных сумм
- Автоматическая публикация релиза на GitHub

### Инфраструктурные файлы

✅ Конфигурация линтера (`.golangci.yml`)
✅ Dependabot для автообновления зависимостей
✅ Шаблоны для issues и Pull Requests
✅ CODEOWNERS для автоматического review
✅ Makefile для упрощения разработки
✅ Dockerfile для запуска в контейнере
✅ Полная документация (README, RELEASE, CONTRIBUTING, CHANGELOG)

## 🚀 Шаги для создания первого релиза

### 1. Обновите README.md

Замените `USERNAME` в README.md на ваше GitHub имя пользователя:

```bash
# Замените USERNAME на ваше имя
sed -i '' 's/USERNAME/your-github-username/g' README.md
sed -i '' 's/USERNAME/your-github-username/g' CHANGELOG.md
sed -i '' 's/USERNAME/your-github-username/g' .github/CODEOWNERS
sed -i '' 's/USERNAME/your-github-username/g' .github/dependabot.yml
```

Или вручную отредактируйте файлы:
- `README.md` - замените все `USERNAME`
- `CHANGELOG.md` - замените ссылки
- `.github/CODEOWNERS` - укажите ваш GitHub username
- `.github/dependabot.yml` - укажите ваш username для reviewers

### 2. Закоммитьте все изменения

```bash
# Добавить все новые файлы
git add .

# Проверить что добавлено
git status

# Закоммитить
git commit -m "chore: add CI/CD workflows and project infrastructure

- Add GitHub Actions workflows for CI and releases
- Add golangci-lint configuration
- Add Makefile for common tasks
- Add Dockerfile for containerization
- Add comprehensive documentation (RELEASE.md, CONTRIBUTING.md)
- Add issue and PR templates
- Add Dependabot configuration
- Add CHANGELOG.md for version tracking"

# Отправить в GitHub
git push origin main
```

### 3. Дождитесь прохождения CI

После push проверьте:
- Перейдите в раздел **Actions** вашего GitHub репозитория
- Убедитесь, что CI workflow прошел успешно (зеленая галочка ✅)
- Если есть ошибки, исправьте их

### 4. Создайте первый релиз

```bash
# Создайте тег для версии 1.0.0
git tag -a v1.0.0 -m "Release v1.0.0

Initial release of Kafka Producer UI

Features:
- Terminal UI for sending messages to Kafka
- mTLS authentication support
- Message history
- JSON validation and formatting
- Configuration persistence
- Multiple SerDe formats (string, json, bytearray)

Test coverage: 84.2%"

# Отправьте тег в GitHub
git push origin v1.0.0
```

### 5. Следите за процессом сборки

1. Перейдите в **Actions** → **Release workflow**
2. Дождитесь завершения сборки (обычно 2-5 минут)
3. Проверьте, что все шаги прошли успешно

### 6. Проверьте релиз

1. Перейдите в **Releases** вашего репозитория
2. Вы должны увидеть релиз **v1.0.0** с:
   - 6 бинарных архивов (.tar.gz и .zip)
   - Файл checksums.txt
   - Release notes

### 7. Протестируйте релиз

Скачайте бинарный файл для вашей платформы и протестируйте:

**macOS (Apple Silicon):**
```bash
curl -LO https://github.com/YOUR-USERNAME/kafka-test/releases/download/v1.0.0/kafka-producer-ui-darwin-arm64.tar.gz
tar -xzf kafka-producer-ui-darwin-arm64.tar.gz
chmod +x kafka-producer-ui-darwin-arm64
./kafka-producer-ui-darwin-arm64
```

**Linux:**
```bash
wget https://github.com/YOUR-USERNAME/kafka-test/releases/download/v1.0.0/kafka-producer-ui-linux-amd64.tar.gz
tar -xzf kafka-producer-ui-linux-amd64.tar.gz
chmod +x kafka-producer-ui-linux-amd64
./kafka-producer-ui-linux-amd64
```

## 🛠️ Использование Makefile

После первого релиза вы можете использовать Makefile для упрощения разработки:

```bash
# Показать все доступные команды
make help

# Собрать проект
make build

# Запустить тесты
make test

# Проверить покрытие
make coverage

# Запустить линтер
make lint

# Собрать для всех платформ
make build-all

# Выполнить все проверки
make check

# Очистить артефакты
make clean
```

## 📊 Следующие шаги

### Настройка Codecov (опционально)

Если хотите отслеживать покрытие кода:

1. Зарегистрируйтесь на [codecov.io](https://codecov.io)
2. Добавьте ваш репозиторий
3. Добавьте `CODECOV_TOKEN` в GitHub Secrets:
   - Settings → Secrets and variables → Actions → New repository secret
   - Name: `CODECOV_TOKEN`
   - Value: токен из Codecov

### Включение защиты ветки main

Рекомендуется защитить ветку main:

1. Settings → Branches → Add rule
2. Branch name pattern: `main`
3. Включите:
   - ✅ Require a pull request before merging
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
   - Выберите checks: CI / Test

### Добавление тем для GitHub

Settings → General → Topics → Add topics:
- `kafka`
- `golang`
- `tui`
- `terminal`
- `cli`
- `mtls`
- `producer`

## 🎯 Создание последующих релизов

Для создания новых версий:

```bash
# Patch release (1.0.0 → 1.0.1)
git tag -a v1.0.1 -m "Bug fixes"
git push origin v1.0.1

# Minor release (1.0.0 → 1.1.0)
git tag -a v1.1.0 -m "New features"
git push origin v1.1.0

# Major release (1.0.0 → 2.0.0)
git tag -a v2.0.0 -m "Breaking changes"
git push origin v2.0.0
```

Не забудьте обновить CHANGELOG.md перед каждым релизом!

## ❓ Возможные проблемы

### CI падает с ошибкой линтера

```bash
# Запустите линтер локально
make lint

# Или
golangci-lint run
```

Исправьте все найденные проблемы.

### Тесты не проходят

```bash
# Запустите тесты локально
make test

# С подробным выводом
go test -v ./...
```

### Release workflow не запускается

Убедитесь что:
- Тег начинается с `v` (v1.0.0, не 1.0.0)
- Вы push'нули тег: `git push origin v1.0.0`
- У вас есть права на создание релизов в репозитории

## 📚 Дополнительные ресурсы

- [RELEASE.md](RELEASE.md) - Подробное руководство по релизам
- [CONTRIBUTING.md](CONTRIBUTING.md) - Руководство для контрибьюторов
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Semantic Versioning](https://semver.org/)

---

**Готовы к первому релизу? Удачи! 🚀**

После создания релиза вы можете удалить этот файл:
```bash
git rm FIRST_RELEASE.md
git commit -m "chore: remove first release guide"
git push
```

