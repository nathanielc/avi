package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Sensor struct {
	partT
	energy float64
}

// Conf format for loading engines from a file
type SensorConf struct {
	Mass float64
	Radius float64
	Energy float64
}

func NewSensor001(pos mgl64.Vec3) *Sensor {
	return &Sensor{
		partT: partT{
			objectT: objectT{
				position: pos,
				mass:     100,
				radius:   0.5,
			},
		},
		energy: 1,
	}
}

func NewSensorFromConf(pos mgl64.Vec3, conf SensorConf) *Sensor {
	return &Sensor{
		partT: partT{
			objectT: objectT{
				position: pos,
				mass:     conf.Mass,
				radius:   conf.Radius,
			},
		},
		energy: conf.Energy,
	}
}

type scanResult struct {
	Position mgl64.Vec3
	Health float64
}

func (self *Sensor) Scan() (scan scanResult, err error) {
	err = self.ship.ConsumeEnergy(self.energy)
	if err != nil {
		return
	}
	scan = scanResult{
		Position: self.ship.position,
		Health: self.ship.health,
	}

	return
}
