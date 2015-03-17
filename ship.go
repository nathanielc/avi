package avi

import (
	"errors"
	"github.com/go-gl/mathgl/mgl64"
)

var ErrOutofEnergy = errors.New("out of energy")

//Exported interface for ships
type Ship interface {
	Launch(orders chan Order, results chan OrderResult)
	GetThrusters() []thruster
	GetWeapons() []weapon
	GetEngines() []engine
	GetSensors() []sensor
}

//Internal representaion of the ship
type shipT struct {
	ship          Ship
	orders        chan Order
	results       chan OrderResult
	thrusters     []*thruster
	weapons       []*weapon
	engines       []*engine
	sensors       []*sensor
	position      mgl64.Vec3
	velocity      mgl64.Vec3
	totalMass     float64
	totalEnergy   float64
	currentEnergy float64
}

//Tell the ship to boot up. This has to happen anytime a part is
// lost and at the start
func (ship *shipT) Boot() {
	ship.totalMass = 0
	for _, thruster := range ship.thrusters {
		ship.totalMass += thruster.Mass
	}
	for _, weapon := range ship.weapons {
		ship.totalMass += weapon.Mass
	}
	for _, engine := range ship.engines {
		ship.totalMass += engine.Mass
	}
	for _, sensor := range ship.sensors {
		ship.totalMass += sensor.Mass
	}
}

//Determine how much power the ship is supplying
func (ship *shipT) Energize() {
	ship.totalEnergy = 0
	for _, engine := range ship.engines {
		ship.totalEnergy += engine.GetOutput()
	}
	ship.currentEnergy = ship.totalEnergy
}

// Consume a given amount of energy for another component on the ship
func (ship *shipT) ConsumeEnergy(amount float64) error {
	self.currentEnergy -= amount
	if self.currentEnergy < 0 {
		self.currentEnergy = 0
		return ErrOutofEnergy
	}
}

// Apply a given amount of thrust in a certain direction
func (ship *shipT) ApplyThrust(dir mgl64.Vec3, force float64) {

}
