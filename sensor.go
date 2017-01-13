package avi

import (
	"errors"
	"math"
	"sync"

	"github.com/go-gl/mathgl/mgl64"
)

const detectionThreshold = 0.0

var NoScanAvalaible = errors.New("No scan available")

type Sensor struct {
	partT
	energy   float64
	power    float64
	lastScan ScanResult

	ships sync.Pool
	ctlps sync.Pool
}

// Conf format for loading engines from a file
type SensorConf struct {
	Mass   float64 `yaml:"mass" json:"mass"`
	Radius float64 `yaml:"radius" json:"radius"`
	Energy float64 `yaml:"energy" json:"energy"`
	Power  float64 `yaml:"power" json:"power"`
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
		ships:  sync.Pool{New: func() interface{} { return make(map[ID]ShipSR) }},
		ctlps:  sync.Pool{New: func() interface{} { return make(map[ID]CtlPSR) }},
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
		ships:  sync.Pool{New: func() interface{} { return make(map[ID]ShipSR) }},
		ctlps:  sync.Pool{New: func() interface{} { return make(map[ID]CtlPSR) }},
	}
}

type ScanResult struct {
	Position      mgl64.Vec3
	Velocity      mgl64.Vec3
	Mass          float64
	Radius        float64
	Health        float64
	Ships         map[ID]ShipSR
	ControlPoints map[ID]CtlPSR

	ships *sync.Pool
	ctlps *sync.Pool
}

func (sr ScanResult) Done() {
	if sr.Ships != nil {
		for k := range sr.Ships {
			delete(sr.Ships, k)
		}
		sr.ships.Put(sr.Ships)
	}
	if sr.ControlPoints != nil {
		for k := range sr.ControlPoints {
			delete(sr.ControlPoints, k)
		}
		sr.ctlps.Put(sr.ControlPoints)
	}
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

func (self *Sensor) Scan() (ScanResult, error) {
	if self.used {
		return ScanResult{}, errors.New("Already used sensor this tick")
	}
	self.used = true

	err := self.ship.ConsumeEnergy(self.energy)
	if err != nil {
		return ScanResult{}, err
	}
	scan := self.lastScan
	self.lastScan = ScanResult{
		Position:      self.ship.position,
		Velocity:      self.ship.velocity,
		Mass:          self.ship.mass,
		Radius:        self.ship.radius,
		Health:        self.ship.health,
		Ships:         self.searchShips(),
		ControlPoints: self.searchCPs(),
		ships:         &self.ships,
		ctlps:         &self.ctlps,
	}
	if scan.Ships == nil {
		return ScanResult{}, NoScanAvalaible
	}
	return scan, nil
}

func (self *Sensor) searchShips() map[ID]ShipSR {
	ships := self.ships.Get().(map[ID]ShipSR)
	for _, ship := range self.ship.sim.ships {
		if ship == self.ship {
			continue
		}

		distance2 := LengthSq(ship.position.Sub(self.ship.position))

		i := self.intensity(distance2)

		if i > detectionThreshold {
			ships[ship.ID()] = ShipSR{
				Fleet:    ship.fleet,
				Position: ship.position,
				Velocity: ship.velocity,
				Radius:   ship.radius,
			}
		}
	}

	return ships
}

func (self *Sensor) searchCPs() map[ID]CtlPSR {
	ctlps := self.ctlps.Get().(map[ID]CtlPSR)
	for _, ctlp := range self.ship.sim.ctlps {
		distance2 := LengthSq(ctlp.position.Sub(self.ship.position))

		i := self.intensity(distance2)

		if i > detectionThreshold {
			ctlps[ctlp.ID()] = CtlPSR{
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

func (self *Sensor) intensity(r2 float64) float64 {
	area := 4 * math.Pi * r2
	return self.power / area
}
