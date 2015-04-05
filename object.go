package avi

import (
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"errors"
)

type Object interface {
	GetID() int64
	setID(int64)
	GetPosition() mgl64.Vec3
	setPosition(mgl64.Vec3)
	GetVelocity() mgl64.Vec3
	setVelocity(mgl64.Vec3)
	GetRadius() float64
	GetMass() float64
	GetHealth() float64
	setHealth(float64)
}

type objectT struct {
	id       int64
	position mgl64.Vec3
	velocity mgl64.Vec3
	radius   float64
	mass     float64
	health   float64
}

func (o *objectT) GetID() int64 {
	return o.id
}

func (o *objectT) setID(id int64) {
	o.id = id
}

func (o *objectT) GetPosition() mgl64.Vec3 {
	return o.position
}

func (o *objectT) setPosition(pos mgl64.Vec3) {
	if math.IsNaN(pos.X()) {
		err := errors.New("NaN detected")
		panic(err)
	}
	o.position = pos
}

func (o *objectT) GetVelocity() mgl64.Vec3 {
	return o.velocity
}

func (o *objectT) setVelocity(v mgl64.Vec3) {
	if math.IsNaN(v.X()) {
		err := errors.New("NaN detected")
		panic(err)
	}
	o.velocity = v
}

func (o *objectT) GetRadius() float64 {
	return o.radius
}

func (o *objectT) GetMass() float64 {
	return o.mass
}

func (o *objectT) GetHealth() float64 {
	return o.health
}

func (o *objectT) setHealth(health float64) {
	o.health = health
}
