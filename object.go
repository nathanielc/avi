package avi

import (
	"errors"
	"math"

	"azul3d.org/engine/lmath"
)

type ID uint32

const NilID = math.MaxUint32

type Object interface {
	ID() ID
	setID(ID)
	Position() lmath.Vec3
	setPosition(lmath.Vec3)
	Velocity() lmath.Vec3
	setVelocity(lmath.Vec3)
	Radius() float64
	Mass() float64
	Health() float64
	setHealth(float64)
}

type objectT struct {
	id       ID
	position lmath.Vec3
	velocity lmath.Vec3
	radius   float64
	mass     float64
	health   float64
}

func (o *objectT) ID() ID {
	return o.id
}

func (o *objectT) setID(id ID) {
	o.id = id
}

func (o *objectT) Position() lmath.Vec3 {
	return o.position
}

func (o *objectT) setPosition(pos lmath.Vec3) {
	if math.IsNaN(pos.X) {
		err := errors.New("NaN detected")
		panic(err)
	}
	o.position = pos
}

func (o *objectT) Velocity() lmath.Vec3 {
	return o.velocity
}

func (o *objectT) setVelocity(v lmath.Vec3) {
	if math.IsNaN(v.X) {
		err := errors.New("NaN detected")
		panic(err)
	}
	o.velocity = v
}

func (o *objectT) Radius() float64 {
	return o.radius
}

func (o *objectT) Mass() float64 {
	return o.mass
}

func (o *objectT) Health() float64 {
	return o.health
}

func (o *objectT) setHealth(health float64) {
	o.health = health
}
