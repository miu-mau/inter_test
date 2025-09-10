### В НАЧАЛЕ РАБОТЫ

после клонирования проекта ввести в терминале:

```

go mod tidy
```

для запуска сервера ввести в терминале:

```

go run main.go
```

примеры запросов в Postman ПРИ url: http://localhost:8080:

## POST /tasks

Запрос

```

POST http://localhost:8080/tasks

{
    "title": "Купить хлеб",
    "description": "Уточнить свежесть",
    "completed": false
}
```

Ответ

```

201 Created
Content-Type: application/json

{
    "id": 1,
    "title": "Купить хлеб",
    "description": "Уточнить свежесть",
    "completed": false
}
```

## GET /tasks

Запрос

```

GET http://localhost:8080/tasks
```

Ответ

```

200 OK
Content-Type: application/json

[
    {
        "id": 2,
        "title": "Купить молоко",
        "description": "купить две шткуи",
        "completed": false
    },
    {
        "id": 1,
        "title": "Купить хлеб",
        "description": "Уточнить свежесть",
        "completed": false
    }
]
```

## GET /tasks/{id}

Запрос

```

GET http://localhost:8080/tasks/1
```

Ответ

```

200 OK
Content-Type: application/json

{
    "id": 1,
    "title": "Купить хлеб",
    "description": "Уточнить свежесть",
    "completed": false
}
```

Запрос с несуществующим id

```

GET http://localhost:8080/tasks/9
```

Ответ

```

404 Not Found
Content-Type: application/json

{
    "error": "task not found"
}
```

## PUT /tasks/{id}

Запрос

```

PUT http://localhost:8080/tasks/1

{
    "title": "Продать хлеб",
    "description": "да побольше",
    "completed": true
}
```

Ответ

```

200 OK
Content-Type: application/json
{
    "id": 1,
    "title": "Продать хлеб",
    "description": "да побольше",
    "completed": true
}
```
## DELETE /tasks/{id}

Запрос

```

DELETE http://localhost:8080/tasks/1
```

Ответ

```

204 No Content
```
