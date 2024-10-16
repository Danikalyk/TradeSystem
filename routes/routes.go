package routes

import (
	"TradeSystem/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRoutes(app *fiber.App, db *pgxpool.Pool) {
	api := app.Group("/api")

	api.Post("/transaction", handlers.TransactionHandler(db))
}
