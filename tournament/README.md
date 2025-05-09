# Турнирная система для Судоку

Система для проведения турниров по решению судоку. Позволяет создавать турниры, регистрировать участников, отслеживать их прогресс и подводить итоги.

## Структура проекта

```
tournament/
├── config/         # Конфигурация приложения
├── database/       # Работа с базой данных и Redis
├── handlers/       # HTTP-обработчики
├── middleware/     # Промежуточное ПО
├── models/         # Модели данных
├── services/       # Бизнес-логика
├── main.go         # Точка входа
├── go.mod          # Зависимости
└── README.md       # Документация
```

## Функциональность

### Турниры

- Создание турнира
- Получение списка турниров
- Получение информации о турнире
- Обновление информации о турнире
- Удаление турнира

### Участники

- Регистрация участников на турнир
- Отслеживание прогресса участников
- Обновление очков и количества решенных задач

### Управление турниром

- Начало турнира
- Завершение турнира
- Подведение итогов
- Ранжирование участников

## API Endpoints

### Турниры

```
POST   /tournaments              # Создание турнира
GET    /tournaments              # Получение списка турниров
GET    /tournaments/:id          # Получение информации о турнире
PUT    /tournaments/:id          # Обновление информации о турнире
DELETE /tournaments/:id          # Удаление турнира
```

### Управление турниром

```
POST   /tournaments/:id/register # Регистрация участника
POST   /tournaments/:id/start    # Начало турнира
POST   /tournaments/:id/progress # Обновление прогресса
POST   /tournaments/:id/finish   # Завершение турнира
```

## Модели данных

### Tournament

```go
type Tournament struct {
    ID          string           // Уникальный идентификатор
    Name        string           // Название турнира
    Description string           // Описание
    StartTime   time.Time        // Время начала
    EndTime     time.Time        // Время окончания
    Status      TournamentStatus // Статус турнира
    CreatedBy   string           // ID создателя
    CreatedAt   time.Time        // Время создания
}
```

### TournamentParticipant

```go
type TournamentParticipant struct {
    ID           string    // Уникальный идентификатор
    TournamentID string    // ID турнира
    UserID       string    // ID пользователя
    Score        int       // Очки
    SolvedCount  int       // Количество решенных задач
    JoinedAt     time.Time // Время регистрации
    LastSolvedAt time.Time // Время последнего решения
}
```

### TournamentResult

```go
type TournamentResult struct {
    ID           string    // Уникальный идентификатор
    TournamentID string    // ID турнира
    UserID       string    // ID пользователя
    Score        int       // Итоговые очки
    Rank         int       // Место в турнире
    SolvedCount  int       // Количество решенных задач
    FinishedAt   time.Time // Время завершения
}
```

## Статусы турнира

- `pending`   - Ожидает начала
- `active`    - Активный
- `finished`  - Завершен
- `cancelled` - Отменен

## Запуск проекта

1. Установите зависимости:
```bash
go mod download
```

2. Настройте конфигурацию в `config/config.go`

3. Запустите сервер:
```bash
go run main.go
```

## Требования

- Go 1.16+
- PostgreSQL
- Redis 