package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

const (
	POWER_ON = Opertation(iota)
	POWER_OFF
)

type POWER_ON_Args struct {
	Power float64
}

type engine struct {
	partT
	Energy        float64
	currentOutput float64
}

func NewEngine001(pos mgl64.Vec3) engine {
	return engine{
		partT: partT{
			Position: pos,
			Mass:     2000,
		},
		Energy: 100,
	}
}

func (self *engine) GetOutput() float64 {
	return self.currentOutput
}

func (self *engine) HandleAction(action Action, ship *shipT) {
	switch action.Opertation {
	case POWER_ON:
		args := action.Args.(POWER_ON_Args)
		self.currentOutput = self.Energy * args.Power
	case POWER_OFF:
		self.currentOutput = 0
	}
}
