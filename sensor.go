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
			objectT: objectT{
				position: pos,
				mass:     100,
				radius:   0.5,
			},
		},
		energy: 1,
	}
}

