package postgres

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
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
