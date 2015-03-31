package nathanielc

import (
	"github.com/golang/glog"
	"github.com/nvcook42/avi"
)

func init() {
	avi.RegisterPilot("duck", NewDuck)
}

type DuckPilot struct {
	avi.GenericPilot
}

func NewDuck() avi.Pilot {
	return &DuckPilot{}
}

func (self *DuckPilot) Tick(tick int64) {

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
