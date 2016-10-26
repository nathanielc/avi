package avi

import (
	"errors"

	"azul3d.org/engine/lmath"
)

type Thruster struct {
	partT
	force  float64
	energy float64
}

// Conf format for loading thrusters from a file
type ThrusterConf struct {
	Mass   float64 `yaml:"mass" json:"mass"`
	Radius float64 `yaml:"radius" json:"radius"`
	Force  float64 `yaml:"force" json:"force"`
	Energy float64 `yaml:"energy" json:"energy"`
}

func NewThruster001(pos lmath.Vec3) *Thruster {
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

func NewThrusterFromConf(pos lmath.Vec3, conf ThrusterConf) *Thruster {
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

func (self *Thruster) GetForce() float64 {
	return self.force
}

// Fire the thruster the length of dir indicates how hard
// to fire the thruster. The length should equal to the
// accerlation to apply to the ship.
func (self *Thruster) Thrust(dir lmath.Vec3) error {
	if self.used {
		return errors.New("Already used thruster this tick")
	}
	self.used = true

	force := self.ship.mass * dir.Length()
	if force > self.force {
		force = self.force
	}
	energy := self.energy * force / self.force
	err := self.ship.ConsumeEnergy(energy)
	if err != nil {
		return err
	}
	n, _ := dir.Normalized()
	self.ship.ApplyAcc(n.MulScalar(force / self.ship.mass))
	return nil
}
