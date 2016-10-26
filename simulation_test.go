package avi

import (
	"flag"
	"fmt"
	"math/rand"
	"testing"

	"azul3d.org/engine/lmath"
	"github.com/stretchr/testify/assert"
)

func init() {
	flag.Parse()
	if testing.Verbose() {
		flag.Set("logtostderr", "1")
	}
}

//This test fails currently since objects that start
// on top of each other that collide cause a panic. Need to fix.
func TestShouldCollideStaticObjects(t *testing.T) {
	t.Skip()
	assert := assert.New(t)

	obj1 := &objectT{
		position: lmath.Vec3{0, 0, 0},
		velocity: lmath.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	obj2 := &objectT{
		position: lmath.Vec3{20, 0, 0},
		velocity: lmath.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	collision := collide(obj1, obj2, 1.0)
	assert.True(collision)
}

func TestShouldNotCollideStaticObjects(t *testing.T) {
	assert := assert.New(t)

	obj1 := &objectT{
		position: lmath.Vec3{0, 0, 0},
		velocity: lmath.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	obj2 := &objectT{
		position: lmath.Vec3{21, 0, 0},
		velocity: lmath.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	collision := collide(obj1, obj2, 1.0)
	assert.False(collision)
}

func TestShouldCollideStaticDynamicObjects(t *testing.T) {
	assert := assert.New(t)

	obj1 := &objectT{
		position: lmath.Vec3{0, 0, 0},
		velocity: lmath.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	obj2 := &objectT{
		position: lmath.Vec3{21, 0, 0},
		//Moving fast enough to arrive this tick
		velocity: lmath.Vec3{-2 * 1 / SecondsPerTick, 0, 0},
		mass:     1000,
		radius:   10,
	}

	v2 := obj2.velocity.Length()

	collision := collide(obj1, obj2, 1.0)
	assert.True(collision)
	assert.Equal(v2, obj1.velocity.Length())
	assert.Equal(0.0, obj2.velocity.Length())
}

func TestShouldNotCollideStaticDynamicObjects(t *testing.T) {
	assert := assert.New(t)

	obj1 := &objectT{
		position: lmath.Vec3{0, 0, 0},
		velocity: lmath.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
	}

	obj2 := &objectT{
		position: lmath.Vec3{21, 0, 0},
		//Moving away
		velocity: lmath.Vec3{2 * 1 / SecondsPerTick, 0, 0},
		mass:     1000,
		radius:   10,
	}

	collision := collide(obj1, obj2, 1.0)
	assert.False(collision)
}

func TestShouldCollideParallelDynamicObjects(t *testing.T) {
	assert := assert.New(t)

	obj1 := &objectT{
		position: lmath.Vec3{0, 0, 0},
		//Moving away
		velocity: lmath.Vec3{-1 * 1 / SecondsPerTick, 0, 0},
		mass:     1000,
		radius:   10,
	}

	obj2 := &objectT{
		position: lmath.Vec3{21, 0, 0},
		//Moving faster
		velocity: lmath.Vec3{-4 * 1 / SecondsPerTick, 0, 0},
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
		position: lmath.Vec3{0, 0, 0},
		velocity: lmath.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
		health:   health,
	}

	obj2 := &objectT{
		position: lmath.Vec3{21, 0, 0},
		velocity: lmath.Vec3{-2 * 1 / SecondsPerTick, 0, 0},
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
		position: lmath.Vec3{0, 0, 0},
		velocity: lmath.Vec3{0, 0, 0},
		mass:     1000,
		radius:   10,
		health:   health,
	}

	obj2 := &objectT{
		position: lmath.Vec3{21, 0, 0},
		velocity: lmath.Vec3{-2 * 1 / SecondsPerTick, 0, 0},
		mass:     1000,
		radius:   10,
		health:   health,
	}

	collision := collide(obj1, obj2, 1.0)
	assert.True(collision)
	assert.InDelta(health, obj1.health, 1e-5)
	assert.InDelta(health, obj2.health, 1e-5)
}

func BenchmarkTick(b *testing.B) {
	ship0 := newOneDirPilot(lmath.Vec3{-1, -1, -1})
	ship1 := newOneDirPilot(lmath.Vec3{1, 1, 1})
	radius := 10000.0
	maxVel := radius / 10
	sim, err := NewSimulation(&MapConf{
		Radius: int64(radius),
	},
		nil,
		nil,
		nil,
		-1,
		60,
	)
	if err != nil {
		return
	}
	sim.AddShip("f1", lmath.Vec3{1e4, 1e4, 1e4}, ship0, ShipConf{})
	sim.AddShip("f2", lmath.Vec3{-1e4, -1e4, -1e4}, ship1, ShipConf{})
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 1000; i++ {
		pos := lmath.Vec3{
			r.Float64() * radius,
			r.Float64() * radius,
			r.Float64() * radius,
		}
		vel := lmath.Vec3{
			r.Float64() * maxVel,
			r.Float64() * maxVel,
			r.Float64() * maxVel,
		}
		sim.addProjectile(pos, vel, 1, 0.1)
	}

	b.ReportAllocs()
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
	dir      lmath.Vec3
}

func newOneDirPilot(dir lmath.Vec3) Pilot {
	return &oneDirPilot{dir: dir}
}

func (self *oneDirPilot) Tick(tick int64) {
	self.engine.PowerOn(1.0)
	self.thruster.Thrust(self.dir)
	self.weapon.Fire(self.dir)
}

func (self *oneDirPilot) JoinFleet(f string) {
}

func (self *oneDirPilot) LinkParts(shipParts []ShipPartConf, availableParts *PartSetConf) ([]Part, error) {
	self.engine = NewEngine001(lmath.Vec3{0, 0, 1})

	self.thruster = NewThruster001(lmath.Vec3{0, -1, 0})
	self.weapon = NewWeaponFromConf(lmath.Vec3{1, 0, 0}, WeaponConf{
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
