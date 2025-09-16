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


## Аутентификация (JWT)

Регистрация → вход → использование access токена → обновление → выход.

### POST /api/auth/register

Запрос

```
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "username": "john",
  "email": "john@example.com",
  "password": "secret123"
}
```

Ответ

```
200 OK
Content-Type: application/json

{
  "id": 1,
  "username": "john",
  "email": "john@example.com",
  "isActive": true
}
```

### POST /api/auth/login

Запрос

```
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{
  "username": "john",
  "password": "secret123"
}
```

Ответ

```
200 OK
Content-Type: application/json

{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTg2MzEzMDcsImlhdCI6MTc1ODAyNjUwNywianRpIjoiMjk0NTY5NWM1YmVmNTc3NjU2ODgyYzQ1YTBjNzk1MTAiLCJzdWIiOiIxIiwidHlwZSI6InJlZnJlc2giLCJ1c2VybmFtZSI6ImpvaG4ifQ.5U6-y8ywmR5OqAuo8O9mgUryhc9eQW-AqQDj1gx6exQ",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTgwMjczMjcsImlhdCI6MTc1ODAyNjQyNywic3ViIjoiMSIsInR5cGUiOiJhY2Nlc3MiLCJ1c2VybmFtZSI6ImpvaG4ifQ.0y_XVuouCBlaorxgk4d9wkRebdAE1BcvGYVNCT8Wvv0",
  "tokenType": "Bearer",
  "expiresIn": 900,
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com",
    "isActive": true
  }
}
```

### GET /api/auth/me

Передавайте access токен:

```
GET http://localhost:8080/api/auth/me
Authorization: Bearer <accessToken>
```

Ответ

```
200 OK
{
  "id": 1,
  "username": "john",
  "email": "john@example.com",
  "isActive": true
}
```

### POST /api/auth/refresh

Запрос

```
POST http://localhost:8080/api/auth/refresh
Content-Type: application/json

{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTgwMjczMjcsImlhdCI6MTc1ODAyNjQyNywic3ViIjoiMSIsInR5cGUiOiJhY2Nlc3MiLCJ1c2VybmFtZSI6ImpvaG4ifQ.0y_XVuouCBlaorxgk4d9wkRebdAE1BcvGYVNCT8Wvv0"
}
```

Ответ

```
200 OK
{
  "accessToken": "eyJhbGciOiJJKHJYgjgygjgIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTgwMjczMjcsImc3ViIjoiMSIsInR5cGUiOiJhY2Nlc3MiLCJ1c2VybmFtZSI6ImpvaG4ifQ.0y_XVuouCBlaorxgk4d9wkRebdAE1BcvGYVNCT8Wvv0",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTgwMjczMjcsImlhdCI6HKHuhkVFGGHDGRFDtrdhjkuKJBHUUI83579HBGJGFUMTc1ODAyNjQyNywic3ViIjoiMSIsInR5cGUiOiJhY2Nlc3MiLCJ1c2VybmFtZSI6ImpvaG4ifQ.0y_XVuouCBlaorxgk4d9wkRebdAE1BcvGYVNCT8Wvv0"",
  "tokenType": "Bearer",
  "expiresIn": 900
}
```

### POST /api/auth/logout

Запрос

```
POST http://localhost:8080/api/auth/logout
Content-Type: application/json

{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTgwMjczMjcsImlhdCI6MTc1ODAyNjQyNywic3ViIjoiMSIsInR5cGUiOiJhY2Nlc3MiLCJ1c2VybmFtZSI6ImpvaG4ifQ.0y_XVuouCBlaorxgk4d9wkRebdAE1BcvGYVNCT8Wvv0"
}
```

Ответ

```
200 OK
{}
```
