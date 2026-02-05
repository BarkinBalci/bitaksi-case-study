package models

type GeoJSON struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}

type DriverLocation struct {
	Location GeoJSON `bson:"location"`
}

func NewDriverLocation(lat, lon float64) *DriverLocation {
	return &DriverLocation{
		Location: GeoJSON{
			Type:        "Point",
			Coordinates: []float64{lon, lat},
		},
	}
}

type BulkResult struct {
	Total      int
	Successful int
	Failed     int
}

type SearchResult struct {
	DriverID  string
	Latitude  float64
	Longitude float64
	Distance  float64
}
