# Быстрый старт

## Запуск программы

```bash
./kafka-producer
```

## Первое использование без mTLS

1. Программа запустится на экране конфигурации
2. Введите адрес Kafka broker (например: `localhost:9092`)
3. Нажмите `Tab` и введите название топика (например: `test-topic`)
4. Оставьте поля сертификатов пустыми для подключения без mTLS
5. Нажмите `F5` для подключения
6. После успешного подключения нажмите `F2` для перехода к отправке сообщений
7. Введите сообщение и нажмите `Enter`

## Использование с mTLS

1. Подготовьте сертификаты (см. раздел в README.md)
2. Запустите программу: `./kafka-producer`
3. Заполните поля:
   - Brokers: `localhost:9093` (или ваш адрес)
   - Topic: `secure-topic`
   - Client Certificate Path: `/path/to/client-cert.pem`
   - Client Key Path: `/path/to/client-key.pem`
   - CA Certificate Path: `/path/to/ca-cert.pem`
4. Нажмите `F9` чтобы сохранить конфигурацию
5. Нажмите `F5` для подключения
6. После успешного подключения нажмите `F2`
7. Отправляйте сообщения

## Отправка JSON сообщений

1. Перейдите на экран отправки (`F2`)
2. Переместитесь на поле "Message Value" (`Tab`)
3. Введите JSON (может быть неформатированным): `{"name":"test","value":123}`
4. Нажмите `F10` для форматирования JSON
5. Нажмите `Enter` для отправки

## Навигация

- `Tab` / `Shift+Tab` - переключение между полями
- `F2` - переключение между экранами
- `F5` - подключение к Kafka
- `F9` - сохранение конфигурации
- `F10` - форматирование JSON
- `Enter` - отправка сообщения
- `Esc` - выход

## Проверка работы

После отправки сообщения вы увидите:
- ✅ "Success" - сообщение отправлено успешно
- ❌ "Failed: ..." - ошибка при отправке

Для успешных сообщений отображается partition и offset.

## Устранение проблем

### "Connection refused"
- Убедитесь, что Kafka запущена
- Проверьте правильность адреса broker

### "Failed to load client certificate"
- Проверьте пути к сертификатам
- Убедитесь, что файлы существуют и доступны для чтения
- Проверьте формат сертификатов (должны быть PEM)

### "Invalid JSON"
- Проверьте синтаксис JSON перед отправкой
- Используйте `F10` для валидации и форматирования

## Тестирование локально

Для быстрого тестирования можно запустить Kafka в Docker:

```bash
# Запуск Kafka без авторизации
docker run -d --name kafka \
  -p 9092:9092 \
  -e KAFKA_ENABLE_KRAFT=yes \
  -e KAFKA_CFG_NODE_ID=1 \
  -e KAFKA_CFG_PROCESS_ROLES=broker,controller \
  -e KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER \
  -e KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \
  -e KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT \
  -e KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
  bitnami/kafka:latest
```

Затем запустите программу и используйте:
- Broker: `localhost:9092`
- Topic: `test-topic` (создастся автоматически)
