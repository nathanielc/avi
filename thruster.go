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

type thruster struct {
	partT
	Force  float64
	Energy float64
}

func NewThruster001(pos mgl64.Vec3) thruster {
	return thruster{
		partT: partT{
			Position: pos,
			Mass:     1000,
		},
		Force:  100,
		Energy: 10,
	}
}

func (self thruster) HandleAction(action Action, ship *shipT) {
	switch action.Opertation {
	case THRUST:
		args := action.Args.(THRUST_Args)
		log.Println("Pos:", self.Position)
		force := self.Force * args.Power
		energy := self.Energy * args.Power
		ship.ConsumeEngery(energy)
		ship.ApplyThrust(args.Direction, force)

	}
}
