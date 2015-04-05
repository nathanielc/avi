package avi

import (
	"fmt"
	"errors"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

const ammoRadius = 0.05

type Weapon struct {
	partT
	energy        float64
	ammoVelocity  float64
	ammoMass      float64
	ammoRadius    float64
	cooldownTicks int64
	lastshot      int64
}

// Conf format for loading weapons from a file
type WeaponConf struct {
	Mass         float64
	Radius       float64
	Energy       float64
	AmmoVelocity float64 `yaml:"ammo_velocity"`
	AmmoMass     float64 `yaml:"ammo_mass"`
	Cooldown     float64
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
		energy:        5,
		ammoVelocity:  1000,
		ammoMass:      1,
		ammoRadius:    ammoRadius,
		cooldownTicks: int64(5.0 / TimePerTick),
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
		energy:        conf.Energy,
		ammoVelocity:  conf.AmmoVelocity,
		ammoMass:      conf.AmmoMass,
		ammoRadius:    ammoRadius,
		cooldownTicks: int64(conf.Cooldown / TimePerTick),
	}
}

func (self *Weapon) Fire(dir mgl64.Vec3) error {
	if l := dir.Len(); math.IsNaN(l) || l == 0 {
		err := errors.New(fmt.Sprintf("Invalid direction %s", dir))
		return err
	}

	if self.lastshot+self.cooldownTicks > self.ship.sim.tick {
		return errors.New("Weapon cooling down")
	}
	self.lastshot = self.ship.sim.tick

	err := self.ship.ConsumeEnergy(self.energy)
	if err != nil {
		return err
	}
	force := self.ammoMass * self.ammoVelocity
	self.ship.ApplyThrust(dir.Mul(-1.0), force)

	norm := dir.Normalize()
	pos := norm.Mul(self.ship.radius + 1).Add(self.ship.position)
	vel := norm.Mul(self.ammoVelocity).Add(self.ship.velocity)

	self.ship.sim.addProjectile(pos, vel, self.ammoMass, self.ammoRadius)

	return nil
}

func (self *Weapon) GetCoolDownTicks() int64 {
	return self.cooldownTicks
}

func (self *Weapon) GetAmmoVel() float64 {
	return self.ammoVelocity
}
