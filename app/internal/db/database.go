package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewConnection(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database connection string is required")
	}

	const (
		maxRetries    = 10
		retryInterval = 3 * time.Second
	)

	var (
		db  *sql.DB
		err error
	)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			return nil, fmt.Errorf("open database failed: %w", err)
		}

		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(30 * time.Minute)

		if err = db.Ping(); err == nil {
			log.Println("database connection established")
			return db, nil
		}

		log.Printf("database connection failed (%d/%d): %v, retrying in %s", attempt, maxRetries, err, retryInterval)
		db.Close()
		time.Sleep(retryInterval)
	}

	return nil, fmt.Errorf("database connection failed after retries: %w", err)
}
