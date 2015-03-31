package nav

import (
	"errors"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	"github.com/nvcook42/avi"
)

var NoMoreWaypoints = errors.New("No more waypoints")

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

func (nav *Nav) SetWaypoint(wp *Waypoint) {
	nav.next = wp
}

func (nav *Nav) AddWaypoint(wp *Waypoint) {
	nav.waypoints.Push(wp)
}

func (nav *Nav) Tick(pos, vel mgl64.Vec3) error {
	if nav.next == nil {
		nav.next = nav.waypoints.Pop()
		if nav.next == nil {
			return NoMoreWaypoints
		}
	}

	glog.V(3).Infoln("Next", nav.next)

	delta := nav.next.Position.Sub(pos)
	distance := delta.Len()

	t := nav.next.Tolerance

	if distance < t {
		glog.V(2).Infoln("Hit waypoint", nav.next, distance, t)
		nav.next = nil
		return nil
	}

	desiredVel := delta.Normalize().Mul(nav.next.MaxSpeed)

	accerlation := desiredVel.Sub(vel)

	return nav.thrust(accerlation)
}

func (nav *Nav) thrust(acc mgl64.Vec3) error {
	glog.V(3).Infoln("Thrusting", acc)
	for _, thruster := range nav.thrusters {
		err := thruster.Thrust(acc)
		if err != nil {
			return err
		}
	}
	return nil
}
