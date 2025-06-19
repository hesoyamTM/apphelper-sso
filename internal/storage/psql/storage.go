package psql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-sso/internal/models"
	"github.com/hesoyamTM/apphelper-sso/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

type PsqlConfig struct {
	Host     string `yaml:"host" env-required:"true" env:"PSQL_HOST"`
	Port     int    `yaml:"port" env-required:"true" env:"PSQL_PORT"`
	User     string `yaml:"user" env-required:"true" env:"PSQL_USER"`
	Password string `yaml:"password" env-required:"true" env:"PSQL_PASSWORD"`
	DB       string `yaml:"db" env-required:"true" env:"PSQL_DATABASE"`
}

func New(ctx context.Context, cfg PsqlConfig) (*Storage, error) {
	const op = "psql.New"

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		pool: pool,
	}, nil
}

func (s *Storage) CrateUser(ctx context.Context, name, surname, login string, passHash []byte) (uuid.UUID, error) {
	const op = "psql.CreateUser"

	query := `INSERT INTO users (name, surname, login, pass_hash) VALUES ($1, $2, $3, $4) RETURNING id`

	row := s.pool.QueryRow(ctx, query, name, surname, login, passHash)

	var id uuid.NullUUID
	if err := row.Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return uuid.Nil, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
			}
		}

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id.UUID, nil
}

func (s *Storage) ProvideUserById(ctx context.Context, id uuid.UUID) (models.User, error) {
	const op = "psql.ProvideUserById"

	query := `SELECT name, surname, login, pass_hash FROM users WHERE id = $1`

	row := s.pool.QueryRow(ctx, query, id)

	var user models.User
	user.UserInfo.Id = id
	user.UserAuth.Id = id
	if err := row.Scan(&user.Name, &user.Surname, &user.Login, &user.PassHash); err != nil {
		if err == pgx.ErrNoRows {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) ProvideUserByLogin(ctx context.Context, login string) (models.User, error) {
	const op = "psql.ProvideUserByLogin"

	query := `SELECT id, name, surname, pass_hash FROM users WHERE login = $1`

	row := s.pool.QueryRow(ctx, query, login)

	var user models.User
	var id uuid.NullUUID
	err := row.Scan(&id, &user.Name, &user.Surname, &user.PassHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user.UserInfo.Id = id.UUID
	user.UserAuth.Id = id.UUID

	return user, nil
}

func (s *Storage) ProvideUsersById(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	const op = "psql.ProvideUsersById"

	if len(ids) == 0 {
		return nil, nil
	}

	args := make([]interface{}, 0, len(ids))
	inParams := make([]string, 0, len(ids))

	for i, id := range ids {
		args = append(args, interface{}(id))
		inParams = append(inParams, fmt.Sprintf("$%d", i+1))
	}

	query := fmt.Sprintf(`SELECT id, name, surname, login, pass_hash FROM users WHERE id in (%s)`, strings.Join(inParams, ","))

	users := make([]models.User, 0)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for rows.Next() {
		var user models.User
		var id uuid.NullUUID

		err := rows.Scan(&id, &user.Name, &user.Surname, &user.Login, &user.PassHash)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
			}

			return nil, fmt.Errorf("%s: %w", op, err)
		}

		user.UserInfo.Id = id.UUID
		user.UserAuth.Id = id.UUID

		users = append(users, user)
	}

	return users, nil
}

func (s *Storage) UpdateUser(ctx context.Context, user models.User) error {
	const op = "psql.UpdateUser"

	query := `UPDATE users SET name = $1, surname = $2, login = $3, pass_hash = $4 WHERE id = $5`

	_, err := s.pool.Exec(ctx, query, user.Name, user.Surname, user.Login, user.PassHash, user.UserAuth.Id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteUser(ctx context.Context, id uuid.UUID) error {
	const op = "psql.DeleteUser"

	query := `DELETE FROM users WHERE id = $1`

	_, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
