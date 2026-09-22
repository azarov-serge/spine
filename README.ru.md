# Spine

Тонкий HTTP-фреймворк на Go — обёртка над [chi](https://github.com/go-chi/chi) в стиле **Express / Fastify / Fiber**.

Handler возвращает `error`. Ответы пишутся через методы `Context`, а не напрямую в `http.ResponseWriter`.

[English](README.md) · [pkg.go.dev](https://pkg.go.dev/github.com/azarov-serge/spine)

## Установка

```bash
go get github.com/azarov-serge/spine@latest
```

## Быстрый старт

```go
package main

import (
	"log"

	"github.com/azarov-serge/spine"
)

func main() {
	app := spine.New(spine.Config{ServiceName: "demo"})

	app.Use(
		spine.RequestID(),
		spine.Logger(),
		spine.Recover(),
	)

	app.GET("/ping", func(c *spine.Context) error {
		return c.Text(200, "pong")
	})

	app.GET("/todos/{id}", func(c *spine.Context) error {
		id := c.Param("id")
		if id == "" {
			return spine.BadRequest("id is required")
		}
		return c.OK(map[string]string{"id": id})
	})

	log.Fatal(app.Run(":8080"))
}
```

## Соответствие

| Express / Fastify / Fiber | Spine |
| ------------------------- | ----- |
| `(req, res) => { ... }` | `func(*Context) error` |
| `res.json(200, data)` | `return c.OK(data)` |
| `next()` в middleware | `return next(c)` |
| `app.use(mw)` | `router.Use(mw)` |
| `app.use('/api', r)` | `router.Group("/api", fn)` |

**Правило:** не пишите в `c.Writer` напрямую — только методы `Context` или возврат `*HTTPError`.  
Непредвиденная ошибка (не `*HTTPError`) → **500** с `internal_error`.

## Обзор API

### App

```go
app := spine.New(spine.Config{
    ServiceName: "demo", // имя в логе старта
    Logger:      nil,    // nil → text slog на stdout
})

app.Use(middlewares...)
app.GET / POST / PUT / PATCH / DELETE(path, handler)
app.Group("/api", func(r spine.Router) { ... })
app.Handler() // http.Handler
app.Run(":8080")
```

### Context

| Метод | Назначение |
| ----- | ---------- |
| `Context()` | `context.Context` для service / БД |
| `Param("id")` | path-параметр (`/todos/{id}`) |
| `Query("q")` | query string |
| `BindJSON(&dst)` | JSON body; лишние поля → 400 |
| `OK(data)` | 200 JSON |
| `Created(data)` | 201 JSON |
| `JSON(status, data)` | любой статус + JSON |
| `Text(status, s)` | plain text |
| `NoContent(status)` | без тела (обычно 204) |

### Ошибки

| Хелпер | Статус | `code` |
| ------ | ------ | ------ |
| `BadRequest` | 400 | `bad_request` |
| `Unauthorized` | 401 | `unauthorized` |
| `Forbidden` | 403 | `forbidden` |
| `NotFound` | 404 | `not_found` |
| `ValidationError` | 422 | `validation_error` |
| `Internal` | 500 | `internal_error` |

```json
{ "code": "not_found", "message": "todo not found" }
```

### Middleware

Встроенные: `RequestID()`, `Logger()`, `Recover()`.

```go
app.Use(spine.RequestID(), spine.Logger(), spine.Recover())

app.Group("/admin", func(admin spine.Router) {
    admin.Use(RequireAPIKey("secret"))
    admin.GET("/stats", stats)
})
```

Middleware родителя копируется во вложенные группы; `Use` группы добавляет поверх.

### Группы

```go
app.Group("/categories", func(cats spine.Router) {
    cats.GET("/", list)
    cats.Group("/{categoryId}/todos", func(todos spine.Router) {
        todos.GET("/{id}", get)
    })
})
```

Path-параметры — синтаксис chi: `{id}`, `{categoryId}`.

## Пример

См. [`examples/basic`](examples/basic).

## Требования

- Go 1.23+

## Лицензия

[MIT](LICENSE)
