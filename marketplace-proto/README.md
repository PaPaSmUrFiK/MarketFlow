# Marketplace Proto

Protobuf определения для Identity сервиса.

## Требования

- Go 1.23+
- protoc (Protocol Buffers compiler)
- protoc-gen-go
- protoc-gen-go-grpc
- Task (taskfile.dev)

## Установка зависимостей

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Генерация кода

Для генерации Go кода из proto файлов:

```bash
task generate
```

или

```bash
task gen
```

## Очистка сгенерированных файлов

```bash
task clean
```

## Структура проекта

```
.
├── proto/
│   └── identity/
│       └── v1/
│           ├── admin.proto    # Административные операции
│           ├── auth.proto     # Аутентификация
│           ├── common.proto   # Общие типы
│           └── user.proto     # Пользовательские операции
└── gen/
    └── go/
        └── identity/
            └── v1/            # Сгенерированный Go код
```

## Сервисы

### AdminService
- Управление приложениями
- Управление пользователями
- Управление ролями
- Управление правами доступа

### AuthService
- Регистрация
- Вход
- OAuth аутентификация
- Обновление токенов
- Выход

### UserService
- Получение информации о текущем пользователе
- Получение информации о пользователе по ID
- Изменение пароля
