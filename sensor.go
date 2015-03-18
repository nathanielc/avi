package avi

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Sensor struct {
	partT
	energy float64
}

func NewSensor001(pos mgl64.Vec3) *Sensor {
	return &Sensor{
		partT: partT{
			Position: pos,
			Mass:     10,
		},
		energy: 1,
	}
}

