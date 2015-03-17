package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

const (
	FIRE = Opertation(iota)
)

type FIRE_Args struct {
	Direction mgl64.Vec3
	Power     float64
}

type weapon struct {
	partT
	Energy float64
}

func NewWeapon001(pos mgl64.Vec3) weapon {
	return weapon{
		partT: partT{
			Position: pos,
			Mass:     100,
		},
		Energy: 5,
	}
}

func (self *weapon) HandleAction(action Action, ship *shipT) {
	switch action.Opertation {
	case FIRE:
		ship.ConsumeEngery(self.energy)
	}
}
