package avi

import (
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)

type MapConf struct {
	Radius        int64
	Asteroids     []asteroidConf
	ControlPoints []controlPointConf `yaml:"control_points"`
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
