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
	dir   mgl64.Vec3
	fired bool
}

func NewJim() avi.Ship {
	return &JimSpaceShip{dir: mgl64.Vec3{1, 1, 1}}
}

func (self *JimSpaceShip) Tick() {
	for _, engine := range self.Engines {
		err := engine.PowerOn(1.0)
		if err != nil {
			logger.Debug.Println("Failed to power engines", err)
		}
	}
	//for _, thruster := range self.Thrusters {
	//	err := thruster.Thrust(self.dir, 1.0)
	//	if err != nil {
	//		logger.Debug.Println("Failed to thrust", err)
	//	}
	//}
	scan, err := self.Sensors[0].Scan()
	if err != nil {
		logger.Debug.Println("Failed to scan", err)
		return
	}
	logger.Debug.Println("Jim health", scan.Health)
	if !self.fired || true {
		self.fired = true
		for _, weapon := range self.Weapons {
			err := weapon.Fire(scan.Position.Mul(-1))
			if err != nil {
				logger.Debug.Println("Failed to fire", err)
			} else {

			}
		}
	}
}
