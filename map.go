package avi

import (
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)

type asteroid struct {
	Mass float64
	Radius float64
	Position []float64
}

type MapConf struct {
	Radius int64
	Asteroids []asteroid
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
