package avi

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadEngine(t *testing.T) {
	assert := assert.New(t)

	data := []byte(`---
engines:
  coal:
    mass: 100.15
    radius: 10
    energy: 1
`)
	conf, err := LoadPartsFromData(data)
	assert.Nil(err)
	coal, ok := conf.Engines["coal"]
	assert.True(ok)
	assert.Equal(EngineConf{
		Mass:   100.15,
		Radius: 10.0,
		Energy: 1.0,
	}, coal)
}

func TestNewEngineFromConf(t *testing.T) {
	assert := assert.New(t)

	conf := EngineConf{
		Mass:   101,
		Radius: 0.1,
		Energy: 10,
	}

	pos := mgl64.Vec3{1, 4, -5}

	engine := NewEngineFromConf(pos, conf)

	assert.NotNil(engine)

	assert.Equal(conf.Mass, engine.GetMass())
	assert.Equal(conf.Radius, engine.GetRadius())
	assert.Equal(conf.Energy, engine.energy)

	assert.Equal(pos, engine.GetPosition())
}

func TestLoadSensor(t *testing.T) {
	assert := assert.New(t)

	data := []byte(`---
sensors:
  antenna:
    mass: 100.15
    radius: 10
    energy: 1
`)
	conf, err := LoadPartsFromData(data)
	assert.Nil(err)
	antenna, ok := conf.Sensors["antenna"]
	assert.True(ok)
	assert.Equal(SensorConf{
		Mass:   100.15,
		Radius: 10.0,
		Energy: 1.0,
	}, antenna)
}

func TestNewSensorFromConf(t *testing.T) {
	assert := assert.New(t)

	conf := SensorConf{
		Mass:   101,
		Radius: 0.1,
		Energy: 10,
	}

	pos := mgl64.Vec3{1, 4, -5}

	sensor := NewSensorFromConf(pos, conf)

	assert.NotNil(sensor)

	assert.Equal(conf.Mass, sensor.GetMass())
	assert.Equal(conf.Radius, sensor.GetRadius())
	assert.Equal(conf.Energy, sensor.energy)

	assert.Equal(pos, sensor.GetPosition())
}

func TestLoadThruster(t *testing.T) {
	assert := assert.New(t)

	data := []byte(`---
thrusters:
  rocket:
    mass: 1000
    radius: 15
    force: 3000
    energy: 2000
`)
	conf, err := LoadPartsFromData(data)
	assert.Nil(err)
	rocket, ok := conf.Thrusters["rocket"]
	assert.True(ok)
	assert.Equal(ThrusterConf{
		Mass:   1000,
		Radius: 15.0,
		Force:  3000.0,
		Energy: 2000,
	}, rocket)
}

func TestNewThrusterFromConf(t *testing.T) {
	assert := assert.New(t)

	conf := ThrusterConf{
		Mass:   1001,
		Radius: 12,
		Force:  1500,
		Energy: 150,
	}

	pos := mgl64.Vec3{1, 4, -5}

	Thruster := NewThrusterFromConf(pos, conf)

	assert.NotNil(Thruster)

	assert.Equal(conf.Mass, Thruster.GetMass())
	assert.Equal(conf.Radius, Thruster.GetRadius())
	assert.Equal(conf.Force, Thruster.force)
	assert.Equal(conf.Energy, Thruster.energy)

	assert.Equal(pos, Thruster.GetPosition())
}

func TestLoadWeapon(t *testing.T) {
	assert := assert.New(t)

	data := []byte(`---
weapons:
  railgun:
    mass: 100.15
    radius: 10
    energy: 1
`)
	conf, err := LoadPartsFromData(data)
	assert.Nil(err)
	railgun, ok := conf.Weapons["railgun"]
	assert.True(ok)
	assert.Equal(WeaponConf{
		Mass:   100.15,
		Radius: 10.0,
		Energy: 1.0,
	}, railgun)
}

func TestNewWeaponFromConf(t *testing.T) {
	assert := assert.New(t)

	conf := WeaponConf{
		Mass:   101,
		Radius: 0.1,
		Energy: 10,
	}

	pos := mgl64.Vec3{1, 4, -5}

	weapon := NewWeaponFromConf(pos, conf)

	assert.NotNil(weapon)

	assert.Equal(conf.Mass, weapon.GetMass())
	assert.Equal(conf.Radius, weapon.GetRadius())
	assert.Equal(conf.Energy, weapon.energy)

	assert.Equal(pos, weapon.GetPosition())
}
