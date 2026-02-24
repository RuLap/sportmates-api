package profile

import "github.com/RuLap/sportmates-api/internal/app/refdata"

type GetProfileResponse struct {
	ID          string                     `json:"id"`
	FirstName   string                     `json:"first_name"`
	LastName    string                     `json:"last_name"`
	Gender      string                     `json:"gender"`
	BirthDate   string                     `json:"birth_date"`
	AvatarURL   string                     `json:"avatar_url"`
	Description string                     `json:"description"`
	City        refdata.GetCityResponse    `json:"city"`
	Sports      []refdata.GetSportResponse `json:"sports"`
}

type SaveProfileRequest struct {
	ID          string   `json:"id,omitempty" validate:"uuid"`
	FirstName   string   `json:"first_name" validate:"required,max=50"`
	LastName    string   `json:"last_name" validate:"required,max=50"`
	Gender      string   `json:"gender" validate:"required"`
	BirthDate   string   `json:"birth_date" validate:"required,datetime=2006-01-02"`
	AvatarURL   string   `json:"avatar_url" validate:"required,url"`
	Description string   `json:"description" validate:"required"`
	CityID      int      `json:"city_id" validate:"required,int"`
	Sports      []string `json:"sports"`
}

type GetUploadURLResponse struct {
	URL string
}

type ConfirmUploadAvatarResponse struct {
	URL string
}
