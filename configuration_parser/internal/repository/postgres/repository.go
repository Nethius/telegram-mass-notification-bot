package postgres

import (
	"configuration_parser/internal/repository"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"
)

const uniqueViolation = "23505"

type Repository struct {
	db *sql.DB
}

func NewRepository() (*Repository, error) {
	connStr, err := getPostgresCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get db credentials from env: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %v", err)
	}

	for i := 0; i < 3; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("db connection is not active: %v", err)
	}

	return &Repository{db: db}, nil
}

func (repo *Repository) Close() error {
	err := repo.db.Close()
	if err != nil {
		return err
	}
	return nil
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

	var userId int64
	row := repo.db.QueryRow(q, userName)
	if err := row.Err(); err != nil {
		return userId, err
	}

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
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return repository.ErrNotExists
	}
	return nil
}

func getPostgresCredentials() (string, error) {
	host, ok := os.LookupEnv("PGHOST")
	if !ok {
		return "", errors.New("failed to get PGHOST from env")
	}

	port, ok := os.LookupEnv("PGPORT")
	if !ok {
		return "", errors.New("failed to get PGPORT from env")
	}

	user, ok := os.LookupEnv("PGUSER")
	if !ok {
		return "", errors.New("failed to get PGUSER from env")
	}

	password, ok := os.LookupEnv("PGPASSWORD")
	if !ok {
		return "", errors.New("failed to get PGPASSWORD from env")
	}

	dbname, ok := os.LookupEnv("PGDATABASE")
	if !ok {
		return "", errors.New("failed to get PGDATABASE from env")
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname), nil
}
