package nathanielc

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/nvcook42/avi"
	"log"
)

type JimSpaceShip struct {
	engines []*avi.Engine
	thrusters []*avi.Thruster
	dir mgl64.Vec3
}

func NewJim(dir mgl64.Vec3) avi.Ship {
	return &JimSpaceShip{dir:dir}
}


func (self *JimSpaceShip) Tick() {
	log.Println("Sending orders")
	for _, engine := range self.engines {
		err := engine.PowerOn(1.0)
		if err != nil {
			log.Println("Failed to power engines", err)
		}
	}
	for _, thruster := range self.thrusters {
		err := thruster.Thrust(self.dir, 1.0)
		if err != nil {
			log.Println("Failed to thrust", err)
		}
	}
}

func (self *JimSpaceShip) GetParts() []avi.Part{
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
	return []avi.Part{
		engine,
		thruster0,
		thruster1,
		weapon,
		sensor,
	}
}
