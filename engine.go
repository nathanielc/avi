package avi

import (
	"errors"
	"github.com/go-gl/mathgl/mgl64"
)

type Engine struct {
	partT
	energy        float64
	currentOutput float64
}

func NewEngine001(pos mgl64.Vec3) *Engine {
	return &Engine{
		partT: partT{
			objectT: objectT{
				position: pos,
				mass:     2000,
				radius:   5,
			},
		},
		energy: 100,
	}
}

func (self *Engine) getOutput() float64 {
	return self.currentOutput
}

func (self *Engine) PowerOn(power float64) error {
	if power > 1 || power < 0 {
		return errors.New("Power must be between 0 and 1")
	}

	self.currentOutput = self.energy * power
	return nil
}
