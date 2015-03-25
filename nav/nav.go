package nav

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	"github.com/nvcook42/avi"
	"math"
)

type Waypoint struct {
	Position  mgl64.Vec3
	MaxSpeed  float64
	Tolerance float64
}

type Nav struct {
	thrusters []*avi.Thruster
	waypoints queue
	next      *Waypoint
}

func NewNav(thrusters []*avi.Thruster) *Nav {
	return &Nav{
		thrusters: thrusters,
		waypoints: queue{nodes: make([]*Waypoint, 5)},
	}
}

func (nav *Nav) AddWaypoint(wp *Waypoint) {
	nav.waypoints.Push(wp)
}

func (nav *Nav) Tick(pos, vel mgl64.Vec3) error {
	if nav.next == nil {
		nav.next = nav.waypoints.Pop()
	}

	glog.V(3).Infoln("Next", nav.next)

	speed := vel.Len()
	//delta := pos.Sub(nav.next.Position)
	delta := nav.next.Position.Sub(pos)
	distance := delta.Len()

	t := nav.next.Tolerance
	t2 := t * t

	if distance < t {
		nav.next = nil
		glog.V(2).Infoln("Hit waypoint", nav.next, distance, t)
		return nil
	}

	hypo := math.Sqrt(distance*distance + t2)
	toleranceAngle := math.Asin(t / hypo)

	realAngle := math.Acos(pos.Dot(delta) / (distance * pos.Len()))

	if realAngle > toleranceAngle {
		nav.thrust(delta)
	} else if speed < nav.next.MaxSpeed {
		nav.thrust(delta)
	}

	return nil
}

func (nav *Nav) thrust(dir mgl64.Vec3) error {
	glog.V(3).Infoln("Thrusting", dir.Normalize())
	for _, thruster := range nav.thrusters {
		err := thruster.Thrust(dir, 1.0)
		if err != nil {
			return err
		}
	}
	return nil
}
