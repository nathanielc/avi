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
	lastScan *ScanResult
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

type ScanResult struct {
	Position      mgl64.Vec3
	Velocity      mgl64.Vec3
	Radius        float64
	Health        float64
	Ships         map[int64]ShipSR
	ControlPoints map[int64]CtlPSR
}

type ShipSR struct {
	Position mgl64.Vec3
	Velocity mgl64.Vec3
	Radius   float64
	Fleet    string
}

type CtlPSR struct {
	Position  mgl64.Vec3
	Velocity  mgl64.Vec3
	Radius    float64
	Points    float64
	Influence float64
}

func (self *Sensor) Scan() (*ScanResult, error) {
	if self.used {
		return nil, errors.New("Already used sensor this tick")
	}
	self.used = true

	err := self.ship.ConsumeEnergy(self.energy)
	if err != nil {
		return nil, err
	}
	scan := self.lastScan
	self.lastScan = &ScanResult{
		Position:      self.ship.position,
		Velocity:      self.ship.velocity,
		Radius:        self.ship.radius,
		Health:        self.ship.health,
		Ships:         self.searchShips(),
		ControlPoints: self.searchCPs(),
	}
	if scan == nil {
		return nil, NoScanAvalaible
	}
	return scan, nil
}

func (self *Sensor) searchShips() map[int64]ShipSR {
	ships := make(map[int64]ShipSR, len(self.ship.sim.ships))
	for _, ship := range self.ship.sim.ships {
		if ship == self.ship {
			continue
		}

		distance := ship.position.Sub(self.ship.position).Len()

		i := self.intensity(distance)

		if i > detectionThreshold {
			ships[ship.GetID()] = ShipSR{
				Fleet:    ship.fleet,
				Position: ship.position,
				Velocity: ship.velocity,
				Radius:   ship.radius,
			}
		}
	}

	return ships
}

func (self *Sensor) searchCPs() map[int64]CtlPSR {
	ctlps := make(map[int64]CtlPSR, len(self.ship.sim.ctlps))
	for _, ctlp := range self.ship.sim.ctlps {
		distance := ctlp.position.Sub(self.ship.position).Len()

		i := self.intensity(distance)

		if i > detectionThreshold {
			ctlps[ctlp.GetID()] = CtlPSR{
				Position:  ctlp.position,
				Velocity:  ctlp.velocity,
				Radius:    ctlp.radius,
				Points:    ctlp.points,
				Influence: ctlp.influence,
			}
		}
	}

	return ctlps
}

func (self *Sensor) intensity(r float64) float64 {
	area := 4 * math.Pi * r * r
	return self.power / area
}
