package refdata

func CityToGetResponse(city *City, region GetRegionResponse) *GetCityResponse {
	dto := GetCityResponse{
		ID:     city.ID,
		Name:   city.Name,
		Region: region,
	}

	return &dto
}

func RegionToGetResponse(region *Region) *GetRegionResponse {
	dto := GetRegionResponse{
		ID:   region.ID,
		Name: region.Name,
	}

	return &dto
}

func SportToGetResponse(sport *Sport) *GetSportResponse {
	dto := GetSportResponse{
		ID:      sport.ID.String(),
		Name:    sport.Name,
		IconURL: sport.IconURL,
	}

	return &dto
}
