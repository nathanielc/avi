package avi

import (
	"errors"
	"fmt"

	"azul3d.org/engine/lmath"
)

func sliceToVec(pos []float64) (lmath.Vec3, error) {
	if len(pos) != 3 {
		return lmath.Vec3{}, errors.New(fmt.Sprintf("Invalid position list must have exactly 3 items: '%v'", pos))
	}

	return lmath.Vec3{pos[0], pos[1], pos[2]}, nil
}
