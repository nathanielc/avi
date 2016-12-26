package avi

import (
	"errors"
	"fmt"
)

func PartNotAvailable(name string) error {
	return errors.New(fmt.Sprintf("part '%s' not available", name))
}

type Part interface {
	Object
	setShip(*shipT)
	reset()
}

type ShipPartConf struct {
	Name     string    `yaml:"name" json:"name"`
	Position []float64 `yaml:"position" json:"position"`
	Type     string    `yaml:"type" json:"type"`
}

type partT struct {
	objectT
	ship *shipT
	used bool
}

func (part *partT) setShip(ship *shipT) {
	part.ship = ship
}

func (part *partT) reset() {
	part.used = false
}

type PartSetConf struct {
	Engines   map[string]EngineConf   `yaml:"engines" json:"engines"`
	Thrusters map[string]ThrusterConf `yaml:"thrusters" json:"thrusters"`
	Weapons   map[string]WeaponConf   `yaml:"weapons" json:"weapons"`
	Sensors   map[string]SensorConf   `yaml:"sensors" json:"sensors"`
}
