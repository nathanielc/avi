package nathanielc

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	"github.com/nvcook42/avi"
	"github.com/nvcook42/avi/nav"
)

var pattern = []*nav.Waypoint{
	&nav.Waypoint{
		Position:  mgl64.Vec3{-100, 0, -100},
		MaxSpeed:  10,
		Tolerance: 30,
	},
	&nav.Waypoint{
		Position:  mgl64.Vec3{-100, 100, -100},
		MaxSpeed:  10,
		Tolerance: 30,
	},
	&nav.Waypoint{
		Position:  mgl64.Vec3{-100, 100, 0},
		MaxSpeed:  10,
		Tolerance: 30,
	},
	&nav.Waypoint{
		Position:  mgl64.Vec3{-100, 100, 100},
		MaxSpeed:  10,
		Tolerance: 30,
	},
	&nav.Waypoint{
		Position:  mgl64.Vec3{-100, 0, 100},
		MaxSpeed:  10,
		Tolerance: 30,
	},
	&nav.Waypoint{
		Position:  mgl64.Vec3{-100, -100, 100},
		MaxSpeed:  10,
		Tolerance: 30,
	},
	&nav.Waypoint{
		Position:  mgl64.Vec3{-100, -100, 0},
		MaxSpeed:  10,
		Tolerance: 30,
	},
	&nav.Waypoint{
		Position:  mgl64.Vec3{-100, -100, -100},
		MaxSpeed:  10,
		Tolerance: 30,
	},
	&nav.Waypoint{
		Position:  mgl64.Vec3{0, 0, 0},
		MaxSpeed:  50,
		Tolerance: 10,
	},
}

func init() {
	avi.RegisterPilot("jim", NewJim)
}

type JimPilot struct {
	avi.GenericPilot
	dir           mgl64.Vec3
	fired         bool
	navComputer   *nav.Nav
	cooldownTicks int64
	target        int64
	ctlp          int64
}

func NewJim() avi.Pilot {
	return &JimPilot{
		dir:           mgl64.Vec3{1, 1, 1},
		cooldownTicks: 1,
		target:        -1,
	}
}

func (self *JimPilot) Tick(tick int64) {
	if self.navComputer == nil {
		self.navComputer = nav.NewNav(self.Thrusters)
		for _, wp := range pattern {
			glog.V(3).Infoln("Adding wp", wp)
			self.navComputer.AddWaypoint(wp)
		}
	}
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
	err = self.navComputer.Tick(scan.Position, scan.Velocity)
	if err != nil {
		glog.V(3).Infoln("Failed to navigate", err)
		return
	}
	glog.V(3).Infoln("Jim", scan.Health, scan.Position, scan.Velocity.Len())
	// Find Control Point
	if !ctlpExists(self.ctlp, scan.ControlPoints) {
		distance := 0.0
		for id, ctlp := range scan.ControlPoints {
			d := ctlp.Position.Sub(scan.Position).Len()
			if d < distance || distance == 0 {
				distance = d
				self.ctlp = id
			}
		}
	}

	if !ctlpExists(self.ctlp, scan.ControlPoints) {
		return
	}

	ctlp := scan.ControlPoints[self.ctlp]

	distance := ctlp.Position.Sub(scan.Position).Len()
	tolerance := (distance - ctlp.Influence) / 2.0
	if tolerance < 10 {
		tolerance = 30
	}
	bias := mgl64.Vec3{0, 0, 1}
	wp := &nav.Waypoint{
		Position:  ctlp.Position.Add(bias.Mul(ctlp.Radius + scan.Radius + tolerance)),
		MaxSpeed:  tolerance * 0.4,
		Tolerance: tolerance,
	}
	self.navComputer.SetWaypoint(wp)

	//Find target ship
	if !shipExists(self.target, scan.Ships) {
		distance := 0.0
		for id, ship := range scan.Ships {
			if ship.Fleet == self.Fleet {
				continue
			}
			d := ship.Position.Sub(scan.Position).Len()
			if d < distance || distance == 0 {
				distance = d
				self.target = id
			}
		}
	}
	if !shipExists(self.target, scan.Ships) {
		return
	}

	targetPos := scan.Ships[self.target].Position
	targetVel := scan.Ships[self.target].Velocity

	if tick%self.cooldownTicks == 0 {
		for _, weapon := range self.Weapons {
			vel := weapon.GetAmmoVel()
			time := scan.Position.Sub(targetPos).Len() / vel

			dir := targetPos.Add(targetVel.Mul(time*1.1)).Sub(scan.Position).Sub(scan.Velocity.Mul(time))

			err := weapon.Fire(dir)
			if err != nil {
				glog.V(3).Infoln("Failed to fire", err)
			}
			self.cooldownTicks = weapon.GetCoolDownTicks()
		}
	}
}

func ctlpExists(target int64, ctlps map[int64]avi.CtlPSR) bool {
	_, ok := ctlps[target]
	return ok
}

func shipExists(target int64, ships map[int64]avi.ShipSR) bool {
	_, ok := ships[target]
	return ok
}
