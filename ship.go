package avi

import (
	"errors"
	"github.com/go-gl/mathgl/mgl64"
	"log"
	"reflect"
)

var ErrOutOfEnergy = errors.New("out of energy")

var thrusterType = reflect.TypeOf(&Thruster{})
var engineType = reflect.TypeOf(&Engine{})
var weaponType = reflect.TypeOf(&Weapon{})
var sensorType = reflect.TypeOf(&Sensor{})

//Exported interface for ships
type Ship interface {
	Launch()
	GetParts() map[string]Part
}

//Internal representaion of the ship
type shipT struct {
	ship          Ship
	parts         map[string]Part
	thrusters     []*Thruster
	weapons       []*Weapon
	engines       []*Engine
	sensors       []*Sensor
	position      mgl64.Vec3
	velocity      mgl64.Vec3
	totalMass     float64
	totalEnergy   float64
	currentEnergy float64
}


func newShip(ship Ship) *shipT{

	newShip := &shipT{
		ship:    ship,
		parts: make(map[string]Part),
		thrusters: make([]*Thruster, 0),
		engines: make([]*Engine, 0),
		weapons: make([]*Weapon, 0),
		sensors: make([]*Sensor, 0),
	}
	
	for id, part := range ship.GetParts() {
		log.Println(id, part)
		newShip.addPart(id, part)

		switch reflect.TypeOf(part) {
		case thrusterType:
			t := part.(*Thruster)
			newShip.thrusters = append(newShip.thrusters, t)
		case engineType:
			e := part.(*Engine)
			newShip.engines = append(newShip.engines, e)
		case weaponType:
			w := part.(*Weapon)
			newShip.weapons = append(newShip.weapons, w)
		case sensorType:
			s := part.(*Sensor)
			newShip.sensors = append(newShip.sensors, s)
		}
	}

	return newShip
}

func (ship *shipT) addPart(id string, part Part) bool {
	_, ok := ship.parts[id]
	if !ok {
		ship.parts[id] = part
	}

	return !ok
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

}

func (ship *shipT) ProcessOrders() {

	for _, part := range ship.parts {
		order := part.GetOrder()
		if order != nil {
			order(ship)
		}
	}
}
