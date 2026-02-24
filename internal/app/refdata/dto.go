package refdata

type GetCityResponse struct {
	ID     int               `json:"id"`
	Name   string            `json:"name"`
	Region GetRegionResponse `json:"region"`
}

type GetRegionResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type GetSportResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IconURL string `json:"icon_url"`
}
