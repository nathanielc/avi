package avi

import (
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"errors"
	"fmt"
)

func PartNotAvailable(name string) error {
	return errors.New(fmt.Sprintf("part '%s' not available", name))
}

type Part interface {
	object
	setShip(*shipT)
}

type partT struct {
	objectT
	ship     *shipT
}

func (part *partT) setShip(ship *shipT) {
	part.ship = ship
}


type PartsConf struct {
	Engines map[string]EngineConf
	Thrusters map[string]ThrusterConf
	Weapons map[string]WeaponConf
	Sensors map[string]SensorConf
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
