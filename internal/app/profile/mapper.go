package profile

import (
	"time"

	"github.com/RuLap/sportmates-api/internal/app/refdata"
)

func ProfileToGetResponse(profile *Profile, city *refdata.GetCityResponse, sports []refdata.GetSportResponse) *GetProfileResponse {
	dto := GetProfileResponse{
		ID:          profile.ID.String(),
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		Gender:      profile.Gender,
		BirthDate:   profile.BirthDate.String(),
		AvatarURL:   profile.AvatarURL,
		Description: profile.Description,
	}

	dto.Sports = sports
	dto.City = *city

	return &dto
}

func SaveRequestToProfile(dto *SaveProfileRequest) (*Profile, error) {
	model := Profile{
		FirstName:   dto.FirstName,
		LastName:    dto.LastName,
		Gender:      dto.Gender,
		AvatarURL:   dto.AvatarURL,
		Description: dto.Description,
		CityID:      dto.CityID,
	}

	birthDate, err := time.Parse(time.DateOnly, dto.BirthDate)
	if err != nil {
		return nil, err
	}
	model.BirthDate = birthDate

	return &model, nil
}
