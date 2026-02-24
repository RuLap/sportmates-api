package profile

import (
	"context"
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("profile not found")
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Profile, error)
	Create(ctx context.Context, model *Profile, sportIDs []string) (*Profile, error)
	Update(ctx context.Context, model *Profile, sportIDs []string) (*Profile, error)
	GetUserSports(ctx context.Context, userID uuid.UUID) ([]*UserSport, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Profile, error) {
	const query = `
		SELECT id, first_name, last_name, gender, birth_date, city_id, avatar_url, description
		FROM profiles
		WHERE id = $1
	`

	var profile Profile
	err := r.db.QueryRow(ctx, query, id).Scan(
		&profile.ID,
		&profile.FirstName,
		&profile.LastName,
		&profile.Gender,
		&profile.BirthDate,
		&profile.CityID,
		&profile.AvatarURL,
		&profile.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find profile by ID: %w", err)
	}

	return &profile, nil
}

func (r *repository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Profile, error) {
	if len(ids) == 0 {
		return []*Profile{}, nil
	}

	query := `
		SELECT id, first_name, last_name, gender, birth_date, avatar_url, description
		FROM profiles WHERE id = ANY($1::uuid[])
	`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(
			&p.ID,
			&p.FirstName,
			&p.LastName,
			&p.Gender,
			&p.BirthDate,
			&p.AvatarURL,
			&p.Description,
		); err != nil {
			return nil, err
		}
		result = append(result, &p)
	}

	if result == nil {
		result = []*Profile{}
	}

	return result, nil
}

func (r *repository) Create(ctx context.Context, profile *Profile, sportIDs []string) (*Profile, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const query = `
		INSERT INTO profiles (id, first_name, last_name, gender, birth_date, city_id, avatar_url, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err = tx.QueryRow(
		ctx,
		query,
		profile.ID,
		profile.FirstName,
		profile.LastName,
		profile.Gender,
		profile.BirthDate,
		profile.CityID,
		profile.AvatarURL,
		profile.Description,
	).Scan(&profile.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	err = r.updateUserSports(ctx, tx, profile.ID, sportIDs)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (r *repository) Update(ctx context.Context, profile *Profile, sportIDs []string) (*Profile, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const query = `
		UPDATE profiles 
		SET first_name = $2, last_name = $3, gender = $4, birth_date = $5,
			city_id = $6, avatar_url = $7, description = $8
		WHERE id = $1
		RETURNING id
	`

	err = tx.QueryRow(
		ctx,
		query,
		profile.ID,
		profile.FirstName,
		profile.LastName,
		profile.Gender,
		profile.BirthDate,
		profile.CityID,
		profile.AvatarURL,
		profile.Description,
	).Scan(&profile.ID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	err = r.updateUserSports(ctx, tx, profile.ID, sportIDs)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (r *repository) GetUserSports(ctx context.Context, userID uuid.UUID) ([]*UserSport, error) {
	const query = `
		SELECT user_id, sport_id
		FROM user_sports
		WHERE user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user sports: %w", err)
	}
	defer rows.Close()

	var result []*UserSport
	for rows.Next() {
		var us UserSport
		err := rows.Scan(&us.UserID, &us.SportID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user sport: %w", err)
		}
		result = append(result, &us)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	if result == nil {
		result = []*UserSport{}
	}

	return result, nil
}

func (r *repository) updateUserSports(ctx context.Context, tx pgx.Tx, userID uuid.UUID, sportIDs []string) error {
	oldSports, err := r.GetUserSports(ctx, userID)
	if err != nil {
		return err
	}

	oldSet := mapset.NewSet[string]()
	for _, s := range oldSports {
		oldSet.Add(s.SportID.String())
	}

	newSet := mapset.NewSet(sportIDs...)

	toInsert := newSet.Difference(oldSet)
	toDelete := oldSet.Difference(newSet)

	if toInsert.Cardinality() > 0 {
		_, err := tx.Exec(ctx, `
			INSERT INTO user_sports (user_id, sport_id)
			SELECT $1, x
			FROM UNNEST($2::uuid[]) AS x
			ON CONFLICT DO NOTHING
		`, userID, toInsert.ToSlice())
		if err != nil {
			return err
		}
	}

	if toDelete.Cardinality() > 0 {
		_, err := tx.Exec(ctx, `
			DELETE FROM user_sports
			WHERE user_id = $1
			AND sport_id = ANY($2)
		`, userID, toDelete.ToSlice())
		if err != nil {
			return err
		}
	}

	return nil
}
