package avi

import (
	"errors"
	"github.com/go-gl/mathgl/mgl64"
	"reflect"
	"math"
)

var ErrOutOfEnergy = errors.New("out of energy")

var thrusterType = reflect.TypeOf(&Thruster{})
var engineType = reflect.TypeOf(&Engine{})
var weaponType = reflect.TypeOf(&Weapon{})
var sensorType = reflect.TypeOf(&Sensor{})


type ShipConf struct {
	Pilot        string
	Texture      string
	HullStrength float64 `yaml:"hull_strength"`
	Position     []float64
	Parts        []ShipPartConf
}

//Internal representaion of the ship
type shipT struct {
	pilot Pilot
	fleet   string
	sim     *Simulation
	texture string
	objectT
	parts         []Part
	thrusters     []*Thruster
	weapons       []*Weapon
	engines       []*Engine
	sensors       []*Sensor
	totalEnergy   float64
	currentEnergy float64
}

func newShip(id int64, sim *Simulation, fleet string, pos mgl64.Vec3, pilot Pilot, conf ShipConf) (*shipT, error) {

	newShip := &shipT{
		sim:       sim,
		fleet:     fleet,
		pilot:      pilot,
		parts:     make([]Part, 0),
		thrusters: make([]*Thruster, 0),
		engines:   make([]*Engine, 0),
		weapons:   make([]*Weapon, 0),
		sensors:   make([]*Sensor, 0),
		texture:   conf.Texture,
	}

	newShip.id = id
	newShip.position = pos

	err := newShip.addParts(conf.Parts)
	if err != nil {
		return nil, err
	}

	newShip.determineSize()
	newShip.health = conf.HullStrength * 4 * math.Pi * newShip.radius

	return newShip, nil
}

func (ship *shipT) addParts(partsConf []ShipPartConf) error {
	parts, err := ship.pilot.LinkParts(partsConf, ship.sim.availableParts)
	if err != nil {
		return err
	}
	for _, part := range parts {
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
	accerlation := dir.Normalize().Mul(force / ship.mass)
	ship.ApplyAcc(accerlation)
}

func (ship *shipT) ApplyAcc(dir mgl64.Vec3) {
	ship.velocity = ship.velocity.Add(dir.Mul(timePerTick))
}

func (ship *shipT) Tick() {
	ship.pilot.Tick(ship.sim.tick)
	for _, part := range ship.parts {
		part.reset()
	}
}
