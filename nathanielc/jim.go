package nathanielc

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/nvcook42/avi"
	"log"
)

type JimSpaceShip struct {
	engines []*avi.Engine
	thrusters []*avi.Thruster
}

func NewJim() avi.Ship {
	return &JimSpaceShip{}
}

func (self *JimSpaceShip) Launch() {
	go self.think()
}

func (self *JimSpaceShip) think() {
	tick := 0
	for {
		tick++
		log.Println("Sending orders")
		tr0 := self.thrusters[0].Thrust(mgl64.Vec3{1,1,1}, 1.0)
		tr1 := self.thrusters[1].Thrust(mgl64.Vec3{1,1,1}, 1.0)
		log.Println("recv")
		if e0 := <-tr0; e0 == nil {
			log.Println("Successfull thrust0")
		} else {
			log.Println("Failed thrust0", e0)
		}
		if e1 := <-tr1; e1 == nil{
			log.Println("Successfull thrust1")
		} else {
			log.Println("Failed thrust1", e1)
		}
	}
}

func (self *JimSpaceShip) GetParts() map[string]avi.Part{
	self.engines = make([]*avi.Engine, 1)
	engine := avi.NewEngine001(mgl64.Vec3{0, 0, 1})
	self.engines[0] = engine

	self.thrusters = make([]*avi.Thruster, 2)
	thruster0 := avi.NewThruster001(mgl64.Vec3{0, -1, 0})
	thruster1 := avi.NewThruster001(mgl64.Vec3{0, 1, 0})
	self.thrusters[0] = thruster0
	self.thrusters[1] = thruster1

	weapon := avi.NewWeapon001(mgl64.Vec3{0, 0, 1})
	sensor := avi.NewSensor001(mgl64.Vec3{0, 0, 1})
	return map[string]avi.Part{
		"engine" : engine,
		"t0" : thruster0,
		"t1" : thruster1,
		"gun" : weapon,
		"sensor" : sensor,
	}
}
