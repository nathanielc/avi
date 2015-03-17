package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Part interface {
	GetPosition() mgl64.Vec3
	GetMass() float64
	HandleAction(action Action, ship *shipT)
}

type partT struct {
	Position mgl64.Vec3
	Mass     float64
}

func (part *partT) GetPosition() mgl64.Vec3 {
	return part.Position
}

func (part *partT) GetMass() float64 {
	return part.Mass
}
