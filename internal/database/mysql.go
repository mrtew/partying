package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func NewMySQL(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to mysql: %w", err)
	}
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}
	fmt.Println("✅ MySQL connected")
	return db, nil
}

// initSchema 自动建表，如果表已经存在则跳过
func initSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id          VARCHAR(36)  PRIMARY KEY,
			username    VARCHAR(50)  UNIQUE NOT NULL,
			password    VARCHAR(255) NOT NULL,
			created_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS rooms (
			id           VARCHAR(36)  PRIMARY KEY,
			name         VARCHAR(100) NOT NULL,
			host_id      VARCHAR(36)  NOT NULL,
			max_capacity INT          DEFAULT 10,
			created_at   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
