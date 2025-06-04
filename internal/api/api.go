package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/vitalikir156/tasker3/internal/api/middleware"
	"github.com/vitalikir156/tasker3/internal/service"
)

// Routers - структура для хранения зависимостей роутов
type Routers struct {
	Service service.Service
}

// NewRouters - конструктор для настройки API
func NewRouters(r *Routers, token string) *fiber.App {
	app := fiber.New()

	// Настройка CORS (разрешенные методы, заголовки, авторизация)
	app.Use(cors.New(cors.Config{
		AllowMethods:     "GET, POST, PUT, DELETE",
		AllowHeaders:     "Accept, Authorization, Content-Type, X-CSRF-Token, X-REQUEST-SomeID",
		ExposeHeaders:    "Link",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Группа маршрутов с авторизацией
	apiGroup := app.Group("/v1", middleware.Authorization(token))

	// Роут для создания задачиinternal/config/config.gointernal/config/config.go
	apiGroup.Post("/tasks", r.Service.CreateTask)
	apiGroup.Get("/tasks", r.Service.GetTasks)
	apiGroup.Get("/tasks/:id", r.Service.GetTask)
	apiGroup.Put("/tasks/:id", r.Service.UpdateTask)
	apiGroup.Delete("/tasks/:id", r.Service.DeleteTask)

	return app
}
