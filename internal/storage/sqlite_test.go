package storage

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/shai38/betweenus/internal/plans"
)

const testDBName = "betweenus.db"

// newTestRepository вспомогательная функция создания репозитория с проверкой на ошибку и функцией закрытия после теста
func newTestRepository(t *testing.T) *SQLiteRepository {
	/*Говорит Go считать эту функцию вспомогательной: если тест упадёт внутри неё,
	сообщение об ошибке будет указывать на вызывающий тест, а не на строку внутри помощника.*/
	t.Helper()

	tempDbPath := filepath.Join(t.TempDir(), testDBName)

	rep, err := NewSQLiteRepository(tempDbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error: %v", err)
	}

	/*Регистрирует функцию, которая выполнится после завершения теста.
	Это удобно для закрытия репозитория; больше не нужно повторять defer в каждом тесте.*/
	t.Cleanup(func() {
		if err := rep.Close(); err != nil {
			t.Errorf("Close() error: %v", err)
		}
	})

	return rep
}

func TestNewSQLiteRepository(t *testing.T) {
	rep := newTestRepository(t)

	var tableName string

	// Проверка, что таблица plans действительно появилась в базе
	const getPlansTable = `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table' AND name = 'plans';
	`

	// QueryRow возвращает одну строку,
	//.Scan() в database/sql считывает значения из результата SQL-запроса и записывает их в Go-переменные
	err := rep.db.QueryRow(getPlansTable).Scan(&tableName)
	if err != nil {
		t.Fatalf("failed to query plans table: %v", err)
	}

	if tableName != "plans" {
		t.Fatalf("got table name %q, want %q", tableName, "plans")
	}
}

func TestAddPlan(t *testing.T) {
	rep := newTestRepository(t)

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

func TestAllPlan(t *testing.T) {
	rep := newTestRepository(t)

	movieNewPlan := plans.NewPlan{
		Title: "Вечер кино",
		Note:  "Выбрать фильм вместе",
		Topic: plans.Movies,
	}
	movieID, err := rep.Add(movieNewPlan)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	restNewPlan := plans.NewPlan{
		Title: "Сходить в ресторан у воды",
		Note:  "Заказать блюдо дня",
		Topic: plans.Places,
	}

	restID, err := rep.Add(restNewPlan)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	res, err := rep.All()

	if err != nil {
		t.Fatalf("All() error: %v", err)
	}
	moviePlan := plans.Plan{
		ID:      movieID,
		NewPlan: movieNewPlan,
	}
	restPlan := plans.Plan{
		ID:      restID,
		NewPlan: restNewPlan,
	}

	if len(res) != 2 {
		t.Fatalf("All() returned %d plans, want 2", len(res))
	}

	for _, plan := range res {
		if plan != moviePlan && plan != restPlan {
			t.Errorf("unexpected plan: got %v", plan)
		}
	}

}

func TestMarkDone(t *testing.T) {
	rep := newTestRepository(t)

	movieNewPlan := plans.NewPlan{
		Title: "Вечер кино",
		Note:  "Выбрать фильм вместе",
		Topic: plans.Movies,
	}
	movieID, err := rep.Add(movieNewPlan)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	if err = rep.MarkDone(movieID); err != nil {
		t.Fatalf("MarkDone() error: %v", err)
	}

	var done bool

	err = rep.db.QueryRow(`SELECT done FROM plans WHERE id = ?`, movieID).Scan(&done)
	if err != nil {
		t.Fatalf("failed to query plans table: %v", err)
	}

	if !done {
		t.Fatalf("got done %t, want true", done)
	}

}

func TestDeletePlan(t *testing.T) {
	rep := newTestRepository(t)

	movieNewPlan := plans.NewPlan{
		Title: "Вечер кино",
		Note:  "Выбрать фильм вместе",
		Topic: plans.Movies,
	}
	movieID, err := rep.Add(movieNewPlan)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	err = rep.Delete(movieID)
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	res, err := rep.All()

	if err != nil {
		t.Fatalf("All() error: %v", err)
	}

	if len(res) != 0 {
		t.Fatalf("All() returned %d plans after deletion, want 0", len(res))
	}

	err = rep.Delete(movieID)
	if !errors.Is(err, plans.ErrPlanNotFound) {
		t.Fatalf("expected not found error while deleting, got: %v", err)
	}
}
