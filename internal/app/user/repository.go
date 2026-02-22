package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	uuid "github.com/google/uuid"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserAlreadyExists   = errors.New("пользователь с таким email существует")
	InvalidEmailOrPassword = errors.New("неверный email или пароль")
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	CreateUser(ctx context.Context, user *User) (*string, error)
	MakeEmailConfirmed(ctx context.Context, userID string) error
	GetByEmailProvider(ctx context.Context, email string, provider Provider) (*User, error)
	GetPasswordHashByEmail(ctx context.Context, email string) (*string, error)
	Close()
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool}
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, email_confirmed
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.EmailConfirmed,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить пользователя по email provider: %w", err)
	}

	return &user, nil
}

func (r *repository) CreateUser(ctx context.Context, user *User) (*string, error) {
	query := `
		INSERT INTO users (email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var userID string
	err := r.pool.QueryRow(
		ctx,
		query,
		user.Email,
		user.Password,
	).Scan(&userID)

	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("не удалось создать пользователя: %w", err)
	}

	return &userID, nil
}

func (r *repository) MakeEmailConfirmed(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET email_confirmed = TRUE
		where id = $1
	`

	result, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("не удалось подтверить почту: %w", err)
	}

	if result.RowsAffected() == 0 {
		return InvalidEmailOrPassword
	}

	return nil
}

func (r *repository) GetByEmailProvider(ctx context.Context, email string, provider Provider) (*User, error) {
	query := `
		SELECT id, email
		FROM users
		WHERE email = $1 AND provider = $2
	`

	var user User
	err := r.pool.QueryRow(
		ctx,
		query,
		email,
		provider,
	).Scan(
		&user.ID,
		&user.Email,
	)

	if err != nil {
		return nil, fmt.Errorf("не удалось получить пользователя по email provider: %w", err)
	}

	return &user, nil
}

func (r *repository) GetPasswordHashByEmail(ctx context.Context, email string) (*string, error) {
	query := `
		SELECT password FROM users WHERE email = $1
	`

	var password *string
	err := r.pool.QueryRow(ctx, query, email).Scan(&password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("не удалось найти пользователя: %w", err)
		}
		return nil, fmt.Errorf("не удалось получить хэш пароля: %w", err)
	}

	return password, nil
}

func (r *repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

func isUniqueConstraintError(err error) bool {
	if pgErr, ok := err.(*pgx.PgError); ok {
		return pgErr.Code == "23505"
	}
	return false
}
