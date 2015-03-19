package main

import (
	"flag"
	"github.com/nvcook42/avi"
	"github.com/nvcook42/avi/logger"
	_ "github.com/nvcook42/avi/ships"
	"os"
	"runtime"
)

var partsFile = flag.String("parts", "parts.yaml", "YAML file where available parts are defined")
var mapFile = flag.String("map", "map.yaml", "YAML file that defines the map")

func main() {

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Load parts
	pf, err := os.Open(*partsFile)
	if err != nil {
		logger.Fatal(err)
	}
	defer pf.Close()
	parts, err := avi.LoadPartsFromFile(pf)

	// Load map
	mf, err := os.Open(*mapFile)
	if err != nil {
		logger.Fatal(err)
	}
	defer mf.Close()
	mp, err := avi.LoadMapFromFile(mf)
	if err != nil {
		logger.Fatal(err)
	}

	// Load fleets
	fleetFiles := flag.Args()
	fleets := make([]*avi.FleetConf, 0, len(fleetFiles))
	for _, fleetFile := range fleetFiles {
		ff, err := os.Open(fleetFile)
		if err != nil {
			logger.Fatal(err)
		}
		defer ff.Close()
		fleet, err := avi.LoadFleetFromFile(ff)
		if err != nil {
			logger.Fatal(err)
		}

		fleets = append(fleets, fleet)
	}

	sim, err := avi.NewSimulation(mp, parts, fleets)
	if err != nil {
		logger.Fatal(err)
	}
	sim.Start()
}
