package refdata

import (
	"context"
	"log/slog"
)

type Service interface {
	GetCityByID(ctx context.Context, id int) (*GetCityResponse, error)
	GetCitiesByRegionID(ctx context.Context, regionID int) ([]*GetCityResponse, error)

	GetRegionByID(ctx context.Context, id int) (*GetRegionResponse, error)
	GetAllRegions(ctx context.Context) ([]*GetRegionResponse, error)

	GetAllSports(ctx context.Context) ([]*GetSportResponse, error)
	GetSportByID(ctx context.Context, id string) (*GetSportResponse, error)
	GetSportsByIDs(ctx context.Context, ids []string) ([]*GetSportResponse, error)
}

type service struct {
	log          *slog.Logger
	locationRepo LocationRepository
	sportRepo    SportRepository
}

func NewService(log *slog.Logger, locationRepo LocationRepository, sportRepo SportRepository) Service {
	return &service{
		log:          log,
		locationRepo: locationRepo,
		sportRepo:    sportRepo,
	}
}

func (s *service) GetCityByID(ctx context.Context, id int) (*GetCityResponse, error) {
	city, err := s.locationRepo.GetCityByID(ctx, id)
	if err != nil {
		return nil, err
	}

	region, err := s.GetRegionByID(ctx, city.RegionID)
	if err != nil {
		return nil, err
	}

	result := CityToGetResponse(city, *region)

	return result, nil
}

func (s *service) GetCitiesByRegionID(ctx context.Context, regionID int) ([]*GetCityResponse, error) {
	cities, err := s.locationRepo.GetCitiesByRegionID(ctx, regionID)
	if err != nil {
		return nil, err
	}

	region, err := s.GetRegionByID(ctx, regionID)
	if err != nil {
		return nil, err
	}

	result := make([]*GetCityResponse, len(cities))
	for i, city := range cities {
		result[i] = CityToGetResponse(city, *region)
	}

	return result, nil
}

func (s *service) GetRegionByID(ctx context.Context, id int) (*GetRegionResponse, error) {
	region, err := s.locationRepo.GetRegionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := RegionToGetResponse(region)

	return result, nil
}

func (s *service) GetAllRegions(ctx context.Context) ([]*GetRegionResponse, error) {
	regions, err := s.locationRepo.GetAllRegions(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*GetRegionResponse, len(regions))
	for i, region := range regions {
		result[i] = RegionToGetResponse(region)
	}

	return result, nil
}

func (s *service) GetAllSports(ctx context.Context) ([]*GetSportResponse, error) {
	sports, err := s.sportRepo.GetAllSports(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*GetSportResponse, len(sports))
	for i, sport := range sports {
		result[i] = SportToGetResponse(sport)
	}

	return result, nil
}

func (s *service) GetSportByID(ctx context.Context, id string) (*GetSportResponse, error) {
	sport, err := s.sportRepo.GetSportByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := SportToGetResponse(sport)

	return result, nil
}

func (s *service) GetSportsByIDs(ctx context.Context, ids []string) ([]*GetSportResponse, error) {
	sports, err := s.sportRepo.GetSportsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := make([]*GetSportResponse, len(sports))
	for i, sport := range sports {
		result[i] = SportToGetResponse(sport)
	}

	return result, nil
}
