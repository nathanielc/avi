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

type Engine struct {
	partT
	energy        float64
	currentOutput float64
}

func NewEngine001(pos mgl64.Vec3) Engine {
	return Engine{
		partT: partT{
			Position: pos,
			Mass:     2000,
		},
		energy: 100,
	}
}

func (self *Engine) GetOutput() float64 {
	return self.currentOutput
}

func (self *Engine) HandleAction(action Action, ship *shipT) {
	switch action.Opertation {
	case POWER_ON:
		args := action.Args.(POWER_ON_Args)
		self.currentOutput = self.energy * args.Power
	case POWER_OFF:
		self.currentOutput = 0
	}
}
