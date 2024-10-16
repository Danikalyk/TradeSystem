package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool

func InitDB(dsn string) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Не удалось разобрать строку подключения: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Не удалось создать пул подключений: %v", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	DBPool = pool
	fmt.Println("Успешно подключено к базе данных")
}

func CloseDB() {
	DBPool.Close()
}
