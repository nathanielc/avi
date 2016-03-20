package avi

import "github.com/go-gl/mathgl/mgl64"

type Drawable interface {
	ID() ID
	Position() mgl64.Vec3
	Radius() float64
	Texture() string
}

type Drawer interface {
	Draw(
		scores map[string]float64,
		drawables []Drawable,
		deleted []ID,
	)
}
