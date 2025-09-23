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

## Тестирование «снизу‑вверх» (bottom‑up)

Модульные тесты проверяют отдельные компоненты в изоляции, без HTTP‑слоя:

- `store_test.go` — CRUD и валидация для задач (`Store`, `UpdateTaskPayload`).

- `users_store_test.go` — уникальность `username/email`, обновление и удаление (`UsersStore`).

- `refresh_store_test.go` — добавление, истечение и отзыв (`RefreshStore`).

- `tokens_test.go` — выпуск access‑токена и валидация `Authorization: Bearer ...`.

Запуск:

```
go test -v ./...
=== RUN   TestTopDown_Flow_AdminAuthUsersTasks
--- PASS: TestTopDown_Flow_AdminAuthUsersTasks (0.14s)
=== RUN   TestRefreshStore_Add_IsValid_Remove
--- PASS: TestRefreshStore_Add_IsValid_Remove (0.12s)
=== RUN   TestStore_Create_Get_Update_Delete
--- PASS: TestStore_Create_Get_Update_Delete (0.00s)
=== RUN   TestTokens_IssueAndValidateAccess
--- PASS: TestTokens_IssueAndValidateAccess (0.00s)
=== RUN   TestUsersStore_CRUD_Uniqueness
--- PASS: TestUsersStore_CRUD_Uniqueness (0.00s)
PASS
ok      todoexample  0.263s
```