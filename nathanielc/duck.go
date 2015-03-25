package nathanielc

import (
	"github.com/golang/glog"
	"github.com/nvcook42/avi"
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
			glog.V(3).Infoln("Failed to power engines", err)
		}
	}
	scan, err := self.Sensors[0].Scan()
	if err != nil {
		glog.V(3).Infoln("Failed to scan", err)
		return
	}

	glog.V(3).Infoln("Duck health", scan.Health)
}
