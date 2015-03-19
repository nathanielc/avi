package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Weapon struct {
	partT
	energy float64
}

// Conf format for loading weapons from a file
type WeaponConf struct {
	Mass float64
	Radius float64
	Energy float64
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

func NewWeaponFromConf(pos mgl64.Vec3, conf WeaponConf) *Weapon {
	return &Weapon{
		partT: partT{
			objectT: objectT{
				position: pos,
				mass:     conf.Mass,
				radius:   conf.Radius,
			},
		},
		energy: conf.Energy,
	}
}

