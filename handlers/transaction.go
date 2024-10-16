package handlers

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryModification struct {
	ItemID   int `json:"item_id"`
	Quantity int `json:"quantity"`
}

type TransactionRequest struct {
	FromPlayerID  int                     `json:"from_player_id"`
	ToPlayerID    int                     `json:"to_player_id"`
	Amount        float64                 `json:"money_amount"`
	InventoryMods []InventoryModification `json:"inventory_mods"`
}

func TransactionHandler(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req TransactionRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Неверный формат запроса",
			})
		}

		ctx := context.Background()
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Не удалось начать транзакцию",
			})
		}
		defer func() {
			if err != nil {
				tx.Rollback(ctx)
			}
		}()

		var fromBalance float64
		err = tx.QueryRow(ctx, "SELECT player_money_amount from players WHERE player_id=$1 FOR UPDATE",
			req.FromPlayerID).Scan(&fromBalance)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Исходный игрок не найден",
			})
		}

		if fromBalance < req.Amount {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Недостаточно средств на балансе игрока",
			})
		}

		_, err = tx.Exec(ctx, "UPDATE players SET player_money_amount = player_money_amount - $1 WHERE player_id = $2", req.Amount, req.FromPlayerID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Не удалось обновить счёт игрока",
			})
		}

		for _, mod := range req.InventoryMods {
			var currentQty int
			err = tx.QueryRow(ctx, "SELECT quantity FROM inventory WHERE player_id=$1 AND item_id=$2 FOR UPDATE", req.FromPlayerID, mod.ItemID).Scan(&currentQty)
			if err != nil {
				if err.Error() == "no rows in result set" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error": fmt.Sprintf("Инвентарь не найден для предмета ID %d", mod.ItemID),
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Не удалось получить инвентарь",
				})
			}

			newQty := currentQty + mod.Quantity
			if newQty < 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("Недостаточно предметов %d", mod.ItemID),
				})
			}

			_, err = tx.Exec(ctx, "UPDATE inventory SET quantity = $1 WHERE player_id = $2 AND item_id = $3", newQty, req.FromPlayerID, mod.ItemID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Не удалось обновить инвентарь",
				})
			}
		}

		var toBalance float64
		err = tx.QueryRow(ctx, "SELECT player_money_amount FROM players WHERE player_id=$1 FOR UPDATE", req.ToPlayerID).Scan(&toBalance)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Целевой счёт не найден",
			})
		}

		_, err = tx.Exec(ctx, "UPDATE players SET player_money_amount = player_money_amount + $1 WHERE player_id = $2", req.Amount, req.ToPlayerID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Не удалось обновить целевой счёт",
			})
		}

		err = tx.Commit(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Не удалось завершить транзакцию",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Транзакция успешно выполнена",
		})
	}
}
