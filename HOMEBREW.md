# Публикация в Homebrew

Это руководство описывает как опубликовать kafka-producer-ui в Homebrew.

## Подготовка

### 1. Добавить поддержку --version

Обновите `main.go`:

```go
var version = "dev" // будет заменено при сборке ldflags

func main() {
    // Проверка флага --version
    if len(os.Args) > 1 && os.Args[1] == "--version" {
        fmt.Printf("kafka-producer-ui version %s\n", version)
        os.Exit(0)
    }
    
    // ... остальной код
}
```

### 2. Создать репозиторий для Homebrew tap

1. На GitHub создайте новый репозиторий: `homebrew-tap`
2. Полное имя будет: `seredavin/homebrew-tap`

```bash
# Клонировать
git clone https://github.com/seredavin/homebrew-tap
cd homebrew-tap

# Создать структуру
mkdir Formula
```

## Метод 1: Ручная публикация

### Шаг 1: Получить SHA256 checksums

После создания релиза, скачайте `checksums.txt` или вычислите:

```bash
# Скачать архивы из релиза
wget https://github.com/seredavin/kafka-test/releases/download/v1.0.1/checksums.txt

# Или вычислить локально
shasum -a 256 kafka-producer-ui-darwin-amd64.tar.gz
shasum -a 256 kafka-producer-ui-darwin-arm64.tar.gz
shasum -a 256 kafka-producer-ui-linux-amd64.tar.gz
shasum -a 256 kafka-producer-ui-linux-arm64.tar.gz
```

### Шаг 2: Обновить формулу

Отредактируйте `kafka-producer-ui.rb`, замените `REPLACE_WITH_ACTUAL_SHA256_*` на реальные checksums.

### Шаг 3: Опубликовать

```bash
cd homebrew-tap
cp /path/to/kafka-producer-ui.rb Formula/
git add Formula/kafka-producer-ui.rb
git commit -m "Add kafka-producer-ui v1.0.1"
git push
```

## Метод 2: Автоматизация с GoReleaser (Рекомендуется)

### Шаг 1: Установить GoReleaser

```bash
# macOS
brew install goreleaser

# Linux
go install github.com/goreleaser/goreleaser@latest
```

### Шаг 2: Настроить токены

```bash
# Создайте Personal Access Token на GitHub с правами:
# - repo (full control)
# - write:packages

# Добавьте в GitHub Secrets:
# HOMEBREW_TAP_GITHUB_TOKEN

# Или локально:
export GITHUB_TOKEN=your_github_token
export HOMEBREW_TAP_GITHUB_TOKEN=your_github_token
```

### Шаг 3: Обновить GitHub Actions workflow

Создайте `.github/workflows/release.yml` (уже существует, нужно обновить):

```yaml
- name: Run GoReleaser
  uses: goreleaser/goreleaser-action@v5
  with:
    version: latest
    args: release --clean
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

### Шаг 4: Создать релиз

```bash
# GoReleaser автоматически:
# 1. Создаст релиз на GitHub
# 2. Соберет бинарные файлы
# 3. Создаст архивы и checksums
# 4. Сгенерирует и опубликует Homebrew формулу в ваш tap

git tag -a v1.0.2 -m "Release v1.0.2"
git push origin v1.0.2
```

## Установка пользователями

### Из вашего tap:

```bash
brew tap seredavin/tap
brew install kafka-producer-ui
```

### Или одной командой:

```bash
brew install seredavin/tap/kafka-producer-ui
```

## Обновление формулы

### При новом релизе:

Если используете GoReleaser - всё автоматически!

Если вручную:

```bash
cd homebrew-tap
# Обновите version и sha256 в Formula/kafka-producer-ui.rb
git commit -am "Update kafka-producer-ui to v1.0.2"
git push
```

Пользователи обновят:

```bash
brew update
brew upgrade kafka-producer-ui
```

## Публикация в официальный Homebrew-core

После стабилизации проекта (обычно после нескольких релизов), можно подать PR в официальный репозиторий:

### Требования:

1. ✅ Проект должен быть стабильным (несколько релизов)
2. ✅ Активная разработка (commits, issues, PR)
3. ✅ Хорошая документация
4. ✅ Лицензия (MIT ✅)
5. ✅ Тесты (84.2% coverage ✅)
6. ✅ CI/CD (GitHub Actions ✅)

### Процесс:

```bash
# Форкните homebrew-core
git clone https://github.com/Homebrew/homebrew-core

# Создайте формулу
cd homebrew-core/Formula
brew create https://github.com/seredavin/kafka-test/archive/v1.0.1.tar.gz

# Отредактируйте формулу
# Тестируйте
brew install --build-from-source kafka-producer-ui
brew test kafka-producer-ui
brew audit --strict kafka-producer-ui

# Создайте PR
git checkout -b kafka-producer-ui
git add Formula/kafka-producer-ui.rb
git commit -m "kafka-producer-ui 1.0.1 (new formula)"
git push origin kafka-producer-ui
```

## Полезные команды

```bash
# Проверить формулу
brew audit --strict Formula/kafka-producer-ui.rb

# Тестировать локально
brew install --build-from-source ./kafka-producer-ui.rb

# Тестировать установку
brew test kafka-producer-ui

# Удалить
brew uninstall kafka-producer-ui
brew untap seredavin/tap
```

## Ресурсы

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [GoReleaser Documentation](https://goreleaser.com/intro/)
- [Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
- [Node for Package Maintainers](https://docs.brew.sh/Node-for-Formula-Authors)

