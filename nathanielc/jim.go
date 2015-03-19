package nathanielc

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/nvcook42/avi"
	"github.com/nvcook42/avi/logger"
)

func init() {
	avi.RegisterShip("jim", NewJim)
}

type JimSpaceShip struct {
	avi.GenericShip
	dir mgl64.Vec3
}

func NewJim() avi.Ship {
	return &JimSpaceShip{dir:mgl64.Vec3{1,1,1}}
}


func (self *JimSpaceShip) Tick() {
	logger.Debugln("Sending orders")
	for _, engine := range self.Engines {
		err := engine.PowerOn(1.0)
		if err != nil {
			logger.Debugln("Failed to power engines", err)
		}
	}
	for _, thruster := range self.Thrusters {
		err := thruster.Thrust(self.dir, 1.0)
		if err != nil {
			logger.Debugln("Failed to thrust", err)
		}
	}
}

