package refdata

import "github.com/google/uuid"

type City struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	RegionID int    `db:"region_id"`
}

type Region struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Sport struct {
	ID      uuid.UUID `db:"id"`
	Name    string    `db:"name"`
	IconURL string    `db:"icon_url"`
}
