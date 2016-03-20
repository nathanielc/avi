package head

import (
	"math"

	"azul3d.org/engine/gfx"
	"azul3d.org/engine/lmath"
)

var unitSphereMesh *gfx.Mesh

// Number of times to subdivide the icosahedron
const recursionDepth = 3

// initializes unitSphereMesh
func init() {
	unitSphereMesh = gfx.NewMesh()

	var verticies []lmath.Vec3
	vertexLookup := make(map[lmath.Vec3]uint32)
	addVertex := func(v lmath.Vec3) uint32 {
		// Normalize vertex so it lands on a unit sphere
		v, _ = v.Normalized()
		if i, ok := vertexLookup[v]; ok {
			return i
		} else {
			i = uint32(len(verticies))
			vertexLookup[v] = i
			verticies = append(verticies, v)
			return i
		}
	}

	// 12 verticies of 3 orthogonal rectangles
	t := (1.0 + math.Sqrt(5.0)) / 2.0
	for _, v := range []lmath.Vec3{
		{-1, t, 0},
		{1, t, 0},
		{-1, -t, 0},
		{1, -t, 0},

		{0, -1, t},
		{0, 1, t},
		{0, -1, -t},
		{0, 1, -t},

		{t, 0, -1},
		{t, 0, 1},
		{-t, 0, -1},
		{-t, 0, 1},
	} {
		addVertex(v)
	}

	// Indices of 20 triangle faces to icosahedron
	indices := []uint32{
		0, 11, 5,
		0, 5, 1,
		0, 1, 7,
		0, 7, 10,
		0, 10, 11,

		1, 5, 9,
		5, 11, 4,
		11, 10, 2,
		10, 7, 6,
		7, 1, 8,

		3, 9, 4,
		3, 4, 2,
		3, 2, 6,
		3, 6, 8,
		3, 8, 9,

		4, 9, 5,
		2, 4, 11,
		6, 2, 10,
		8, 6, 7,
		9, 8, 1,
	}

	getMiddlePoint := func(v0, v1 lmath.Vec3) lmath.Vec3 {
		m := lmath.Vec3{
			(v0.X + v1.X) / 2,
			(v0.Y + v1.Y) / 2,
			(v0.Z + v1.Z) / 2,
		}
		n, _ := m.Normalized()
		return n
	}

	//Refine indices by cutting each triangle into 4 smaller triangles
	for i := 0; i < recursionDepth; i++ {
		indices2 := []uint32{}
		// Loop through each triangle
		for j := 0; j < len(indices); j += 3 {
			a := addVertex(getMiddlePoint(verticies[indices[j]], verticies[indices[j+1]]))
			b := addVertex(getMiddlePoint(verticies[indices[j+1]], verticies[indices[j+2]]))
			c := addVertex(getMiddlePoint(verticies[indices[j+2]], verticies[indices[j]]))

			// Add four new triangles
			indices2 = append(indices2,
				indices[j], a, c,
				indices[j+1], b, a,
				indices[j+2], c, b,
				a, b, c,
			)
		}
		indices = indices2
	}

	// Convert vertices and compute TexCoords
	unitSphereMesh.TexCoords = []gfx.TexCoordSet{{
		Slice: make([]gfx.TexCoord, len(verticies)),
	}}
	unitSphereMesh.Vertices = make([]gfx.Vec3, len(verticies))
	for i, v := range verticies {
		unitSphereMesh.Vertices[i] = gfx.ConvertVec3(v)
		unitSphereMesh.TexCoords[0].Slice[i].U = float32(0.5 + math.Atan2(v.Z, v.X)/(2*math.Pi))
		unitSphereMesh.TexCoords[0].Slice[i].V = float32(0.5 - math.Asin(v.Y)/math.Pi)
	}
	// Set indices
	unitSphereMesh.Indices = indices
}
