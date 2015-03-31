package avi

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)

func PartNotAvailable(name string) error {
	return errors.New(fmt.Sprintf("part '%s' not available", name))
}

type Part interface {
	Object
	setShip(*shipT)
	reset()
}

type partT struct {
	objectT
	ship *shipT
	used bool
}

func (part *partT) setShip(ship *shipT) {
	part.ship = ship
}

func (part *partT) reset() {
	part.used = false
}

type PartsConf struct {
	Engines   map[string]EngineConf
	Thrusters map[string]ThrusterConf
	Weapons   map[string]WeaponConf
	Sensors   map[string]SensorConf
}

func LoadPartsFromFile(f io.Reader) (*PartsConf, error) {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return LoadPartsFromData(data)
}

func LoadPartsFromData(in []byte) (*PartsConf, error) {
	conf := PartsConf{}
	err := yaml.Unmarshal(in, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
