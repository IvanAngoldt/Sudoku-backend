# Микросервис для управления пользователями

Этот микросервис предоставляет REST API для управления пользователями, включая аутентификацию, управление профилем и статистикой.

## API Endpoints

### Основные операции с пользователями
- `GET /` - Получить список всех пользователей
- `GET /{id}` - Получить пользователя по ID
- `POST /` - Создать нового пользователя
- `PATCH /{id}` - Обновить существующего пользователя
- `DELETE /{id}` - Удалить пользователя

### Аутентификация и профиль
- `POST /auth` - Аутентификация пользователя
- `GET /me` - Получить информацию о текущем пользователе
- `GET /me/info` - Получить детальную информацию о текущем пользователе
- `PATCH /me/info` - Обновить информацию о текущем пользователе

### Проверка данных
- `GET /check-username` - Проверить доступность имени пользователя
- `GET /check-email` - Проверить доступность email

### Аватары
- `POST /me/avatar` - Загрузить аватар пользователя
- `GET /{id}/avatar` - Получить аватар пользователя

### Статистика
- `GET /{id}/statistics` - Получить статистику пользователя
- `PATCH /{id}/statistics` - Обновить статистику пользователя

## Структура пользователя

```json
{
    "id": "string",
    "username": "string",
    "email": "string",
    "password": "string",
    "avatar_url": "string",
    "statistics": {
        "games_played": "integer",
        "games_won": "integer",
        "average_time": "float"
    }
}
```

## Установка и запуск

1. Установите Go (версия 1.21 или выше)
2. Клонируйте репозиторий
3. Перейдите в директорию проекта
4. Установите зависимости:
   ```bash
   go mod download
   ```
5. Создайте файл конфигурации `.env` со следующими параметрами:
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your_password
   DB_NAME=users_db
   SERVER_PORT=8080
   ```
6. Запустите сервер:
   ```bash
   go run main.go
   ```

Сервер будет доступен по адресу: http://localhost:8082 