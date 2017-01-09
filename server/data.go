package server

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/nathanielc/avi"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	mapsPath     = "maps"
	partSetsPath = "part_sets"
	fleetsPath   = "fleets"
	replaysPath  = "replays"
)

type data struct {
	mapsPath,
	partSetsPath,
	fleetsPath,
	replaysPath string
}

func newData(dir string) (*data, error) {
	mp := path.Join(dir, mapsPath)
	pp := path.Join(dir, partSetsPath)
	fp := path.Join(dir, fleetsPath)
	rp := path.Join(dir, replaysPath)
	for _, p := range []string{mp, pp, fp, rp} {
		if err := os.MkdirAll(p, 0755); err != nil {
			return nil, err
		}
	}
	return &data{
		mapsPath:     mp,
		partSetsPath: pp,
		fleetsPath:   fp,
		replaysPath:  rp,
	}, nil
}

func (d *data) Maps() (map[string]avi.MapConf, error) {
	maps := make(map[string]avi.MapConf)
	err := d.readDir(d.mapsPath, func(id, ext string, data []byte) error {
		var m avi.MapConf
		switch ext {
		case ".yaml":
			if err := yaml.Unmarshal(data, &m); err != nil {
				return err
			}
		}
		maps[id] = m
		return nil
	})
	return maps, err
}
func (d *data) Map(id string) (avi.MapConf, error) {
	p := path.Join(d.mapsPath, id+".yaml")
	var m avi.MapConf
	err := unmarshalYaml(p, &m)
	return m, err
}
func (d *data) PartSets() (map[string]avi.PartSetConf, error) {
	partSets := make(map[string]avi.PartSetConf)
	err := d.readDir(d.partSetsPath, func(id, ext string, data []byte) error {
		var ps avi.PartSetConf
		switch ext {
		case ".yaml":
			if err := yaml.Unmarshal(data, &ps); err != nil {
				return err
			}
		}
		partSets[id] = ps
		return nil
	})
	return partSets, err
}
func (d *data) PartSet(id string) (avi.PartSetConf, error) {
	p := path.Join(d.partSetsPath, id+".yaml")
	var ps avi.PartSetConf
	err := unmarshalYaml(p, &ps)
	return ps, err
}
func (d *data) Fleets() (map[string]avi.FleetConf, error) {
	fleets := make(map[string]avi.FleetConf)
	err := d.readDir(d.fleetsPath, func(id, ext string, data []byte) error {
		var f avi.FleetConf
		switch ext {
		case ".yaml":
			if err := yaml.Unmarshal(data, &f); err != nil {
				return err
			}
		}
		fleets[id] = f
		return nil
	})
	return fleets, err
}
func (d *data) Fleet(id string) (avi.FleetConf, error) {
	p := path.Join(d.fleetsPath, id+".yaml")
	var f avi.FleetConf
	err := unmarshalYaml(p, &f)
	return f, err
}

func (d *data) NewReplay(gameID string) Replay {
	fpath := path.Join(d.replaysPath, gameID+".ravi")
	return Replay{
		GameID: gameID,
		fpath:  fpath,
	}
}

func (d *data) Replays() ([]Replay, error) {
	var replays []Replay
	fs, err := ioutil.ReadDir(d.replaysPath)
	if err != nil {
		return nil, err
	}
	for _, f := range fs {
		if f.Mode().IsRegular() {
			name := f.Name()
			ext := path.Ext(name)
			gameID := strings.TrimSuffix(name, ext)
			r := Replay{
				GameID: gameID,
				Date:   f.ModTime(),
				fpath:  path.Join(d.replaysPath, name),
			}
			replays = append(replays, r)
		}
	}
	return replays, nil
}

func (d *data) Replay(gameID string) (Replay, error) {
	p := path.Join(d.replaysPath, gameID+".ravi")
	fi, err := os.Stat(p)
	if err != nil {
		return Replay{}, err
	}
	return Replay{
		GameID: gameID,
		Date:   fi.ModTime(),
		fpath:  p,
	}, nil
}

func (d *data) readDir(dir string, newF func(id, ext string, data []byte) error) error {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		if f.Mode().IsRegular() {
			ext := path.Ext(f.Name())
			id := strings.TrimSuffix(path.Base(f.Name()), ext)
			r, err := os.Open(path.Join(dir, f.Name()))
			if err != nil {
				return err
			}
			defer r.Close()
			data, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			if err := newF(id, ext, data); err != nil {
				return err
			}
		}
	}
	return nil
}

func unmarshalYaml(p string, o interface{}) error {
	f, err := os.Open(p)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %q", p)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrapf(err, "failed to read file %q", p)
	}
	return yaml.Unmarshal(data, o)
}
