package avi

import "azul3d.org/engine/lmath"

type Drawable interface {
	ID() ID
	Position() lmath.Vec3
	Radius() float64
	Texture() string
}

type Drawer interface {
	Draw(
		scores map[string]float64,
		new,
		existing []Drawable,
		deleted []ID,
	)
}
