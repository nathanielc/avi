package main

import (
	"flag"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/golang/glog"
	"github.com/golang/snappy"
	"github.com/nathanielc/avi"
	"github.com/nathanielc/avi/head"
	_ "github.com/nathanielc/avi/ships"
)

var partsFile = flag.String("parts", "parts.yaml", "YAML file where available parts are defined.")
var mapFile = flag.String("map", "map.yaml", "YAML file that defines the map.")
var saveFile = flag.String("save", "save.avi", "Where to save the simulation data.")
var maxTime = flag.Duration("max-time", 20*time.Minute, "Optional maximum time to simulate, (note this is simulation time not real time).")
var simFPS = flag.Int("fps", 60, "Target FPS")
var live = flag.Bool("live", false, "whether to display the simulation live.")
var cpuProfile = flag.String("cpuprofile", "", "if defined save a cpu profile to path.")
var memProfile = flag.String("memprofile", "", "if defined save a mem profile to path.")
var replay = flag.String("replay", "", "if defined load save file and replay it")

func main() {

	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var mem io.Writer
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		mem = f
	}

	if *replay != "" {
		f, err := os.Open(*replay)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		s := snappy.NewReader(f)
		updates := head.ProtoStreamUpdates(s)
		head.Run(updates)
		return
	}

	// Load parts
	pf, err := os.Open(*partsFile)
	if err != nil {
		glog.Error(err)
		return
	}
	defer pf.Close()
	parts, err := avi.LoadPartsFromFile(pf)
	if err != nil {
		glog.Error(err)
		return
	}

	// Load map
	mf, err := os.Open(*mapFile)
	if err != nil {
		glog.Error(err)
		return
	}
	defer mf.Close()
	mp, err := avi.LoadMapFromFile(mf)
	if err != nil {
		glog.Error(err)
		return
	}

	// Load fleets
	fleetFiles := flag.Args()
	fleets := make([]*avi.FleetConf, 0, len(fleetFiles))
	for _, fleetFile := range fleetFiles {
		ff, err := os.Open(fleetFile)
		if err != nil {
			glog.Error(err)
			return
		}
		defer ff.Close()
		fleet, err := avi.LoadFleetFromFile(ff)
		if err != nil {
			glog.Error(err)
			return
		}

		fleets = append(fleets, fleet)
	}

	var drawer avi.Drawer
	f, _ := os.Create(*saveFile)
	defer f.Close()
	s := snappy.NewBufferedWriter(f)
	drawer = head.NewProtoStream(s, *simFPS)
	if *live {
		ls := head.NewLiveStream()
		drawer = head.NewProxyStream(drawer, ls)
		go head.Run(ls.Updates())
	}

	sim, err := avi.NewSimulation(
		mp,
		parts,
		fleets,
		drawer,
		*maxTime,
		int64(*simFPS),
		mem,
	)
	if err != nil {
		glog.Error(err)
		return
	}
	sim.Start()
}
