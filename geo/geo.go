package geo

type Point struct {
	Type        string    `json:"type,omitempty" bson:"type,omitempty"`
	Coordinates []float64 `json:"coordinates,omitempty" bson:"coordinates,omitempty"`
}
