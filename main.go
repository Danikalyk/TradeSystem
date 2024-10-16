package main

import "TradeSystem/db"

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/trade_system?sslmode=disable"

	db.InitDB(dsn)
	defer db.CloseDB()
}
