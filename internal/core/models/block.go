package models

type BlockGeometryType string

const (
	BlockGeometryTypePolygon    = "Polygon"
	BlockGeometryTypeLineString = "LineString"
)

func (b BlockGeometryType) IsValid() bool {
	switch b {
	case BlockGeometryTypePolygon, BlockGeometryTypeLineString:
		return true
	default:
		return false
	}
}

type Block struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	GeometryType BlockGeometryType `json:"geometry_type"`     // "Polygon" or "LineString"
	Coordinates  [][]float64       `json:"coordinates"`       // For Polygon and LineString: [[lon, lat], [lon, lat], ...]
	CardID       *string           `json:"card_id,omitempty"` // Optional association to a card, can be null
	// Card         *Card             `json:"card,omitempty"` // Optional metadata for UI display
}

type CreateBlockRequest struct {
	Name         string            `json:"name" binding:"required"`
	GeometryType BlockGeometryType `json:"geometry_type" binding:"required"`
	Coordinates  [][]float64       `json:"coordinates" binding:"required"` // Expecting array of [lon, lat] pairs
}
