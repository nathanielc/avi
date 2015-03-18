package main

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/nvcook42/avi"
	"github.com/nvcook42/avi/nathanielc"
	"runtime"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	ship0 := nathanielc.NewJim(mgl64.Vec3{-1,-1,-1})
	ship1 := nathanielc.NewJim(mgl64.Vec3{1,1,1})
	sim := avi.NewSimulation(100000)
	sim.AddShip(mgl64.Vec3{100, 100, 100}, ship0)
	sim.AddShip(mgl64.Vec3{-100, -100, -100}, ship1)
	sim.Start()
}


