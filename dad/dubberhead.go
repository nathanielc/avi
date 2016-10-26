package nathanielc

import (
	"math"
	"math/rand"

	"azul3d.org/engine/lmath"

	"github.com/golang/glog"
	"github.com/nathanielc/avi"
	"github.com/nathanielc/avi/nav"
)

func init() {
	avi.RegisterPilot("DubberHead", NewDubberHead)
}

var seed int64 = 0

type DubberHeadPilot struct {
	avi.GenericPilot
	dir           lmath.Vec3
	fired         bool
	navComputer   *nav.Nav
	cooldownTicks int64
	target        avi.ID
	targetI       velPoint
	targetF       velPoint
	ctlp          avi.ID
	ctlpBias      lmath.Vec3
	ctlpBiasRand  *rand.Rand
}

type velPoint struct {
	velocity lmath.Vec3
	tick     int64
}

func NewDubberHead() avi.Pilot {
	seed++
	ctlpBiasRand := rand.New(rand.NewSource(seed))
	n, _ := (lmath.Vec3{ctlpBiasRand.Float64(), ctlpBiasRand.Float64(), ctlpBiasRand.Float64()}).Normalized()
	return &DubberHeadPilot{
		dir:           lmath.Vec3{1, 1, 1},
		cooldownTicks: 1,
		target:        avi.NilID,
		ctlp:          avi.NilID,
		ctlpBias:      n,
		ctlpBiasRand:  ctlpBiasRand,
	}
}

func (self *DubberHeadPilot) Tick(tick int64) {
	if self.navComputer == nil {
		self.navComputer = nav.NewNav(self.Thrusters)
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
	if err != nil {
		if glog.V(4) {
			glog.Infoln("Failed to navigate", err)
		}
	}
	if glog.V(4) {
		glog.Infoln("DubberHead", scan.Health, scan.Position, scan.Velocity.Length())
	}
	self.navCtlP(tick, scan)

	self.fire(tick, scan)

}

func (self *DubberHeadPilot) navCtlP(time int64, scan avi.ScanResult) {

	// Find Control Point
	if !ctlpExists(self.ctlp, scan.ControlPoints) {
		points := 0.0
		for id, ctlp := range scan.ControlPoints {
			p := ctlp.Points
			if p > points {
				points = p
				self.ctlp = id
			}
		}
	}

	if !ctlpExists(self.ctlp, scan.ControlPoints) {
		return
	}

	ctlp := scan.ControlPoints[self.ctlp]

	distance := ctlp.Position.Sub(scan.Position).Length()
	tolerance := (distance - ctlp.Influence) / 2.0
	if time%10 == 0 {
		tolerance /= 100
	}

	if tolerance < 10 {
		tolerance = 30
	}
	wp := nav.Waypoint{
		Position:  ctlp.Position.Add(self.ctlpBias.MulScalar(ctlp.Radius + scan.Radius + tolerance)),
		MaxSpeed:  tolerance * 0.4,
		Tolerance: tolerance,
	}
	self.navComputer.SetWaypoint(wp)
}

func (self *DubberHeadPilot) fire(tick int64, scan avi.ScanResult) {

	//Find target ship
	if !shipExists(self.target, scan.Ships) {
		distance := 0.0
		for id, ship := range scan.Ships {
			if ship.Fleet == self.Fleet {
				continue
			}
			d := ship.Position.Sub(scan.Position).Length()
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

	if targetPos.Sub(scan.Position).LengthSq() > 1e6 {
		self.target = avi.NilID
		if glog.V(3) {
			glog.Infoln("Target is too far away choosing another target")
		}
	}

	if tick%self.cooldownTicks == 0 {
		self.targetF.tick = tick
		self.targetF.velocity = targetVel
		for _, weapon := range self.Weapons {
			vel := weapon.GetAmmoVel()

			acc := self.targetF.velocity.
				Sub(self.targetI.velocity).
				MulScalar(1.0 / (avi.SecondsPerTick * float64(self.targetF.tick-self.targetI.tick)))

			if glog.V(4) {
				glog.Infoln("Acc: ", acc)
			}

			deltaPos := targetPos.Sub(scan.Position)
			deltaVel := targetVel.Sub(scan.Velocity)
			time := calcT(deltaPos, deltaVel, vel)
			if time < 0 {
				if glog.V(3) {
					glog.Infoln("Target out of range", time)
				}
				continue
			}

			dir := deltaVel.Add(deltaPos.MulScalar(1 / time)).Add(acc.MulScalar(time * 0.5))

			if glog.V(3) {
				glog.Infoln(dir, dir.Length())
			}

			err := weapon.Fire(dir)
			if err != nil {
				if glog.V(3) {
					glog.Infoln("Failed to fire", err)
				}
			}
			self.cooldownTicks = weapon.GetCoolDownTicks()
		}
		self.targetI.tick = tick
		self.targetI.velocity = targetVel
	}
}

func calcT(deltaPos, deltaVel lmath.Vec3, va float64) float64 {

	vt := deltaVel.Length()
	x := deltaPos.Length()
	nVel, _ := deltaVel.Normalized()
	nPos, _ := deltaPos.Normalized()
	ctheta := nPos.Dot(nVel)

	a := va*va - vt*vt
	b := 2 * x * vt * ctheta
	c := -x * x

	det := b*b - 4*a*c

	if det < 0 {
		return -1
	}

	t1 := (-b + math.Sqrt(det)) / (2 * a)
	t2 := (-b - math.Sqrt(det)) / (2 * a)

	if glog.V(4) {
		glog.Infoln(deltaPos, deltaVel, va, vt, x, t1, t2)
	}

	if t1 < t2 && t1 > 0 || t2 < 0 {
		return t1
	}

	return t2
}

func ctlpExists(target avi.ID, ctlps map[avi.ID]avi.CtlPSR) bool {
	_, ok := ctlps[target]
	return ok
}

func shipExists(target avi.ID, ships map[avi.ID]avi.ShipSR) bool {
	_, ok := ships[target]
	return ok
}
