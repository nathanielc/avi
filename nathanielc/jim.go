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
	avi.RegisterShip("jim", NewJim)
}

type JimSpaceShip struct {
	avi.GenericShip
	dir           mgl64.Vec3
	fired         bool
	navComputer   *nav.Nav
	cooldownTicks int64
	target        int64
}

func NewJim() avi.Ship {
	return &JimSpaceShip{
		dir:           mgl64.Vec3{1, 1, 1},
		cooldownTicks: 1,
		target:        -1,
	}
}

func (self *JimSpaceShip) Tick(tick int64) {
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
	glog.V(3).Infoln("Jim", scan.Health, scan.Position, scan.Velocity.Len(), len(scan.Ships))
	if !targetExists(self.target, scan.Ships) {
		distance := 0.0
		for id, ship := range scan.Ships {
			d := ship.GetPosition().Sub(scan.Position).Len()
			if d < distance || distance == 0 {
				distance = d
				self.target = id
			}
		}
	}
	if !targetExists(self.target, scan.Ships) {
		return
	}
	targetPos := scan.Ships[self.target].GetPosition()
	targetVel := scan.Ships[self.target].GetVelocity()

	wp := &nav.Waypoint{
		Position:  targetPos,
		MaxSpeed:  100,
		Tolerance: 150,
	}
	self.navComputer.SetWaypoint(wp)

	if tick%self.cooldownTicks == 0 {
		for _, weapon := range self.Weapons {
			vel := weapon.GetAmmoVel()
			time := scan.Position.Sub(targetPos).Len() / vel

			dir := targetPos.Add(targetVel.Mul(time)).Sub(scan.Position).Sub(scan.Velocity)

			err := weapon.Fire(dir)
			if err != nil {
				glog.V(3).Infoln("Failed to fire", err)
			}
			self.cooldownTicks = weapon.GetCoolDownTicks()
		}
	}
}

func targetExists(target int64, ships map[int64]avi.Object) bool {
	_, ok := ships[target]
	return ok
}
