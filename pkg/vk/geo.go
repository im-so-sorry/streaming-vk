package vk

type CoordinateType struct {
	latitude  int
	longitude int
}

type GeoType struct {
	Type string
	Coordinates CoordinateType
	Place interface{}
}
