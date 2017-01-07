package avi

import "github.com/go-gl/mathgl/mgl64"

func LengthSq(v mgl64.Vec3) float64 {
	return v.X()*v.X() + v.Y()*v.Y() + v.Z()*v.Z()
}
