package avi

import (
	"log"
	"time"
)

type Simulation struct {
	ships []*shipT
	tick  int64
}

func NewSimulation() *Simulation {
	return &Simulation{}
}

func (self *Simulation) AddShip(ship Ship) {
	self.ships = append(self.ships, newShip(ship))
}

func (self *Simulation) Start() {
	log.Println("Starting AVI Simulation")

	for _, shipChan := range self.ships {
		shipChan.ship.Launch()
	}

	self.loop()
}

func (self *Simulation) loop() {

	ticker := time.Tick(time.Microsecond)
	for {
		_ = <-ticker
		log.Println("tick:", self.tick)
		for _, ship := range self.ships {
			ship.Energize()
			ship.ProcessOrders()
		}
		self.tick++
	}

}

