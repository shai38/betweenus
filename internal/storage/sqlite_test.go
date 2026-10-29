package storage

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/shai38/betweenus/internal/couples"
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

func createTestCouple(t *testing.T, rep *SQLiteRepository) int64 {
	t.Helper()

	coupleID, err := rep.CreateCouple(100, 200)
	if err != nil {
		t.Fatalf("CreateCouple() error: %v", err)
	}

	return coupleID
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
	coupleID := createTestCouple(t, rep)

	newPlan := plans.NewPlan{
		Title: "Вечер кино",
		Note:  "Выбрать фильм вместе",
		Topic: plans.Movies,
	}

	id, err := rep.Add(newPlan, coupleID)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("Add() returned ID %d, want positive ID", id)
	}
}

func TestAddPlanRejectsUnknownCoupleID(t *testing.T) {
	rep := newTestRepository(t)
	coupleID := createTestCouple(t, rep)

	newPlan := plans.NewPlan{Title: "Сходить в кино"}
	if _, err := rep.Add(newPlan, coupleID); err != nil {
		t.Fatalf("Add() with existing couple ID error: %v", err)
	}

	const unknownCoupleID int64 = 999999
	if _, err := rep.Add(newPlan, unknownCoupleID); err == nil {
		t.Fatal("Add() with unknown couple ID error = nil, want foreign key constraint error")
	}
}

func TestAllPlan(t *testing.T) {
	rep := newTestRepository(t)
	coupleID := createTestCouple(t, rep)

	movieNewPlan := plans.NewPlan{
		Title: "Вечер кино",
		Note:  "Выбрать фильм вместе",
		Topic: plans.Movies,
	}
	movieID, err := rep.Add(movieNewPlan, coupleID)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	restNewPlan := plans.NewPlan{
		Title: "Сходить в ресторан у воды",
		Note:  "Заказать блюдо дня",
		Topic: plans.Places,
	}

	restID, err := rep.Add(restNewPlan, coupleID)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	res, err := rep.All(coupleID)

	if err != nil {
		t.Fatalf("All() error: %v", err)
	}
	moviePlan := plans.Plan{
		ID:       movieID,
		NewPlan:  movieNewPlan,
		CoupleID: coupleID,
	}
	restPlan := plans.Plan{
		ID:       restID,
		NewPlan:  restNewPlan,
		CoupleID: coupleID,
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

func TestAllPlansAreScopedToCouple(t *testing.T) {
	rep := newTestRepository(t)
	firstCoupleID := createTestCouple(t, rep)
	secondCoupleID, err := rep.CreateCouple(300, 400)
	if err != nil {
		t.Fatalf("CreateCouple() error: %v", err)
	}

	firstPlan := plans.NewPlan{Title: "Вечер кино"}
	firstPlanID, err := rep.Add(firstPlan, firstCoupleID)
	if err != nil {
		t.Fatalf("Add() for first couple error: %v", err)
	}

	secondPlan := plans.NewPlan{Title: "Поход в музей"}
	secondPlanID, err := rep.Add(secondPlan, secondCoupleID)
	if err != nil {
		t.Fatalf("Add() for second couple error: %v", err)
	}

	firstCouplePlans, err := rep.All(firstCoupleID)
	if err != nil {
		t.Fatalf("All() for first couple error: %v", err)
	}
	if len(firstCouplePlans) != 1 {
		t.Fatalf("All() for first couple returned %d plans, want 1", len(firstCouplePlans))
	}
	if firstCouplePlans[0].ID != firstPlanID || firstCouplePlans[0].CoupleID != firstCoupleID {
		t.Errorf("All() for first couple returned unexpected plan: %+v", firstCouplePlans[0])
	}

	secondCouplePlans, err := rep.All(secondCoupleID)
	if err != nil {
		t.Fatalf("All() for second couple error: %v", err)
	}
	if len(secondCouplePlans) != 1 {
		t.Fatalf("All() for second couple returned %d plans, want 1", len(secondCouplePlans))
	}
	if secondCouplePlans[0].ID != secondPlanID || secondCouplePlans[0].CoupleID != secondCoupleID {
		t.Errorf("All() for second couple returned unexpected plan: %+v", secondCouplePlans[0])
	}
}

func TestPlanIsSharedByBothCoupleMembers(t *testing.T) {
	rep := newTestRepository(t)
	coupleID := createTestCouple(t, rep)

	planID, err := rep.Add(plans.NewPlan{Title: "Вечер кино"}, coupleID)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	firstUserCouple, err := rep.GetCoupleByUserID(100)
	if err != nil {
		t.Fatalf("GetCoupleByUserID() for first user error: %v", err)
	}
	secondUserCouple, err := rep.GetCoupleByUserID(200)
	if err != nil {
		t.Fatalf("GetCoupleByUserID() for second user error: %v", err)
	}
	if firstUserCouple.ID != secondUserCouple.ID {
		t.Fatalf("couple IDs = %d and %d, want the same couple", firstUserCouple.ID, secondUserCouple.ID)
	}

	for _, member := range []struct {
		name     string
		coupleID int64
	}{
		{name: "first user", coupleID: firstUserCouple.ID},
		{name: "second user", coupleID: secondUserCouple.ID},
	} {
		t.Run(member.name, func(t *testing.T) {
			got, err := rep.All(member.coupleID)
			if err != nil {
				t.Fatalf("All() error: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("All() returned %d plans, want 1", len(got))
			}
			if got[0].ID != planID {
				t.Errorf("plan ID = %d, want %d", got[0].ID, planID)
			}
		})
	}
}

func TestSetDone(t *testing.T) {
	rep := newTestRepository(t)
	coupleID := createTestCouple(t, rep)

	movieNewPlan := plans.NewPlan{
		Title: "Вечер кино",
		Note:  "Выбрать фильм вместе",
		Topic: plans.Movies,
	}
	movieID, err := rep.Add(movieNewPlan, coupleID)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	if err = rep.SetDone(movieID, coupleID, true); err != nil {
		t.Fatalf("SetDone() error: %v", err)
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

func TestSetDoneOnlyChangesPlansInSameCouple(t *testing.T) {
	rep := newTestRepository(t)
	firstCoupleID := createTestCouple(t, rep)
	secondCoupleID, err := rep.CreateCouple(300, 400)
	if err != nil {
		t.Fatalf("CreateCouple() error: %v", err)
	}

	firstPlanID, err := rep.Add(plans.NewPlan{Title: "Сходить в кино"}, firstCoupleID)
	if err != nil {
		t.Fatalf("Add() for first couple error: %v", err)
	}
	secondPlanID, err := rep.Add(plans.NewPlan{Title: "Устроить пикник"}, secondCoupleID)
	if err != nil {
		t.Fatalf("Add() for second couple error: %v", err)
	}

	for _, plan := range []struct {
		id       int64
		coupleID int64
	}{
		{id: firstPlanID, coupleID: firstCoupleID},
		{id: secondPlanID, coupleID: secondCoupleID},
	} {
		if err := rep.SetDone(plan.id, plan.coupleID, true); err != nil {
			t.Fatalf("SetDone(..., true) error: %v", err)
		}
		if err := rep.SetDone(plan.id, plan.coupleID, false); err != nil {
			t.Fatalf("SetDone(..., false) error: %v", err)
		}
	}

	if err := rep.SetDone(firstPlanID, secondCoupleID, true); !errors.Is(err, plans.ErrPlanNotFound) {
		t.Errorf("SetDone() for another couple's plan error = %v, want %v", err, plans.ErrPlanNotFound)
	}
	if err := rep.SetDone(secondPlanID, firstCoupleID, true); !errors.Is(err, plans.ErrPlanNotFound) {
		t.Errorf("SetDone() for another couple's plan error = %v, want %v", err, plans.ErrPlanNotFound)
	}

	for _, couple := range []struct {
		id     int64
		planID int64
	}{
		{id: firstCoupleID, planID: firstPlanID},
		{id: secondCoupleID, planID: secondPlanID},
	} {
		got, err := rep.All(couple.id)
		if err != nil {
			t.Fatalf("All() error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("All() returned %d plans, want 1", len(got))
		}
		if got[0].ID != couple.planID {
			t.Errorf("plan ID = %d, want %d", got[0].ID, couple.planID)
		}
		if got[0].Done {
			t.Errorf("plan %d done = true after rejected update, want false", couple.planID)
		}
	}
}

func TestDeletePlan(t *testing.T) {
	rep := newTestRepository(t)
	coupleID := createTestCouple(t, rep)

	movieNewPlan := plans.NewPlan{
		Title: "Вечер кино",
		Note:  "Выбрать фильм вместе",
		Topic: plans.Movies,
	}
	movieID, err := rep.Add(movieNewPlan, coupleID)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	err = rep.DeletePlan(movieID)
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	res, err := rep.All(coupleID)

	if err != nil {
		t.Fatalf("All() error: %v", err)
	}

	if len(res) != 0 {
		t.Fatalf("All() returned %d plans after deletion, want 0", len(res))
	}

	err = rep.DeletePlan(movieID)
	if !errors.Is(err, plans.ErrPlanNotFound) {
		t.Fatalf("expected not found error while deleting, got: %v", err)
	}
}

func TestCreateInvitation(t *testing.T) {
	repo := newTestRepository(t)

	invitation := couples.Invitation{
		Code:            "test-code-123",
		CreatedByUserID: 12345,
	}

	err := repo.CreateInvitation(invitation)
	if err != nil {
		t.Fatalf("CreateInvitation() error = %v", err)
	}

	got, err := repo.GetInvitationByCode(invitation.Code)
	if err != nil {
		t.Fatalf("GetInvitationByCode() error = %v", err)
	}

	if got.Code != invitation.Code {
		t.Errorf("Code = %q, want %q", got.Code, invitation.Code)
	}

	if got.CreatedByUserID != invitation.CreatedByUserID {
		t.Errorf(
			"CreatedByUserID = %d, want %d",
			got.CreatedByUserID,
			invitation.CreatedByUserID,
		)
	}
}
