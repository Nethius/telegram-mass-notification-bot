package postgres

import (
	"configuration_parser/internal/repository"
	"database/sql"
	"github.com/lib/pq"
)

const uniqueViolation = "23505"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) InsertUser(userId int64, userName string) error {
	q := `INSERT INTO users (id, userName, is_active) VALUES ($1, $2, true)`

	if _, err := repo.db.Exec(q, userId, userName); err != nil {
		if e, ok := err.(*pq.Error); ok {
			if e.Code == uniqueViolation {
				return repository.ErrAlreadyExists
			}
		}
		return err
	}

	return nil
}

func (repo *Repository) GetUser(userName string) (int64, error) {
	q := `SELECT id FROM users WHERE username = $1`

	var userId int64 = 0
	row := repo.db.QueryRow(q, userName)
	if err := row.Scan(&userId); err != nil {
		return userId, err
	}

	return userId, nil
}

func (repo *Repository) AddNotificationAccess(userId int64, userNameWithAccess string) error {
	q := `INSERT INTO notification_access (user_id, username_with_access) VALUES ($1, $2)`

	if _, err := repo.db.Exec(q, userId, userNameWithAccess); err != nil {
		if e, ok := err.(*pq.Error); ok {
			if e.Code == uniqueViolation {
				return repository.ErrAlreadyExists
			}
		}
		return err
	}

	return nil
}

func (repo *Repository) RemoveNotificationAccess(userId int64, userNameWithAccess string) error {
	q := `DELETE FROM notification_access where user_id = $1 and username_with_access = $2`

	res, err := repo.db.Exec(q, userId, userNameWithAccess)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			if count == 0 {
				return repository.ErrNotExists
			}
			return nil
		}
		return err
	}
	return err
}
