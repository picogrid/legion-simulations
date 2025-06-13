package models

type Category string

func (c Category) String() string {
	return string(c)
}

const (
	// CATEGORY_DEVICE is an entity representing a physical or logical unit that contains entities.
	// Ex. A Portal that contains two entities: a Radar entity and an Axis Camera entity.
	CATEGORY_DEVICE Category = "DEVICE"
	// CATEGORY_DETECTION is an entity representing a specific detection.
	// EX. A detection of a person, a vehicle, a gunshot or etc.
	CATEGORY_DETECTION Category = "DETECTION"
	// CATEGORY_ALERT is an entity representing an alert.
	// EX. An alert of a person, a vehicle, a gunshot, a storm, etc..
	CATEGORY_ALERT Category = "ALERT"
	// CATEGORY_WEATHER is an entity representing a weather condition.
	// Ex. A weather condition or system
	CATEGORY_WEATHER Category = "WEATHER"
	// CATEGORY_GEOMETRIC is an entity representing a geometric object.
	// Ex. A point, a line, a polygon, a circle, etc.
	CATEGORY_GEOMETRIC Category = "GEOMETRIC"
	// CATEGORY_ZONE is an entity representing a zone or region.
	// Ex. A zone or region of a building, a zone or region of a road, etc.
	CATEGORY_ZONE Category = "ZONE"
	// CATEGORY_SENSOR is an entity representing a sensor.
	// Ex. A Camera, Radar, etc.
	CATEGORY_SENSOR Category = "SENSOR"
	// CATEGORY_VEHICLE is an entity representing a vehicle.
	// Ex. A Car, Tank, Truck
	CATEGORY_VEHICLE Category = "VEHICLE"
	// CATEGORY_UXV is an entity representing an Unmanned Vehicle.
	// Ex. A UAV, USV, USG, etc.
	CATEGORY_UXV Category = "UXV"
)
