package avi

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl64"
	"errors"
)

func sliceToVec(pos []float64) (mgl64.Vec3, error) {
	if len(pos) != 3 {
		return mgl64.Vec3{}, errors.New(fmt.Sprintf("Invalid position list must have exactly 3 items: '%v'", pos))
	}

	return mgl64.Vec3{pos[0], pos[1], pos[2]}, nil
}
