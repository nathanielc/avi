package avi

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/nvcook42/avi/logger"
)

const ammoRadius = 0.05

type Weapon struct {
	partT
	energy       float64
	ammoVelocity float64
	ammoMass     float64
	ammoRadius   float64
}

// Conf format for loading weapons from a file
type WeaponConf struct {
	Mass         float64
	Radius       float64
	Energy       float64
	AmmoVelocity float64 `yaml:"ammo_velocity"`
	AmmoMass     float64 `yaml:"ammo_mass"`
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
		energy:       5,
		ammoVelocity: 1000,
		ammoMass:     1,
		ammoRadius:   ammoRadius,
	}
}

func NewWeaponFromConf(pos mgl64.Vec3, conf WeaponConf) *Weapon {
	logger.Debug.Println(conf)
	return &Weapon{
		partT: partT{
			objectT: objectT{
				position: pos,
				mass:     conf.Mass,
				radius:   conf.Radius,
			},
		},
		energy:       conf.Energy,
		ammoVelocity: conf.AmmoVelocity,
		ammoMass:     conf.AmmoMass,
		ammoRadius:   ammoRadius,
	}
}

func (self *Weapon) Fire(dir mgl64.Vec3) error {

	err := self.ship.ConsumeEnergy(self.energy)
	if err != nil {
		return err
	}
	force := self.ammoMass * self.ammoVelocity
	self.ship.ApplyThrust(dir.Mul(-1.0), force)

	norm := dir.Normalize()
	pos := norm.Mul(self.ship.radius + 1).Add(self.ship.position)
	vel := norm.Mul(self.ammoVelocity)

	self.ship.sim.addProjectile(pos, vel, self.ammoMass, self.ammoRadius)

	return nil
}
