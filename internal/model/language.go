package model

type LanguageInfo struct {
	Code        string    `json:"code"`        // e.g., "eng", "tur"
	Name        string    `json:"name"`        // e.g., "English", "Turkish"
	Coordinates []float64 `json:"coordinates"` // [lat, long]
	Flag        string    `json:"flag"`        // URL to flag image
	Country     string    `json:country`
}
