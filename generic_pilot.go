package avi

import (
	"errors"
	"fmt"
)

type GenericPilot struct {
	Fleet     string
	Engines   []*Engine
	Thrusters []*Thruster
	Weapons   []*Weapon
	Sensors   []*Sensor
}

func (self *GenericPilot) JoinFleet(fleet string) {
	self.Fleet = fleet
}

func (self *GenericPilot) LinkParts(shipParts []ShipPartConf, availableParts PartSetConf) ([]Part, error) {
	parts := make([]Part, 0)
	self.Engines = make([]*Engine, 0)
	self.Thrusters = make([]*Thruster, 0)
	self.Weapons = make([]*Weapon, 0)
	self.Sensors = make([]*Sensor, 0)
	for _, part := range shipParts {
		switch part.Type {
		case "engine":
			if engineConf, ok := availableParts.Engines[part.Name]; !ok {
				return nil, PartNotAvailable(part.Name)
			} else {
				pos, err := sliceToVec(part.Position)
				if err != nil {
					return nil, err
				}
				engine := NewEngineFromConf(pos, engineConf)
				self.Engines = append(self.Engines, engine)
				parts = append(parts, engine)
			}
		case "thruster":
			if thrusterConf, ok := availableParts.Thrusters[part.Name]; !ok {
				return nil, PartNotAvailable(part.Name)
			} else {
				pos, err := sliceToVec(part.Position)
				if err != nil {
					return nil, err
				}
				thruster := NewThrusterFromConf(pos, thrusterConf)
				self.Thrusters = append(self.Thrusters, thruster)
				parts = append(parts, thruster)
			}
		case "weapon":
			if weaponConf, ok := availableParts.Weapons[part.Name]; !ok {
				return nil, PartNotAvailable(part.Name)
			} else {
				pos, err := sliceToVec(part.Position)
				if err != nil {
					return nil, err
				}
				weapon := NewWeaponFromConf(pos, weaponConf)
				self.Weapons = append(self.Weapons, weapon)
				parts = append(parts, weapon)
			}
		case "sensor":
			if sensorConf, ok := availableParts.Sensors[part.Name]; !ok {
				return nil, PartNotAvailable(part.Name)
			} else {
				pos, err := sliceToVec(part.Position)
				if err != nil {
					return nil, err
				}
				sensor := NewSensorFromConf(pos, sensorConf)
				self.Sensors = append(self.Sensors, sensor)
				parts = append(parts, sensor)
			}
		default:
			return nil, errors.New(fmt.Sprintf("Unknown part type '%s'", part.Type))
		}
	}

	return parts, nil
}
