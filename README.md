### В НАЧАЛЕ РАБОТЫ

после клонирования проекта ввести в терминале:

```

go mod tidy
```

для запуска сервера ввести в терминале:

```

go run .
```

примеры запросов в Postman ПРИ url: http://localhost:8080:


## Админ (управление пользователями)

Регистрация → вход → использование access токена → обновление → выход.

### POST /api/auth/login

Запрос

```

POST http://localhost:8080/api/auth/login
Content-Type: application/json

{ 
    "username": "admin", 
    "password": "admin123"
}
```

Ответ

```

200 OK
Content-Type: application/json

{
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTgxMDcwMjMsImlhdCI6MTc1ODEwNjEyMywic3ViIjoiMSIsInR5cGUiOiJhY2Nlc3MiLCJ1c2VybmFtZSI6ImFkbWluIn0.DN5qOGFWNZSe_c2bcO78R8B9JKqJBXj6C_212R9P0Go",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTg3MTA5MjMsImlhdCI6MTc1ODEwNjEyMywianRpIjoiYzY2NTZjY2QyNzc5MTg2YjAxYzdhZGVkNGEzNzcxMzYiLCJzdWIiOiIxIiwidHlwZSI6InJlZnJlc2giLCJ1c2VybmFtZSI6ImFkbWluIn0.I9WckyNS-zKB4aw8qkd1-Po0LXpOPe74FDwlIGaanog",
    "tokenType": "Bearer",
    "expiresIn": 900,
    "user": {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "isActive": true,
        "isAdmin": true
    }
}
```

## GET /api/users

Запрос

```

GET http://localhost:8080/api/users
Authorization: Bearer <accessToken>
```

Ответ

```

[
    {
        "id": 2,
        "username": "john",
        "email": "john@example.com",
        "isActive": true,
        "isAdmin": false
    },
    {
        "id": 3,
        "username": "emily",
        "email": "emily@example.com",
        "isActive": true,
        "isAdmin": false
    },
    {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "isActive": true,
        "isAdmin": true
    }
]
```

## GET /api/users/{id}

Запрос

```

GET http://localhost:8080/api/users/2
Authorization: Bearer <accessToken>
```

Ответ

```

{
    "id": 2,
    "username": "john",
    "email": "john@example.com",
    "isActive": true,
    "isAdmin": false
}
```

## PUT /api/users/{id}

Запрос

```

PUT http://localhost:8080/api/users/2
Content-Type: application/json,
Authorization: Bearer <accessToken>

{
  "username": "john",
  "email": "johnDoe@example.com",
  "isActive": true,
  "isAdmin": false
}
```

Ответ

```

{
  "username": "john",
  "email": "johnDoe@example.com",
  "isActive": true,
  "isAdmin": false
}
```

## DELETE /api/users{id}

Запрос

```

DELETE http://localhost:8080/api/users/2
Authorization: Bearer <accessToken>
```

Ответ

```

204 NO Content

GET http://localhost:8080/api/users

[
    {
        "id": 3,
        "username": "emily",
        "email": "emily@example.com",
        "isActive": true,
        "isAdmin": false
    },
    {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "isActive": true,
        "isAdmin": true
    }
]
```