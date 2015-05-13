package avi

import (
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)

type RulesConf struct {
	Score        float64
	MaxFleetMass float64 `yaml:"max_fleet_mass"`
}

type MapConf struct {
	Radius         int64
	Asteroids      []asteroidConf
	ControlPoints  []controlPointConf `yaml:"control_points"`
	StartingPoints [][]float64        `yaml:"starting_points"`
    Rules          RulesConf
}

func LoadMapFromFile(f io.Reader) (*MapConf, error) {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return LoadMapFromData(data)
}

func LoadMapFromData(in []byte) (*MapConf, error) {
	conf := MapConf{}
	err := yaml.Unmarshal(in, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
