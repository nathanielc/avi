package main

import (
	"github.com/nvcook42/avi"
	"github.com/nvcook42/avi/nathanielc"
	"runtime"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	ship := nathanielc.NewJim()
	eng := avi.NewSimulation()
	eng.AddShip(ship)
	eng.Start()
}
