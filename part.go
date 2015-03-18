package avi


type Part interface {
	object
	setShip(*shipT)
}

type partT struct {
	objectT
	ship     *shipT
}

func (part *partT) setShip(ship *shipT) {
	part.ship = ship
}
