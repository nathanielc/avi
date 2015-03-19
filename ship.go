package avi

import (
	"errors"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/nvcook42/avi/logger"
	"reflect"
)

var ErrOutOfEnergy = errors.New("out of energy")

var thrusterType = reflect.TypeOf(&Thruster{})
var engineType = reflect.TypeOf(&Engine{})
var weaponType = reflect.TypeOf(&Weapon{})
var sensorType = reflect.TypeOf(&Sensor{})

//Exported interface for ships
type Ship interface {
	Tick()
	LinkParts([]ShipPartConf, *PartsConf) ([]Part, error)
}

//Internal representaion of the ship
type shipT struct {
	ship          Ship
	fleet         string
	sim           *Simulation
	objectT
	parts         []Part
	thrusters     []*Thruster
	weapons       []*Weapon
	engines       []*Engine
	sensors       []*Sensor
	totalEnergy   float64
	currentEnergy float64
}

type shipFactory func() Ship
var registeredShips = make(map[string]shipFactory)

//Register a ship to make it available
func RegisterShip(name string, sf shipFactory) {
	registeredShips[name] = sf
}

// Get a registered ship by name
func getShipByName(name string) Ship {
	if sf, ok := registeredShips[name]; ok {
		return sf()
	}
	return nil
}


func newShip(sim *Simulation, fleet string, pos mgl64.Vec3, ship Ship, parts []ShipPartConf) (*shipT, error) {

	newShip := &shipT{
		sim: sim,
		fleet: fleet,
		ship:    ship,
		parts: make([]Part, 0),
		thrusters: make([]*Thruster, 0),
		engines: make([]*Engine, 0),
		weapons: make([]*Weapon, 0),
		sensors: make([]*Sensor, 0),
	}

	newShip.position = pos


	err := newShip.addParts(parts)
	if err != nil {
		return nil, err
	}

	newShip.determineSize()


	return newShip, nil
}

func (ship *shipT) addParts(partsConf []ShipPartConf) error {
	parts, err := ship.ship.LinkParts(partsConf, ship.sim.availableParts)
	if err != nil {
		return err
	}
	for _, part := range parts {
		logger.Debug.Println("Adding part to ship")
		ship.parts = append(ship.parts, part)
		part.setShip(ship)

		ship.mass += part.GetMass()

		switch reflect.TypeOf(part) {
		case thrusterType:
			t := part.(*Thruster)
			ship.thrusters = append(ship.thrusters, t)
		case engineType:
			e := part.(*Engine)
			ship.engines = append(ship.engines, e)
		case weaponType:
			w := part.(*Weapon)
			ship.weapons = append(ship.weapons, w)
		case sensorType:
			s := part.(*Sensor)
			ship.sensors = append(ship.sensors, s)
		}
	}
	return nil
}


func (ship *shipT) determineSize() {

	maxRadius := 0.0

	for _, part := range ship.parts {
		radius := part.GetPosition().Len() + part.GetRadius()
		if radius > maxRadius {
			maxRadius = radius
		}
	}
	ship.radius = maxRadius
}

//Determine how much power the ship is supplying
func (ship *shipT) Energize() {
	ship.totalEnergy = 0
	for _, engine := range ship.engines {
		ship.totalEnergy += engine.getOutput()
	}
	ship.currentEnergy = ship.totalEnergy
}

// Consume a given amount of energy for another component on the ship
func (ship *shipT) ConsumeEnergy(amount float64) error {
	ship.currentEnergy -= amount
	if ship.currentEnergy < 0 {
		ship.currentEnergy = 0
		return ErrOutOfEnergy
	}
	return nil
}

// Apply a given amount of thrust in a certain direction
func (ship *shipT) ApplyThrust(dir mgl64.Vec3, force float64) {
	accerlation := dir.Mul(force / ship.mass)
	ship.velocity = ship.velocity.Add(accerlation.Mul(timePerTick))
}

func (ship *shipT) Tick() {
	ship.ship.Tick()
}


