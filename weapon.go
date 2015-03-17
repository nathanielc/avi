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

type Weapon struct {
	partT
	energy float64
}

func NewWeapon001(pos mgl64.Vec3) Weapon {
	return Weapon{
		partT: partT{
			Position: pos,
			Mass:     100,
		},
		energy: 5,
	}
}

func (self *Weapon) HandleAction(action Action, ship *shipT) {
	switch action.Opertation {
	case FIRE:
		ship.ConsumeEnergy(self.energy)
	}
}
