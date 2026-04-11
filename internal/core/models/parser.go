package models

import "time"

// The structure matching the Python Pandas JSON output
type ParsedEntry struct {
	Day       string `json:"Day"`
	Date      string `json:"Date"`      // e.g. "1-Mar"
	Time      string `json:"Time"`      // e.g. "10:55"
	Venue     string `json:"Venue"`     // e.g. "Kingdom Hall"
	Conductor string `json:"Conductor"` // e.g. "Chinuji Ogwu"
	Remark    string `json:"Remark"`    // e.g. "House-to-House"
}

// The structure matching the python GeoJson output
type MapBlocksResponse struct {
	Address     string        `json:"-"`
	BlocksCount int           `json:"blocks_count"`
	Polygons    [][][]float64 `json:"polygons"` // Nested array: [Polygon][Point][Lat/Lon]
}

// CacheWrapper allows us to save metadata like timestamps
type MapCacheWrapper struct {
	Data        MapBlocksResponse `json:"data"`
	LastUpdated time.Time         `json:"last_updated"`
}
