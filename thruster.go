package avi

import (
	"github.com/go-gl/mathgl/mgl64"
	"log"
)

type Thruster struct {
	partT
	force  float64
	energy float64

	thrustChan chan error
	thrustOrder func(chan error, *shipT)
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

func (self *Thruster) Thrust(dir mgl64.Vec3, power float64) error {

	force := self.force * power
	energy := self.energy * power
	err := self.ship.ConsumeEnergy(energy)
	if err != nil {
		log.Println("thrust err")
		return err
	}
	self.ship.ApplyThrust(dir, force)
	return nil
}
