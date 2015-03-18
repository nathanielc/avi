package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Part interface {
	GetPosition() mgl64.Vec3
	GetMass() float64
	GetOrder() Order
}

type partT struct {
	Position mgl64.Vec3
	Mass     float64
	currentOrder Order
}

func (part *partT) GetPosition() mgl64.Vec3 {
	return part.Position
}

func (part *partT) GetMass() float64 {
	return part.Mass
}

func (part *partT) GetOrder() Order {
	order := part.currentOrder
	part.currentOrder = nil
	return order
}


