package postgreSQL

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"sso/internal/config"
	"sso/internal/domain/models"
	"sso/internal/storage"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func New(cfg *config.Config) (*Storage, error) {
	const op = "Storage.PostgreSQL.New"

	db, err := sql.Open("pgx", cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(email string, passHash []byte) (int64, error) {
	const op = "Storage.PostgreSQL.SaveUser"
	var id int64
	err := s.db.QueryRow("INSERT INTO users(email, pass_hash, timestamp) VALUES ($1, $2, $3) RETURNING id", email, passHash, time.Now()).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, fmt.Errorf("%s:%w", op, storage.ErrUserAlreadyExists)
	}
	if err != nil {
		return 0, fmt.Errorf("%s:%w", op, err)
	}
	return id, nil
}

func (s *Storage) GetUserByEmail(email string) (*models.User, error) {
	const op = "Storage.PostgreSQL.GetUserByEmail"
	row := s.db.QueryRow("SELECT id, email, pass_hash FROM users WHERE email = $1", email)
	user := &models.User{}

	err := row.Scan(&user.Id, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s:%w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s:%w", op, err)
	}
	return &models.User{
		Id:       user.Id,
		Email:    user.Email,
		PassHash: user.PassHash,
	}, nil
}

func (s *Storage) GetAppById(appId int) (*models.App, error) {
	const op = "Storage.PostgreSQL.GetAppById"
	row := s.db.QueryRow("SELECT id, name, secret FROM apps WHERE id = $1", appId)
	app := &models.App{}

	err := row.Scan(&app.Id, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s:%w", op, storage.ErrAppNotFound)
		}
		return nil, fmt.Errorf("%s:%w", op, err)
	}
	return &models.App{
		Id:     app.Id,
		Name:   app.Name,
		Secret: app.Secret,
	}, nil
}

func (s *Storage) IsAdmin(userId int64) (bool, error) {
	const op = "Storage.PostgreSQL.IsAdmin"
	row := s.db.QueryRow("SELECT is_admin FROM users WHERE id = $1", userId)
	var isAdmin bool

	err := row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s:%w", op, storage.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s:%w", op, err)
	}
	return isAdmin, nil
}

func (s *Storage) SetAdmin(userId int64, isAdmin bool) (bool, error) {
	const op = "Storage.PostgreSQL.SetAdmin"
	_, err := s.db.Exec("UPDATE users SET is_admin = $1 WHERE id = $2", isAdmin, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s:%w", op, storage.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s:%w", op, err)
	}
	return true, nil
}
