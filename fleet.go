package avi

import (
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)


type FleetConf struct {
	Name   string
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
