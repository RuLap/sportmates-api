package refdata

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("sport not found")
)

type SportRepository interface {
	GetAllSports(ctx context.Context) ([]*Sport, error)
	GetSportByID(ctx context.Context, id string) (*Sport, error)
	GetSportsByIDs(ctx context.Context, ids []string) ([]*Sport, error)
}

type sportRepository struct {
	db *pgxpool.Pool
}

func NewSportRepository(db *pgxpool.Pool) SportRepository {
	return &sportRepository{db: db}
}

func (r *sportRepository) GetAllSports(ctx context.Context) ([]*Sport, error) {
	const query = `
		SELECT id, name, icon_url
		FROM sports
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sports := make([]*Sport, 0)
	for rows.Next() {
		var sport Sport
		err := rows.Scan(
			&sport.ID,
			&sport.Name,
			&sport.IconURL,
		)
		if err != nil {
			return nil, err
		}
		sports = append(sports, &sport)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sports, nil
}

func (r *sportRepository) GetSportByID(ctx context.Context, id string) (*Sport, error) {
	const query = `
		SELECT id, name, icon_url
		FROM sports
		WHERE id = $1::uuid
	`

	var sport Sport
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sport.ID,
		&sport.Name,
		&sport.IconURL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query sport by id: %w", err)
	}

	return &sport, nil
}

func (r *sportRepository) GetSportsByIDs(ctx context.Context, ids []string) ([]*Sport, error) {
	if len(ids) == 0 {
		return []*Sport{}, nil
	}

	query := `SELECT id, name, icon_url FROM sports WHERE id = ANY($1::uuid[])`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sports := make([]*Sport, 0)
	for rows.Next() {
		var s Sport
		if err := rows.Scan(&s.ID, &s.Name, &s.IconURL); err != nil {
			return nil, err
		}
		sports = append(sports, &s)
	}

	return sports, nil
}
