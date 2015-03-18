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
	Tick()
	GetParts() []Part
}

//Internal representaion of the ship
type shipT struct {
	ship          Ship
	objectT
	parts         []Part
	thrusters     []*Thruster
	weapons       []*Weapon
	engines       []*Engine
	sensors       []*Sensor
	totalEnergy   float64
	currentEnergy float64
}


func newShip(pos mgl64.Vec3, ship Ship) *shipT{

	newShip := &shipT{
		ship:    ship,
		parts: make([]Part, 0),
		thrusters: make([]*Thruster, 0),
		engines: make([]*Engine, 0),
		weapons: make([]*Weapon, 0),
		sensors: make([]*Sensor, 0),
	}

	newShip.position = pos


	newShip.addParts()

	newShip.determineSize()


	return newShip
}

func (ship *shipT) addParts() {
	for _, part := range ship.ship.GetParts() {
		log.Println(part)
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
	log.Println("Ship energized", ship.totalEnergy)
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


// Single Direction ship

type oneDirShip struct {
	engine *Engine
	thruster *Thruster
	dir mgl64.Vec3
}

func newOneDirShip(dir mgl64.Vec3) Ship {
	return &oneDirShip{dir:dir}
}


func (self *oneDirShip) Tick() {
	self.engine.PowerOn(1.0)
	self.thruster.Thrust(self.dir, 1.0)
}

func (self *oneDirShip) GetParts() []Part{
	self.engine = NewEngine001(mgl64.Vec3{0, 0, 1})

	self.thruster = NewThruster001(mgl64.Vec3{0, -1, 0})
	return []Part{
		self.engine,
		self.thruster,
	}
}
