package storage

import (
	"path/filepath"
	"testing"

	"github.com/shai38/betweenus/internal/plans"
)

const testDBName = "betweenus.db"

func TestNewSQLiteRepository(t *testing.T) {
	tempDbPath := filepath.Join(t.TempDir(), testDBName)
	rep, err := NewSQLiteRepository(tempDbPath)
	if err != nil {
		// Регистрирует ошибку и немедленно завершает текущую тестовую функцию,
		// не успевая зарегистрировать defer
		t.Fatalf("NewSQLiteRepository(%v) returned error: %v", tempDbPath, err)
	}

	defer func() {
		if err := rep.Close(); err != nil {
			t.Errorf("Close() error: %v", err)
		}
	}()

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

func TestAddPlan(t *testing.T) {
	tempDbPath := filepath.Join(t.TempDir(), testDBName)

	rep, err := NewSQLiteRepository(tempDbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error: %v", err)
	}

	defer func() {
		if err := rep.Close(); err != nil {
			t.Errorf("Close() error: %v", err)
		}
	}()

	newPlan := plans.NewPlan{
		Title: "Вечер кино",
		Note:  "Выбрать фильм вместе",
		Topic: plans.Movies,
	}

	id, err := rep.Add(newPlan)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("Add() returned ID %d, want positive ID", id)
	}
}
