package jac

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	"github.com/nathanielc/avi"
	"github.com/nathanielc/avi/nav"
)

func init() {
	avi.RegisterPilot("jill", NewJill)
}

type JillPilot struct {
	avi.GenericPilot
	dir           mgl64.Vec3
	fired         bool
	navComputer   *nav.Nav
	cooldownTicks int64
	target        int64
}

func NewJill() avi.Pilot {
	return &JillPilot{
		dir:           mgl64.Vec3{1, 1, 1},
		cooldownTicks: 1,
		target:        -1,
	}
}

func (self *JillPilot) Tick(tick int64) {
	if self.navComputer == nil {
		self.navComputer = nav.NewNav(self.Thrusters)

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
	}
	glog.V(3).Infoln("jill", scan.Health, scan.Position, scan.Velocity.Len(), len(scan.Ships))
	if !targetExists(self.target, scan.Ships) {
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
	if !targetExists(self.target, scan.Ships) {
		return
	}
	targetPos := scan.Ships[self.target].Position
	targetVel := scan.Ships[self.target].Velocity

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

func targetExists(target int64, ships map[int64]avi.ShipSR) bool {
	_, ok := ships[target]
	return ok
}
