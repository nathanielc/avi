package nathanielc

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/nvcook42/avi"
	"log"
)

type JimSpaceShip struct {
	orders  chan avi.Order
	results chan avi.OrderResult
}

func NewJim() avi.Ship {
	return &JimSpaceShip{}
}

func (self *JimSpaceShip) Launch(orders chan avi.Order, results chan avi.OrderResult) {
	self.orders = orders
	self.results = results

	go self.think()
}

func (self *JimSpaceShip) think() {
	moveOrder := avi.Order{
		Actions: []avi.Action{
			avi.Action{
				PartID:     1,
				Opertation: avi.THRUST,
				Args: avi.THRUSTArgs{
					Direction: mgl64.Vec3{1, 1, 1},
					Power:     0.50,
				},
			},
		},
	}
	for {
		log.Println("Sending orders")
		self.orders <- moveOrder
		results := <-self.results
		for id, success := range results.Actions {
			if !success {
				log.Println("Failed action", id)
			}
		}
	}
}

func (self *JimSpaceShip) GetParts() []avi.Part {
	thruster0 := avi.NewThruster001(mgl64.Vec3{0, -1, 0})
	thruster1 := avi.NewThruster001(mgl64.Vec3{0, 1, 0})
	return []avi.Part{
		thruster0,
		thruster1,
	}
}
