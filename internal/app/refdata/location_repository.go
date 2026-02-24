package refdata

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCityNotFound   = errors.New("city not found")
	ErrRegionNotFound = errors.New("region not found")
)

type LocationRepository interface {
	GetCityByID(ctx context.Context, id int) (*City, error)
	GetCitiesByRegionID(ctx context.Context, regionID int) ([]*City, error)

	GetRegionByID(ctx context.Context, id int) (*Region, error)
	GetAllRegions(ctx context.Context) ([]*Region, error)
}

type locationRepository struct {
	db *pgxpool.Pool
}

func NewLocationRepository(db *pgxpool.Pool) LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) GetCityByID(ctx context.Context, id int) (*City, error) {
	query := `
		SELECT id, name, region_id
		FROM cities
		WHERE id = $1
	`

	var city City
	err := r.db.QueryRow(ctx, query, id).Scan(&city.ID, &city.Name, &city.RegionID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCityNotFound
		}
		return nil, fmt.Errorf("query city by id: %w", err)
	}

	return &city, nil
}

func (r *locationRepository) GetCitiesByRegionID(ctx context.Context, regionID int) ([]*City, error) {
	query := `
		SELECT id, name, region_id
		FROM cities
		WHERE region_id = $1
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, regionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cities := make([]*City, 0)
	for rows.Next() {
		var city City
		err := rows.Scan(&city.ID, &city.Name, &city.RegionID)
		if err != nil {
			return nil, err
		}

		cities = append(cities, &city)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cities, nil
}

func (r *locationRepository) GetAllRegions(ctx context.Context) ([]*Region, error) {
	query := `
		SELECT id, name
		FROM regions
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	regions := make([]*Region, 0)
	for rows.Next() {
		var region Region
		err := rows.Scan(&region.ID, &region.Name)
		if err != nil {
			return nil, err
		}

		regions = append(regions, &region)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return regions, nil
}

func (r *locationRepository) GetRegionByID(ctx context.Context, id int) (*Region, error) {
	query := `
		SELECT id, name
		FROM regions
		WHERE id = $1
	`

	var region Region
	err := r.db.QueryRow(ctx, query, id).Scan(&region.ID, &region.Name)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRegionNotFound
		}
		return nil, fmt.Errorf("query region by id: %w", err)
	}

	return &region, nil
}
