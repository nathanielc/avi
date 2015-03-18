package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

type object interface {
	Move()
	GetPosition() mgl64.Vec3
	GetRadius() float64
	GetMass() float64
}

type objectT struct {
	position mgl64.Vec3
	velocity mgl64.Vec3
	radius   float64
	mass     float64
}

func (o *objectT) Move() {
	o.position = o.position.Add(o.velocity)
}

func (o *objectT) GetPosition() mgl64.Vec3 {
	return o.position
}

func (o *objectT) GetRadius() float64 {
	return o.radius
}

func (o *objectT) GetMass() float64 {
	return o.mass
}
