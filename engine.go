package avi

import (
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
			Position: pos,
			Mass:     2000,
		},
		energy: 100,
	}
}

func (self *Engine) getOutput() float64 {
	return self.currentOutput
}

