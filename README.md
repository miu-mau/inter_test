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

### Как это работает (up-down)

- **BuildRouter()**: собирает весь HTTP, подключает обработчики, а также инициализирует хранилища: задачи (Store), пользователи (UsersStore), рефреш‑токены (RefreshStore). Там же выполняется сид администратора (admin/admin123).


- **httptest**: тесты не поднимают реальный порт. Вместо этого `httptest.NewRecorder()` и `httptest.NewRequest()` прогоняют реальные HTTP‑запросы прямо через роутер в памяти.


- **Авторизация**: тест логинится админом на `/api/auth/login`, получает `accessToken`, и использует его в заголовке `Authorization: Bearer ...` для запросов к админским эндпоинтам.


- **Изоляция**: каждое выполнение тестов создаёт новые in-memoru storage, поэтому состояние чистое и независимое от запуска к запуску.


- `integration_test.go` — сценарий «сверху вниз»: вход админом → список пользователей → создание и чтение задач.

## Вывод

```
go run .
Server listening on :8080
```

```
go test -v ./...
=== RUN   TestTopDown_Flow_AdminAuthUsersTasks
--- PASS: TestTopDown_Flow_AdminAuthUsersTasks (0.16s)
PASS
ok      todoexample  0.165s
```