package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

const (
	SIMPLE_SCAN = Opertation(iota)
)

type sensor struct {
	partT
	Energy float64
}

func NewSensor001(pos mgl64.Vec3) sensor {
	return sensor{
		partT: partT{
			Position: pos,
			Mass:     10,
		},
		Energy: 1,
	}
}

func (self *sensor) HandleAction(action Action, ship *shipT) {
	switch action.Opertation {
	case SIMPLE_SCAN:
		ship.ConsumeEngery(self.energy)
	}
}
