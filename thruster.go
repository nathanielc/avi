package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Thruster struct {
	partT
	force  float64
	energy float64
}

// Conf format for loading thrusters from a file
type ThrusterConf struct {
	Mass float64
	Radius float64
	Force float64
	Energy float64
}

func NewThruster001(pos mgl64.Vec3) *Thruster {
	return &Thruster{
		partT: partT{
			objectT: objectT{
				position: pos,
				mass:     1500,
				radius:   2,
			},
		},
		force:  100,
		energy: 10,
	}
}

func NewThrusterFromConf(pos mgl64.Vec3, conf ThrusterConf) *Thruster {
	return &Thruster{
		partT: partT{
			objectT: objectT{
				position: pos,
				mass:     conf.Mass,
				radius:   conf.Radius,
			},
		},
		force:  conf.Force,
		energy: conf.Energy,
	}
}

func (self *Thruster) Thrust(dir mgl64.Vec3, power float64) error {

	force := self.force * power
	energy := self.energy * power
	err := self.ship.ConsumeEnergy(energy)
	if err != nil {
		return err
	}
	self.ship.ApplyThrust(dir, force)
	return nil
}
