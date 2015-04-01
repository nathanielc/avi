package avi

import (
	"github.com/go-gl/mathgl/mgl64"
	"testing"
)

func BenchmarkLoop(b *testing.B) {
	ship0 := newOneDirPilot(mgl64.Vec3{-1, -1, -1})
	ship1 := newOneDirPilot(mgl64.Vec3{1, 1, 1})
	sim, err := NewSimulation(&MapConf{
		Radius: 1000,
	},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return
	}
	sim.AddShip("f1", mgl64.Vec3{100, 100, 100}, ship0, ShipConf{})
	sim.AddShip("f2", mgl64.Vec3{-100, -100, -100}, ship1, ShipConf{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sim.doTick()
	}
}

// Single Direction pilot
type oneDirPilot struct {
	engine   *Engine
	thruster *Thruster
	dir      mgl64.Vec3
}

func newOneDirPilot(dir mgl64.Vec3) Pilot {
	return &oneDirPilot{dir: dir}
}

func (self *oneDirPilot) Tick(tick int64) {
	self.engine.PowerOn(1.0)
	self.thruster.Thrust(self.dir)
}

func (self *oneDirPilot) LinkParts(shipParts []ShipPartConf, availableParts *PartsConf) ([]Part, error) {
	self.engine = NewEngine001(mgl64.Vec3{0, 0, 1})

	self.thruster = NewThruster001(mgl64.Vec3{0, -1, 0})
	return []Part{
		self.engine,
		self.thruster,
	}, nil
}
