package avi

import (
	"flag"
	"fmt"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	flag.Parse()
	if testing.Verbose() {
		flag.Set("logtostderr", "1")
	}
}

//This test fails currently since objects that start
// on top of each other that collide cause a panic. Need to fix.
func TestShouldColideStaticObjects(t *testing.T) {
	assert := assert.New(t)

	obj1 := &objectT{
		position: mgl64.Vec3{0, 0, 0},
		velocity: mgl64.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	obj2 := &objectT{
		position: mgl64.Vec3{20, 0, 0},
		velocity: mgl64.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	collision := collide(obj1, obj2, 1.0)
	assert.True(collision)
}

func TestShouldNotColideStaticObjects(t *testing.T) {
	assert := assert.New(t)

	obj1 := &objectT{
		position: mgl64.Vec3{0, 0, 0},
		velocity: mgl64.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	obj2 := &objectT{
		position: mgl64.Vec3{21, 0, 0},
		velocity: mgl64.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	collision := collide(obj1, obj2, 1.0)
	assert.False(collision)
}

func TestShouldColideStaticDynamicObjects(t *testing.T) {
	assert := assert.New(t)

	obj1 := &objectT{
		position: mgl64.Vec3{0, 0, 0},
		velocity: mgl64.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	obj2 := &objectT{
		position: mgl64.Vec3{21, 0, 0},
		//Moving fast enough to arrive this tick
		velocity: mgl64.Vec3{-2 * 1 / TimePerTick, 0, 0},
		mass:     1000,
		radius:   10,
	}

	v2 := obj2.velocity.Len()

	collision := collide(obj1, obj2, 1.0)
	assert.True(collision)
	assert.Equal(v2, obj1.velocity.Len())
	assert.Equal(0, obj2.velocity.Len())
}

func TestShouldNotColideStaticDynamicObjects(t *testing.T) {
	assert := assert.New(t)

	obj1 := &objectT{
		position: mgl64.Vec3{0, 0, 0},
		velocity: mgl64.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	obj2 := &objectT{
		position: mgl64.Vec3{21, 0, 0},
		//Moving away
		velocity: mgl64.Vec3{2 * 1 / TimePerTick, 0, 0},
		mass:     1000,
		radius:   10,
	}

	collision := collide(obj1, obj2, 1.0)
	assert.False(collision)
}

func TestShouldColideParallelDynamicObjects(t *testing.T) {
	assert := assert.New(t)

	obj1 := &objectT{
		position: mgl64.Vec3{0, 0, 0},
		//Moving away
		velocity: mgl64.Vec3{-1 * 1 / TimePerTick, 0, 0},
		mass:     1000,
		radius:   10,
	}

	obj2 := &objectT{
		position: mgl64.Vec3{21, 0, 0},
		//Moving faster
		velocity: mgl64.Vec3{-4 * 1 / TimePerTick, 0, 0},
		mass:     1000,
		radius:   10,
	}

	collision := collide(obj1, obj2, 1.0)
	assert.True(collision)
}

func TestCollisionShouldDoDamage(t *testing.T) {
	assert := assert.New(t)

	health := 10.0

	obj1 := &objectT{
		position: mgl64.Vec3{0, 0, 0},
		velocity: mgl64.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
		health:   health,
	}

	obj2 := &objectT{
		position: mgl64.Vec3{21, 0, 0},
		velocity: mgl64.Vec3{-2 * 1 / TimePerTick, 0, 0},
		mass:     1000,
		radius:   10,
		health:   health,
	}

	collision := collide(obj1, obj2, 0.2)
	assert.True(collision)
	assert.True(health > obj1.health, fmt.Sprint(obj1.health))
	assert.True(health > obj2.health, fmt.Sprint(obj2.health))
}

func TestElasticCollisionShouldNotDoDamage(t *testing.T) {
	assert := assert.New(t)

	health := 10.0

	obj1 := &objectT{
		position: mgl64.Vec3{0, 0, 0},
		velocity: mgl64.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
		health:   health,
	}

	obj2 := &objectT{
		position: mgl64.Vec3{21, 0, 0},
		velocity: mgl64.Vec3{-2 * 1 / TimePerTick, 0, 0},
		mass:     1000,
		radius:   10,
		health:   health,
	}

	collision := collide(obj1, obj2, 1.0)
	assert.True(collision)
	assert.InDelta(health, obj1.health, 1e-5)
	assert.InDelta(health, obj2.health, 1e-5)
}

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
	sim.AddShip("f1", mgl64.Vec3{1e4, 1e4, 1e4}, ship0, ShipConf{})
	sim.AddShip("f2", mgl64.Vec3{-1e4, -1e4, -1e4}, ship1, ShipConf{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sim.doTick()
	}
}

// Single Direction pilot
type oneDirPilot struct {
	engine   *Engine
	thruster *Thruster
	weapon   *Weapon
	dir      mgl64.Vec3
}

func newOneDirPilot(dir mgl64.Vec3) Pilot {
	return &oneDirPilot{dir: dir}
}

func (self *oneDirPilot) Tick(tick int64) {
	self.engine.PowerOn(1.0)
	self.thruster.Thrust(self.dir)
	self.weapon.Fire(self.dir)
}

func (self *oneDirPilot) JoinFleet(f string) {
}

func (self *oneDirPilot) LinkParts(shipParts []ShipPartConf, availableParts *PartsConf) ([]Part, error) {
	self.engine = NewEngine001(mgl64.Vec3{0, 0, 1})

	self.thruster = NewThruster001(mgl64.Vec3{0, -1, 0})
	self.weapon = NewWeaponFromConf(mgl64.Vec3{1, 0, 0}, WeaponConf{
		Mass:         1000,
		Radius:       1,
		Energy:       1,
		AmmoVelocity: 10,
		AmmoMass:     0.1,
		Cooldown:     0.1,
	})
	return []Part{
		self.engine,
		self.thruster,
		self.weapon,
	}, nil
}
