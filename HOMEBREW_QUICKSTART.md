# Быстрый старт: Публикация в Homebrew

## 🎯 Минимальные шаги для публикации

### 1. Создать homebrew-tap репозиторий

```bash
# На GitHub создайте новый public репозиторий
# Имя: homebrew-tap
# URL: https://github.com/seredavin/homebrew-tap
```

### 2. Создать Personal Access Token

1. GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Generate new token (classic)
3. Название: "GoReleaser Homebrew"
4. Права: выберите `repo` (полный доступ)
5. Сохраните токен в безопасном месте

### 3. Добавить токен в GitHub Secrets

1. Ваш репозиторий kafka-test → Settings → Secrets and variables → Actions
2. New repository secret
3. Name: `HOMEBREW_TAP_GITHUB_TOKEN`
4. Secret: вставьте токен из шага 2
5. Add secret

### 4. Включить GoReleaser workflow

У вас уже есть два workflow для релизов:
- `release.yml` - текущий (ручная сборка)
- `release-goreleaser.yml` - новый (с Homebrew tap)

**Вариант A: Использовать оба (рекомендуется):**
- Оставить оба workflow активными
- Переименовать jobs чтобы не конфликтовали

**Вариант B: Только GoReleaser:**
```bash
# Удалить старый workflow
git rm .github/workflows/release.yml
git commit -m "chore: switch to GoReleaser for releases"
```

### 5. Создать следующий релиз

```bash
# Внесите изменения
git add .
git commit -m "feat: add --version and --help flags"

# Создайте тег
git tag -a v1.0.2 -m "Release v1.0.2"
git push origin main
git push origin v1.0.2
```

### 6. Что произойдет

GoReleaser автоматически:
1. ✅ Соберет бинарные файлы для всех платформ
2. ✅ Создаст релиз на GitHub
3. ✅ Сгенерирует Homebrew формулу
4. ✅ Опубликует формулу в seredavin/homebrew-tap
5. ✅ Пользователи смогут установить через brew!

### 7. Установка пользователями

После публикации:

```bash
brew tap seredavin/tap
brew install kafka-producer-ui

# Или одной командой:
brew install seredavin/tap/kafka-producer-ui
```

## 🔍 Проверка

После релиза проверьте:

1. Релиз создан: https://github.com/seredavin/kafka-test/releases
2. Формула опубликована: https://github.com/seredavin/homebrew-tap/blob/main/Formula/kafka-producer-ui.rb

## 📝 Текущий статус

✅ Добавлен флаг `--version`
✅ Добавлен флаг `--help`
✅ Создан `.goreleaser.yml` конфиг
✅ Создан workflow `release-goreleaser.yml`
✅ Создана базовая формула `kafka-producer-ui.rb` (для примера)

## ⏭️ Следующие шаги

1. Создайте репозиторий `seredavin/homebrew-tap` на GitHub
2. Добавьте `HOMEBREW_TAP_GITHUB_TOKEN` в secrets
3. Решите какой workflow использовать (оба или только GoReleaser)
4. Создайте релиз v1.0.2 с поддержкой --version
5. GoReleaser автоматически опубликует в ваш tap!

---

**Готовы опубликовать в Homebrew!** 🍺

