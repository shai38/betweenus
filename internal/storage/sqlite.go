package storage

import (
	"database/sql"

	"github.com/shai38/betweenus/internal/plans"
	_ "modernc.org/sqlite"
)

const createPlansTable = `
	CREATE TABLE IF NOT EXISTS plans (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		note TEXT NOT NULL DEFAULT '',
		topic INTEGER NOT NULL,
		done INTEGER NOT NULL DEFAULT 0 CHECK (done IN (0, 1))
	);
`

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

	// Создает таблицу, если ее нет
	if _, err := db.Exec(createPlansTable); err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteRepository{db: db}, nil
}

// Закрывает соединение
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// Добавляет новый план в базу данных и возвращает ID созданного плана и ошибку, если она есть
func (r *SQLiteRepository) Add(plan plans.NewPlan) (plans.ID, error) {
	result, err := r.db.Exec(
		`INSERT INTO plans (title, note, topic)
		 VALUES (?, ?, ?)`,
		plan.Title, plan.Note, plan.Topic,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return plans.ID(id), nil
}
