package avi

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/go-gl/mathgl/mgl64"
)

func TestSubspaceIndex(t *testing.T) {
	assert := assert.New(t)

	index := int8(0)
	for k := 0; k < 3; k++ {
		for j := 0; j < 3; j++ {
			for i := 0; i < 3; i++ {
				subspace := calcSubspaceIndex(int64(i), int64(j), int64(k))
				assert.Equal(index, subspace)

				index++
			}
		}
	}
}

func BenchmarkLoop(b *testing.B) {
	ship0 := newOneDirShip(mgl64.Vec3{-1,-1,-1})
	ship1 := newOneDirShip(mgl64.Vec3{1,1,1})
	sim := NewSimulation(100000)
	sim.AddShip(mgl64.Vec3{100, 100, 100}, ship0)
	sim.AddShip(mgl64.Vec3{-100, -100, -100}, ship1)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
		sim.doTick()
    }
}
