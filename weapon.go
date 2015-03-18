package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Weapon struct {
	partT
	energy float64
}

func NewWeapon001(pos mgl64.Vec3) *Weapon {
	return &Weapon{
		partT: partT{
			objectT: objectT{
				position: pos,
				mass:     1000,
				radius:   1,
			},
		},
		energy: 5,
	}
}

