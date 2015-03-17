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
	orders := make(chan Order)
	results := make(chan OrderResult)
	self.ships = append(self.ships,
		&shipT{
			ship:    ship,
			orders:  orders,
			results: results,
			parts:   ship.GetParts(),
		})
}

func (self *Simulation) Start() {
	log.Println("Starting AVI Simulation")

	for _, shipChan := range self.ships {
		shipChan.ship.Launch(shipChan.orders, shipChan.results)
	}

	self.loop()
}

func (self *Simulation) loop() {

	ticker := time.Tick(time.Microsecond)
	for {
		_ = <-ticker
		for _, ship := range self.ships {
			ship.Energize()
			select {
			case order := <-ship.orders:
				self.processOrder(ship, order)
			default:
				log.Println("No orders from ship", ship.ship, self.tick)
			}
		}
		self.tick++
	}

}

func (self *Simulation) processOrder(ship *shipT, order Order) {

	result := OrderResult{
		Actions: make([]bool, len(order.Actions)),
	}
	for i, action := range order.Actions {
		part := ship.parts[action.PartID]
		part.HandleAction(action, ship)
		result.Actions[i] = false
	}

	ship.results <- result

}
