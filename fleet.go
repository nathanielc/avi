package avi

type FleetConf struct {
	Name  string     `yaml:"name" json:"name"`
	Ships []ShipConf `yaml:"ships" json:"ships"`
}
