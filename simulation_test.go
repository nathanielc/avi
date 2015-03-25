package avi

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/nvcook42/avi/logger"
	"testing"
)

func BenchmarkLoop(b *testing.B) {
	logger.Init()
	ship0 := newOneDirShip(mgl64.Vec3{-1, -1, -1})
	ship1 := newOneDirShip(mgl64.Vec3{1, 1, 1})
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

// Single Direction ship
type oneDirShip struct {
	engine   *Engine
	thruster *Thruster
	dir      mgl64.Vec3
}

func newOneDirShip(dir mgl64.Vec3) Ship {
	return &oneDirShip{dir: dir}
}

func (self *oneDirShip) Tick() {
	self.engine.PowerOn(1.0)
	self.thruster.Thrust(self.dir, 1.0)
}

func (self *oneDirShip) LinkParts(shipParts []ShipPartConf, availableParts *PartsConf) ([]Part, error) {
	self.engine = NewEngine001(mgl64.Vec3{0, 0, 1})

	self.thruster = NewThruster001(mgl64.Vec3{0, -1, 0})
	return []Part{
		self.engine,
		self.thruster,
	}, nil
}
