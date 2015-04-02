package nathanielc

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	"github.com/nvcook42/avi"
	"github.com/nvcook42/avi/nav"
	"math"
)

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
	targetI       velPoint
	targetF       velPoint
	ctlp          int64
}

type velPoint struct {
	velocity mgl64.Vec3
	tick     int64
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
	}
	for _, engine := range self.Engines {
		err := engine.PowerOn(1.0)
		if err != nil {
			glog.V(4).Infoln("Failed to power engines", err)
		}
	}
	scan, err := self.Sensors[0].Scan()
	if err != nil {
		glog.V(4).Infoln("Failed to scan", err)
		return
	}
	err = self.navComputer.Tick(scan.Position, scan.Velocity)
	if err != nil {
		glog.V(4).Infoln("Failed to navigate", err)
	}
	glog.V(4).Infoln("Jim", scan.Health, scan.Position, scan.Velocity.Len())
	self.navCtlP(scan)

	self.fire(tick, scan)

}

func (self *JimPilot) navCtlP(scan *avi.ScanResult) {

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
}

func (self *JimPilot) fire(tick int64, scan *avi.ScanResult) {

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
				self.targetI = velPoint{}
			}
		}
	}
	if !shipExists(self.target, scan.Ships) {
		return
	}

	target := scan.Ships[self.target]
	targetPos := target.Position
	targetVel := target.Velocity

	if tick%self.cooldownTicks == 0 {
		self.targetF.tick = tick
		self.targetF.velocity = targetVel
		for _, weapon := range self.Weapons {
			vel := weapon.GetAmmoVel()

			acc := self.targetF.velocity.
				Sub(self.targetI.velocity).
				Mul(1.0 / (avi.TimePerTick * float64(self.targetF.tick-self.targetI.tick)))

			glog.V(4).Infoln("Acc: ", acc)

			deltaPos := targetPos.Sub(scan.Position)
			deltaVel := targetVel.Sub(scan.Velocity)
			time := calcT(deltaPos, deltaVel, vel)
			if time < 0 {
				glog.V(3).Infoln("Target out of range", time)
				continue
			}

			dir := deltaVel.Add(deltaPos.Mul(1/time)).Add(acc.Mul(time*0.5))

			glog.V(3).Infoln(dir, dir.Len())

			err := weapon.Fire(dir)
			if err != nil {
				glog.V(3).Infoln("Failed to fire", err)
			}
			self.cooldownTicks = weapon.GetCoolDownTicks()
			self.targetI = velPoint{}
		}
	} else {
		//Track target accerlation
		if self.targetI.tick == 0 {
			self.targetI.tick = tick
			self.targetI.velocity = targetVel
		}
	}
}

func calcT(deltaPos, deltaVel mgl64.Vec3, va float64) float64 {

	vt := deltaVel.Len()
	x := deltaPos.Len()


	ctheta := deltaPos.Normalize().Dot(deltaVel.Normalize())

	a := va*va - vt*vt
	b := 2*x*vt*ctheta
	c := -x*x


	det := b*b-4*a*c

	if det < 0 {
		 return -1
	}

	t1 := (-b + math.Sqrt(det)) / (2*a)
	t2 := (-b - math.Sqrt(det)) / (2*a)

	glog.V(4).Infoln(deltaPos, deltaVel, va, vt, x, t1, t2)

	if t1 < t2 && t1 > 0  || t2 < 0{
		return t1
	}

	return t2
}

func ctlpExists(target int64, ctlps map[int64]avi.CtlPSR) bool {
	_, ok := ctlps[target]
	return ok
}

func shipExists(target int64, ships map[int64]avi.ShipSR) bool {
	_, ok := ships[target]
	return ok
}
