# Auth Service

Микросервис для аутентификации и авторизации пользователей с использованием JWT токенов.

## Описание

Сервис предоставляет API для регистрации и аутентификации пользователей, интегрируясь с микросервисом users. После успешной аутентификации или регистрации генерируется JWT токен, который может быть использован для доступа к защищенным ресурсам через API Gateway.

## Требования

- Go 1.21 или выше
- Микросервис users (для аутентификации и регистрации)
- API Gateway (для проверки JWT токенов)

## Установка и запуск

1. Клонируйте репозиторий
2. Создайте файл `.env` в корневой директории проекта со следующими переменными:
   ```env
   AUTH_PORT=8081
   JWT_SECRET=your-secret-key-here
   USERS_SERVICE_URL=http://localhost:8082
   ```
3. Установите зависимости:
   ```bash
   go mod tidy
   ```
4. Запустите сервис:
   ```bash
   go run main.go
   ```

## API Endpoints

### Регистрация пользователя

```http
POST /auth/register
Content-Type: application/json

{
    "username": "user123",
    "email": "user@example.com",
    "password": "password123"
}
```

Успешный ответ:
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Вход пользователя

```http
POST /auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

Успешный ответ:
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Валидация данных

- Email: должен быть валидным email адресом
- Пароль: минимум 6 символов
- Имя пользователя: минимум 3 символа

## Интеграция

Сервис интегрируется с:
- Микросервисом users для аутентификации и регистрации
- API Gateway для проверки JWT токенов

## Безопасность

- Все пароли хранятся в зашифрованном виде
- JWT токены подписываются с использованием HS256
- Срок действия токена - 24 часа
- Валидация входных данных
- Защита от SQL-инъекций

## Обработка ошибок

Сервис возвращает следующие коды ошибок:
- 400 Bad Request - неверный формат данных
- 401 Unauthorized - неверные учетные данные
- 500 Internal Server Error - внутренняя ошибка сервера

## Логирование

Сервис использует стандартное логирование Go для отслеживания:
- Запросов к API
- Ошибок аутентификации
- Ошибок интеграции с другими сервисами

## Мониторинг

Для мониторинга доступны следующие метрики:
- Количество успешных/неуспешных аутентификаций
- Время ответа сервиса
- Количество активных сессий 