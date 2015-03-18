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
			Position: pos,
			Mass:     1000,
		},
		force:  100,
		energy: 10,
	}
}

func (self *Thruster) Thrust(dir mgl64.Vec3, power float64) <-chan error {

	ch := make(chan error)
	self.currentOrder = func (ship *shipT) {
		defer close(ch)
		log.Println("thrust", ch)
		force := self.force * power
		energy := self.energy * power
		err := ship.ConsumeEnergy(energy)
		if err != nil {
			log.Println("thrust err", ch)
			ch <- err
			return
		}
		ship.ApplyThrust(dir, force)
		ch <- nil
	}
	return ch
}
