package avi

type RulesConf struct {
	Score        float64 `yaml:"score" json:"score"`
	MaxFleetMass float64 `yaml:"max_fleet_mass" json:"max_fleet_mass"`
}

type MapConf struct {
	Radius         int64              `yaml:"radius" json:"radius"`
	Asteroids      []AsteroidConf     `yaml:"asteroids" json:"asteroids"`
	ControlPoints  []ControlPointConf `yaml:"control_points" json:"control_points"`
	StartingPoints [][]float64        `yaml:"starting_points" json:"starting_points"`
	Rules          RulesConf          `yaml:"rules" json:"rules"`
}
