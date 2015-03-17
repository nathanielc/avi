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
				PartID:     "t0",
				Opertation: avi.THRUST,
				Args: avi.THRUST_Args{
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

func (self *JimSpaceShip) GetParts() map[string]avi.Part{
	engine := avi.NewEngine001(mgl64.Vec3{0, 0, 1})
	thruster0 := avi.NewThruster001(mgl64.Vec3{0, -1, 0})
	thruster1 := avi.NewThruster001(mgl64.Vec3{0, 1, 0})
	weapon := avi.NewWeapon001(mgl64.Vec3{0, 0, 1})
	sensor := avi.NewSensor001(mgl64.Vec3{0, 0, 1})
	return map[string]avi.Part{
		"engine" : &engine,
		"t0" : &thruster0,
		"t1" : &thruster1,
		"gun" : &weapon,
		"sensor" : &sensor,
	}
}
