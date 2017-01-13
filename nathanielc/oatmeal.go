package nathanielc

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	"github.com/nathanielc/avi"
	"github.com/nathanielc/avi/nav"
)

func init() {
	avi.RegisterPilot("oatmeal", NewOatmeal)
}

type OatmealPilot struct {
	avi.GenericPilot
	navComputer *nav.Nav
	ctlp        avi.ID
	targetVel   float64
	maxForce    float64
}

func NewOatmeal() avi.Pilot {
	if glog.V(4) {
		glog.Infoln("New OATMEAL")
	}
	return &OatmealPilot{
		targetVel: 9,
	}
}

func (self *OatmealPilot) Tick(tick int64) {
	if self.navComputer == nil {
		self.navComputer = nav.NewNav(self.Thrusters)
		self.navComputer.AddWaypoint(nav.Waypoint{
			Position:  mgl64.Vec3{0, 800, 100},
			MaxSpeed:  50,
			Tolerance: 10,
		})
		self.navComputer.AddWaypoint(nav.Waypoint{
			Position:  mgl64.Vec3{0, 600, 0},
			MaxSpeed:  40,
			Tolerance: 10,
		})
		self.navComputer.AddWaypoint(nav.Waypoint{
			Position:  mgl64.Vec3{0, 500, 0},
			MaxSpeed:  20,
			Tolerance: 10,
		})
		self.navComputer.AddWaypoint(nav.Waypoint{
			Position:  mgl64.Vec3{0, 100, 0},
			MaxSpeed:  self.targetVel * 2.0,
			Tolerance: 10,
		})
		self.navComputer.AddWaypoint(nav.Waypoint{
			Position:  mgl64.Vec3{0, 0, 0},
			MaxSpeed:  self.targetVel * 2.0,
			Tolerance: 10,
		})
		if glog.V(4) {
			glog.Infoln("Initialized nav")
		}
	}
	if self.maxForce == 0.0 {
		for _, thruster := range self.Thrusters {
			self.maxForce += thruster.GetForce()
		}
	}
	for _, engine := range self.Engines {
		err := engine.PowerOn(1.0)
		if err != nil {
			if glog.V(4) {
				glog.Infoln("Failed to power engines", err)
			}
		}
	}
	scan, err := self.Sensors[0].Scan()
	if err != nil {
		if glog.V(4) {
			glog.Infoln("Failed to scan", err)
		}
		return
	}
	defer scan.Done()

	err = self.navComputer.Tick(scan.Position, scan.Velocity)
	if err == nav.NoMoreWaypoints {
		//Perform orbit manuver
		if glog.V(5) {
			glog.Infoln("Perform orbit manuver")
		}
		self.orbit(scan)
	} else if err != nil {
		if glog.V(4) {
			glog.Infoln("Failed to navigate", err)
		}
	}

	// Pick first control point
	for id := range scan.ControlPoints {
		self.ctlp = id
		break
	}
}

func (self *OatmealPilot) orbit(scan avi.ScanResult) {

	ctlp := scan.ControlPoints[self.ctlp]
	vel := scan.Velocity.Len()
	force := scan.Mass * vel * vel / (ctlp.Radius)

	if force > self.maxForce {
		if glog.V(2) {
			glog.Infof("Not enough thruster force to have stable orbit, max: %f needed: %f", self.maxForce, force)
		}
	}
	n := ctlp.Position.Sub(scan.Position).Normalize()
	accerlation := n.Mul(force / scan.Mass)
	scaled := accerlation.Mul(1.0 / float64(len(self.Thrusters)))
	for _, thruster := range self.Thrusters {
		thruster.Thrust(scaled)
	}
}
