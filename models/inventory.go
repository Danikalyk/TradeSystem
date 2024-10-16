package models

type Inventory struct {
	AccountID int `json:"player_id"`
	ItemID    int `json:"item_id"`
	Quantity  int `json:"quantity"`
}
