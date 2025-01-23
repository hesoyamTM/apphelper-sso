package psql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hesoyamTM/apphelper-sso/internal/models"
	"github.com/hesoyamTM/apphelper-sso/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(user, pass, host, db string, port int) (*Storage, error) {
	const op = "psql.New"

	ctx := context.Background()
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, pass, host, port, db)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		pool: pool,
	}, nil
}

func (s *Storage) CrateUser(ctx context.Context, name, surname, login string, passHash []byte) (int64, error) {
	const op = "psql.CreateUser"

	query := `INSERT INTO users (name, surname, login, pass_hash) VALUES ($1, $2, $3, $4) RETURNING id`

	row := s.pool.QueryRow(ctx, query, name, surname, login, passHash)

	var id int64
	if err := row.Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
			}
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) ProvideUserById(ctx context.Context, id int64) (models.User, error) {
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
	var id int64
	err := row.Scan(&id, &user.Name, &user.Surname, &user.PassHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user.UserInfo.Id = id
	user.UserAuth.Id = id

	return user, nil
}

func (s *Storage) ProvideUsersById(ctx context.Context, ids []int64) ([]models.User, error) {
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
		var id int64

		err := rows.Scan(&id, &user.Name, &user.Surname, &user.Login, &user.PassHash)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
			}

			return nil, fmt.Errorf("%s: %w", op, err)
		}

		user.UserInfo.Id = id
		user.UserAuth.Id = id

		users = append(users, user)
	}

	return users, nil
}
