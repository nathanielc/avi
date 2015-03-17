package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

const (
	THRUST = Opertation(iota)
)

type THRUST_Args struct {
	Direction mgl64.Vec3
	Power     float64
}

type Thruster struct {
	partT
	force  float64
	energy float64
}

func NewThruster001(pos mgl64.Vec3) Thruster {
	return Thruster{
		partT: partT{
			Position: pos,
			Mass:     1000,
		},
		force:  100,
		energy: 10,
	}
}

func (self *Thruster) HandleAction(action Action, ship *shipT) {
	switch action.Opertation {
	case THRUST:
		args := action.Args.(THRUST_Args)
		force := self.force * args.Power
		energy := self.energy * args.Power
		ship.ConsumeEnergy(energy)
		ship.ApplyThrust(args.Direction, force)

	}
}
