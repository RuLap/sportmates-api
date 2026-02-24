package profile

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID          uuid.UUID `db:"id"`
	FirstName   string    `db:"first_name"`
	LastName    string    `db:"last_name"`
	Gender      string    `db:"gender"`
	BirthDate   time.Time `db:"birth_date"`
	CityID      int       `db:"city_id"`
	AvatarURL   string    `db:"avatar_url"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type UserSport struct {
	UserID  uuid.UUID `db:"user_id"`
	SportID uuid.UUID `db:"sport_id"`
}
