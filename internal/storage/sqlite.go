package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/shai38/betweenus/internal/couples"
	"github.com/shai38/betweenus/internal/plans"
	_ "modernc.org/sqlite"
)

const createPlansTable = `
	CREATE TABLE IF NOT EXISTS plans (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		note TEXT NOT NULL DEFAULT '',
		topic INTEGER NOT NULL,
		done INTEGER NOT NULL DEFAULT 0 CHECK (done IN (0, 1)),
		couple_id INTEGER NOT NULL REFERENCES couples(id)
	);
`

const createInvitationsTable = `
	CREATE TABLE IF NOT EXISTS invitations (
        code TEXT PRIMARY KEY,
        created_by_user_id INTEGER NOT NULL
	);
`

const createCouplesTable = `
	CREATE TABLE IF NOT EXISTS couples (
		id INTEGER PRIMARY KEY,
		first_user_id INTEGER NOT NULL,
		second_user_id INTEGER NOT NULL,
		CHECK (first_user_id <> second_user_id)
	);
`

const createUniqueCoupleIndex = `
    CREATE UNIQUE INDEX IF NOT EXISTS idx_couples_unique_pair
    ON couples (
        MIN(first_user_id, second_user_id),
        MAX(first_user_id, second_user_id)
    );
`

type SQLiteRepository struct {
	db *sql.DB
}

type rowQuerier interface {
	QueryRow(query string, args ...any) *sql.Row
}

type execQuerier interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}

	// Проверяет доступность базы
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	queries := []string{
		createPlansTable,
		createInvitationsTable,
		createCouplesTable,
		createUniqueCoupleIndex,
	}

	err = createTables(db, queries)

	if err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteRepository{db: db}, nil
}

// createTables cоздает таблицы, если их нет
func createTables(db *sql.DB, queries []string) error {
	for _, sqlQuery := range queries {
		if _, err := db.Exec(sqlQuery); err != nil {
			return err
		}
	}

	return nil
}

// Закрывает соединение
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// Добавляет новый план в базу данных и возвращает ID созданного плана и ошибку, если она есть
func (r *SQLiteRepository) Add(plan plans.NewPlan, coupleID int64) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO plans (title, note, topic, couple_id)
		 VALUES (?, ?, ?, ?)`,
		plan.Title,
		plan.Note,
		plan.Topic,
		coupleID,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Возвращает все записи планов и ошибку
func (r *SQLiteRepository) All(coupleID int64) ([]plans.Plan, error) {
	rows, err := r.db.Query(
		`SELECT id, title, note, topic, done
		FROM plans
		WHERE couple_id = ?`,
		coupleID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []plans.Plan
	var id int64
	var title, note string
	var topic plans.Topic
	var done bool

	for rows.Next() {
		err := rows.Scan(&id, &title, &note, &topic, &done)
		if err != nil {
			return nil, err
		}
		plan := plans.Plan{
			ID: id,
			NewPlan: plans.NewPlan{
				Title: title,
				Note:  note,
				Topic: topic,
			},
			Done:     done,
			CoupleID: coupleID,
		}
		result = append(result, plan)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// SetDone меняет статус плана - отмечает выполненным, также может вернуть его в активные.
func (r *SQLiteRepository) SetDone(planID int64, coupleID int64, done bool) error {
	result, err := r.db.Exec(
		`UPDATE plans 
		SET done = ?
		WHERE id = ?
		AND couple_id = ?`,
		done,
		planID,
		coupleID,
	)
	if err != nil {
		return err
	}

	rowsCount, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsCount == 0 {
		return fmt.Errorf("%w: ID = %v, coupleID = %v", plans.ErrPlanNotFound, planID, coupleID)
	}

	return nil
}

// Удаляет план
func (r *SQLiteRepository) DeletePlan(planID int64) error {
	result, err := r.db.Exec(`DELETE FROM plans WHERE id = ?`, planID)

	if err != nil {
		return err
	}

	rowsCount, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsCount == 0 {
		return fmt.Errorf("%w: ID = %v", plans.ErrPlanNotFound, planID)
	}

	return nil
}

func (r *SQLiteRepository) CreateInvitation(inv couples.Invitation) error {
	return createInvitation(r.db, inv)
}

func (r *SQLiteRepository) GetInvitationByCode(code string) (couples.Invitation, error) {
	return getInvitationByCode(r.db, code)
}

func (r *SQLiteRepository) GetInvitationByCreatorUserID(userID int64) (couples.Invitation, error) {
	return getInvitationByCreatorUserID(r.db, userID)
}

func (r *SQLiteRepository) DeleteInvitationsByCreatorID(userID int64) error {
	return deleteInvitationsByCreatorID(r.db, userID)
}

func (r *SQLiteRepository) CreateCouple(firstUserID int64, secondUserID int64) (int64, error) {
	return createCouple(r.db, firstUserID, secondUserID)
}

func (r *SQLiteRepository) GetCoupleByID(coupleID int64) (couples.Couple, error) {
	couple := couples.Couple{}

	err := r.db.QueryRow(
		`SELECT id, first_user_id, second_user_id FROM couples WHERE id = ?`,
		coupleID,
	).Scan(
		&couple.ID,
		&couple.FirstUserID,
		&couple.SecondUserID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = couples.ErrCoupleNotFound
		}
		return couples.Couple{}, err
	}

	return couple, nil
}

func (r *SQLiteRepository) GetCoupleByUserID(userID int64) (couples.Couple, error) {
	return getCoupleByUserID(r.db, userID)
}

func (r *SQLiteRepository) AcceptInvitation(
	code string,
	receiverID int64,
) (int64, error) {

	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	inv, err := getInvitationByCode(tx, code)
	if err != nil {
		return 0, err
	}

	_, err = getCoupleByUserID(tx, inv.CreatedByUserID)
	if err == nil {
		return 0, couples.ErrOnlyOneCouplePerUser
	}
	if !errors.Is(err, couples.ErrCoupleNotFound) {
		return 0, err
	}

	_, err = getCoupleByUserID(tx, receiverID)
	if err == nil {
		return 0, couples.ErrOnlyOneCouplePerUser
	}
	if !errors.Is(err, couples.ErrCoupleNotFound) {
		return 0, err
	}

	coupleID, err := createCouple(tx, inv.CreatedByUserID, receiverID)
	if err != nil {
		return 0, err
	}

	if err := deleteInvitationsByCreatorID(tx, inv.CreatedByUserID); err != nil {
		return 0, err
	}
	err = deleteInvitationsByCreatorID(tx, receiverID)
	if err != nil && !errors.Is(err, couples.ErrInvitationNotFound) {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return coupleID, nil
}

func (r *SQLiteRepository) CreateInvitationIfEligible(
	invitation couples.Invitation,
) (couples.Invitation, error) {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return couples.Invitation{}, err
	}
	defer tx.Rollback()

	// Проверяем, состоит ли пользователь в паре
	_, err = getCoupleByUserID(tx, invitation.CreatedByUserID)
	if err == nil {
		return couples.Invitation{}, couples.ErrOnlyOneCouplePerUser
	}
	if !errors.Is(err, couples.ErrCoupleNotFound) {
		return couples.Invitation{}, err
	}

	// Проверяем, есть ли уже активное приглашение
	existing, err := getInvitationByCreatorUserID(
		tx,
		invitation.CreatedByUserID,
	)
	if err == nil {
		// Изменений не было, поэтому можно просто завершить транзакцию
		if err := tx.Commit(); err != nil {
			return couples.Invitation{}, err
		}
		return existing, nil
	}
	if !errors.Is(err, couples.ErrInvitationNotFound) {
		return couples.Invitation{}, err
	}

	// Создаём приглашение в той же транзакции
	if err := createInvitation(tx, invitation); err != nil {
		return couples.Invitation{}, err
	}

	if err := tx.Commit(); err != nil {
		return couples.Invitation{}, err
	}

	return invitation, nil
}

func createInvitation(db execQuerier, inv couples.Invitation) error {
	_, err := db.Exec(
		`INSERT INTO invitations (code, created_by_user_id) 
		VALUES(?, ?)`,
		inv.Code,
		inv.CreatedByUserID,
	)

	return err
}

func getCoupleByUserID(db rowQuerier, userID int64) (couples.Couple, error) {
	couple := couples.Couple{}

	err := db.QueryRow(
		`SELECT id, first_user_id, second_user_id 
		FROM couples 
		WHERE first_user_id = ? OR second_user_id = ?`,
		userID,
		userID,
	).Scan(
		&couple.ID,
		&couple.FirstUserID,
		&couple.SecondUserID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = couples.ErrCoupleNotFound
		}
		return couples.Couple{}, err
	}

	return couple, nil
}

func createCouple(db execQuerier, firstUserID int64, secondUserID int64) (int64, error) {
	if firstUserID == secondUserID {
		return 0, couples.ErrSameUser
	}

	result, err := db.Exec(
		`INSERT INTO couples (first_user_id, second_user_id)
		VALUES (?, ?)`,
		firstUserID,
		secondUserID,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func getInvitationByCode(
	db rowQuerier,
	code string,
) (couples.Invitation, error) {
	inv := couples.Invitation{}

	err := db.QueryRow(
		`SELECT code, created_by_user_id FROM invitations WHERE code = ?`,
		code,
	).Scan(
		&inv.Code,
		&inv.CreatedByUserID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = couples.ErrInvitationNotFound
		}
		return couples.Invitation{}, err
	}

	return inv, nil
}

func getInvitationByCreatorUserID(
	db rowQuerier,
	userID int64,
) (couples.Invitation, error) {
	inv := couples.Invitation{}

	err := db.QueryRow(
		`SELECT code, created_by_user_id FROM invitations WHERE created_by_user_id = ?`,
		userID,
	).Scan(
		&inv.Code,
		&inv.CreatedByUserID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = couples.ErrInvitationNotFound
		}
		return couples.Invitation{}, err
	}

	return inv, nil
}

func deleteInvitationsByCreatorID(db execQuerier, userID int64) error {
	result, err := db.Exec(`DELETE FROM invitations WHERE created_by_user_id = ?`, userID)

	if err != nil {
		return err
	}

	rowsCount, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsCount == 0 {
		return fmt.Errorf("%w: created_by_user_id = %v", couples.ErrInvitationNotFound, userID)
	}

	return nil
}
