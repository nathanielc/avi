package main

import (
	"github.com/nvcook42/avi"
	"github.com/nvcook42/avi/nathanielc"
	"runtime"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	ship := nathanielc.NewJim()
	sim := avi.NewSimulation()
	sim.AddShip(ship)
	sim.Start()
}
