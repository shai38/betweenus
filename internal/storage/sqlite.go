package storage

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Проверяет доступность базы
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteRepository{db: db}, nil
}
