package main

import (
	"flag"
	"compress/gzip"
	"github.com/nvcook42/avi"
	"github.com/nvcook42/avi/logger"
	_ "github.com/nvcook42/avi/ships"
	"os"
	"runtime"
)

var partsFile = flag.String("parts", "parts.yaml", "YAML file where available parts are defined")
var mapFile = flag.String("map", "map.yaml", "YAML file that defines the map")
var saveFile = flag.String("save", "save.avi", "Where to save the simulation data")

func main() {

	flag.Parse()

	logger.Init()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Load parts
	pf, err := os.Open(*partsFile)
	if err != nil {
		logger.Error.Fatal(err)
	}
	defer pf.Close()
	parts, err := avi.LoadPartsFromFile(pf)

	// Load map
	mf, err := os.Open(*mapFile)
	if err != nil {
		logger.Error.Fatal(err)
	}
	defer mf.Close()
	mp, err := avi.LoadMapFromFile(mf)
	if err != nil {
		logger.Error.Fatal(err)
	}

	// Load fleets
	fleetFiles := flag.Args()
	fleets := make([]*avi.FleetConf, 0, len(fleetFiles))
	for _, fleetFile := range fleetFiles {
		ff, err := os.Open(fleetFile)
		if err != nil {
			logger.Error.Fatal(err)
		}
		defer ff.Close()
		fleet, err := avi.LoadFleetFromFile(ff)
		if err != nil {
			logger.Error.Fatal(err)
		}

		fleets = append(fleets, fleet)
	}

	f, _ := os.Create(*saveFile)
	defer f.Close()
	g := gzip.NewWriter(f)
	defer g.Close()
	stream := avi.NewStream(g)

	sim, err := avi.NewSimulation(mp, parts, fleets, stream)
	if err != nil {
		logger.Error.Fatal(err)
	}
	sim.Start()
}
