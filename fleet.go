package avi

import (
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)

type ShipPartConf struct {
	Name     string
	Position []float64
	Type     string
}

type ShipConf struct {
	Name     string
	Texture  string
	Position []float64
	Parts    []ShipPartConf
}

type FleetConf struct {
	Name   string
	Center []float64
	Ships  []ShipConf
}

func LoadFleetFromFile(f io.Reader) (*FleetConf, error) {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return LoadFleetFromData(data)

}

func LoadFleetFromData(in []byte) (*FleetConf, error) {
	conf := FleetConf{}
	err := yaml.Unmarshal(in, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
