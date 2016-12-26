package avi

import (
	"errors"
	"fmt"
	"math"

	"azul3d.org/engine/lmath"
)

var OutOfAmmoError = errors.New("Out of ammunition")

type Weapon struct {
	partT
	energy        float64
	ammoVelocity  float64
	ammoMass      float64
	ammoRadius    float64
	ammoCapacity  int64
	cooldownTicks int64
	lastshot      int64
}

// Conf format for loading weapons from a file
type WeaponConf struct {
	Mass         float64 `yaml:"mass" json:"mass"`
	Radius       float64 `yaml:"radius" json:"radius"`
	Energy       float64 `yaml:"energy" json:"energy"`
	AmmoVelocity float64 `yaml:"ammo_velocity" json:"ammo_velocity"`
	AmmoMass     float64 `yaml:"ammo_mass" json:"ammo_mass"`
	AmmoCapacity int64   `yaml:"ammo_capacity" json:"ammo_capacity"`
	AmmoRadius   float64 `yaml:"ammo_radius" json:"ammo_radius"`
	Cooldown     float64 `yaml:"cooldown" json:"cooldown"`
}

func NewWeapon001(pos lmath.Vec3) *Weapon {
	return &Weapon{
		partT: partT{
			objectT: objectT{
				position: pos,
				mass:     1e6,
				radius:   1,
			},
		},
		energy:        5,
		ammoVelocity:  1000,
		ammoMass:      1,
		ammoRadius:    0.05,
		ammoCapacity:  1e5,
		cooldownTicks: int64(5.0 / SecondsPerTick),
	}
}

func NewWeaponFromConf(pos lmath.Vec3, conf WeaponConf) *Weapon {
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
		ammoRadius:    conf.AmmoRadius,
		ammoCapacity:  conf.AmmoCapacity,
		cooldownTicks: int64(conf.Cooldown / SecondsPerTick),
	}
}

func (self *Weapon) Mass() float64 {
	return self.mass + float64(self.ammoCapacity)*self.ammoMass
}

func (self *Weapon) Fire(dir lmath.Vec3) error {
	if l := dir.LengthSq(); math.IsNaN(l) || l == 0 {
		err := errors.New(fmt.Sprintf("Invalid direction %s", dir))
		return err
	}

	if self.ammoCapacity <= 0 {
		return OutOfAmmoError
	}

	self.ammoCapacity--
	self.ship.mass -= self.ammoMass

	if self.lastshot+self.cooldownTicks > self.ship.sim.tick {
		return errors.New("Weapon cooling down")
	}
	self.lastshot = self.ship.sim.tick

	err := self.ship.ConsumeEnergy(self.energy)
	if err != nil {
		return err
	}
	force := self.ammoMass * self.ammoVelocity
	self.ship.ApplyThrust(dir.MulScalar(-1.0), force)

	norm, _ := dir.Normalized()
	pos := norm.MulScalar(self.ship.radius + 1).Add(self.ship.position)
	vel := norm.MulScalar(self.ammoVelocity).Add(self.ship.velocity)

	self.ship.sim.addProjectile(pos, vel, self.ammoMass, self.ammoRadius)

	return nil
}

func (self *Weapon) GetCoolDownTicks() int64 {
	return self.cooldownTicks
}

func (self *Weapon) GetAmmoVel() float64 {
	return self.ammoVelocity
}
