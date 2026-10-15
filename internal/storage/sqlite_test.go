package storage

import (
	"path/filepath"
	"testing"
)

func TestNewSQLiteRepository(t *testing.T) {
	tempDbPath := filepath.Join(t.TempDir(), "betweenus.db")
	rep, err := NewSQLiteRepository(tempDbPath)
	if err != nil {
		// Регистрирует ошибку и немедленно завершает текущую тестовую функцию,
		// не успевая зарегистрировать defer
		t.Fatalf("NewSQLiteRepository(%v) returned error: %v", tempDbPath, err)
	}

	defer rep.Close()

	var tableName string

	// Проверка, что таблица plans действительно появилась в базе
	const getPlansTable = `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table' AND name = 'plans';
	`

	// QueryRow возвращает одну строку,
	//.Scan() в database/sql считывает значения из результата SQL-запроса и записывает их в Go-переменные
	err = rep.db.QueryRow(getPlansTable).Scan(&tableName)
	if err != nil {
		t.Fatalf("failed to query plans table: %v", err)
	}

	if tableName != "plans" {
		t.Fatalf("got table name %q, want %q", tableName, "plans")
	}
}
