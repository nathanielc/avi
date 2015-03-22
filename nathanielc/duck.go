
package nathanielc

import (
	"github.com/nvcook42/avi"
	"github.com/nvcook42/avi/logger"
)

func init() {
	avi.RegisterShip("duck", NewDuck)
}

type DuckSpaceShip struct {
	avi.GenericShip
}

func NewDuck() avi.Ship {
	return &DuckSpaceShip{}
}


func (self *DuckSpaceShip) Tick() {

	for _, engine := range self.Engines {
		err := engine.PowerOn(1.0)
		if err != nil {
			logger.Debug.Println("Failed to power engines", err)
		}
	}
	scan, err := self.Sensors[0].Scan()
	if err != nil {
		logger.Debug.Println("Failed to scan", err)
		return
	}

	logger.Debug.Println("Duck health", scan.Health)
}

