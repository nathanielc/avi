package avi

import (
	"errors"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

const detectionThreshold = 0.0

var NoScanAvalaible = errors.New("No scan available")

type Sensor struct {
	partT
	energy   float64
	power    float64
	lastScan *scanResult
}

// Conf format for loading engines from a file
type SensorConf struct {
	Mass   float64
	Radius float64
	Energy float64
	Power  float64
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
		power:  1,
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
		power:  conf.Power,
	}
}

type scanResult struct {
	Position      mgl64.Vec3
	Velocity      mgl64.Vec3
	Health        float64
	Ships         map[int64]Object
	ControlPoints map[int64]Object
}

func (self *Sensor) Scan() (*scanResult, error) {
	if self.used {
		return nil, errors.New("Already used sensor this tick")
	}
	self.used = true

	err := self.ship.ConsumeEnergy(self.energy)
	if err != nil {
		return nil, err
	}
	scan := self.lastScan
	self.lastScan = &scanResult{
		Position:      self.ship.position,
		Velocity:      self.ship.velocity,
		Health:        self.ship.health,
		Ships:         self.searchShips(),
		ControlPoints: self.searchCPs(),
	}
	if scan == nil {
		return nil, NoScanAvalaible
	}
	return scan, nil
}

func (self *Sensor) searchShips() map[int64]Object {
	ships := make(map[int64]Object, len(self.ship.sim.ships))
	for _, ship := range self.ship.sim.ships {
		if ship == self.ship || ship.fleet == self.ship.fleet {
			continue
		}

		distance := ship.position.Sub(self.ship.position).Len()

		i := self.intensity(distance)

		if i > detectionThreshold {
			ships[ship.GetID()] = ship
		}
	}

	return ships
}

func (self *Sensor) searchCPs() map[int64]Object {
	ctlps := make(map[int64]Object, len(self.ship.sim.ctlps))
	for _, ctlp := range self.ship.sim.ctlps {
		distance := ctlp.position.Sub(self.ship.position).Len()

		i := self.intensity(distance)

		if i > detectionThreshold {
			ctlps[ctlp.GetID()] = ctlp
		}
	}

	return ctlps
}

func (self *Sensor) intensity(r float64) float64 {
	area := 4 * math.Pi * r * r
	return self.power / area
}
