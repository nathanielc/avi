package nathanielc

import (
	"github.com/golang/glog"
	"github.com/nathanielc/avi"
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
			if glog.V(3) {
				glog.Infoln("Failed to power engines", err)
			}
		}
	}
	scan, err := self.Sensors[0].Scan()
	if err != nil {
		if glog.V(3) {
			glog.Infoln("Failed to scan", err)
		}
		return
	}

	if glog.V(3) {
		glog.Infoln("Duck health", scan.Health)
	}
}
