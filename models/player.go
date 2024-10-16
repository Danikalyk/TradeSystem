package models

type Player struct {
	ID      uint    `json:"player_id"`
	Name    string  `json:"player_name"`
	Balance float64 `json:"player_money_amount"`
}
