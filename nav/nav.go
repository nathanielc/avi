package nav

import (
	"errors"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	"github.com/nathanielc/avi"
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
	set       bool
	next      Waypoint
}

func NewNav(thrusters []*avi.Thruster) *Nav {
	return &Nav{
		thrusters: thrusters,
		waypoints: queue{nodes: make([]Waypoint, 5)},
	}
}

func (nav *Nav) SetWaypoint(wp Waypoint) {
	nav.set = true
	nav.next = wp
}

func (nav *Nav) ClearWaypoints() {
	nav.set = false
	nav.waypoints.Clear()
}

func (nav *Nav) AddWaypoint(wp Waypoint) {
	nav.waypoints.Push(wp)
}

func (nav *Nav) Tick(pos, vel mgl64.Vec3) error {
	if !nav.set {
		var ok bool
		nav.next, ok = nav.waypoints.Pop()
		if !ok {
			return NoMoreWaypoints
		}
	}

	if glog.V(3) {
		glog.Infoln("Next", nav.next)
	}

	delta := nav.next.Position.Sub(pos)
	distance := delta.Len()

	t := nav.next.Tolerance

	if distance < t {
		if glog.V(2) {
			glog.Infoln("Hit waypoint", nav.next, distance, t)
		}
		nav.set = false
		return nil
	}

	desiredVel := delta.Normalize().Mul(nav.next.MaxSpeed)

	accerlation := desiredVel.Sub(vel)

	return nav.thrust(accerlation)
}

func (nav *Nav) thrust(acc mgl64.Vec3) error {
	if glog.V(3) {
		glog.Infoln("Thrusting", acc)
	}
	scaled := acc.Mul(1.0 / float64(len(nav.thrusters)))
	for _, thruster := range nav.thrusters {
		err := thruster.Thrust(scaled)
		if err != nil {
			return err
		}
	}
	return nil
}
