package head

import (
	"errors"
	"io"
	"math"
	"os"

	"github.com/momchil-atanasov/go-data-front/decoder/obj"

	"azul3d.org/engine/gfx"
)

type Model struct {
	Meshes   []*gfx.Mesh
	Textures []*gfx.Texture
	Shader   *gfx.Shader
}

func (m *Model) Copy() *Model {
	meshes := make([]*gfx.Mesh, len(m.Meshes))
	for i, mesh := range m.Meshes {
		meshes[i] = mesh.Copy()
	}
	textures := make([]*gfx.Texture, len(m.Textures))
	for i, t := range m.Textures {
		textures[i] = t.Copy()
		textures[i].Source = t.Source
	}
	return &Model{
		Meshes:   meshes,
		Textures: textures,
		Shader:   m.Shader.Copy(),
	}
}

func openObjFile(path string) ([]*Model, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadObj(f)
}

func loadObj(r io.Reader) ([]*Model, error) {
	dec := obj.NewDecoder(obj.DefaultLimits())
	model, err := dec.Decode(r)
	if err != nil {
		return nil, err
	}
	models := make([]*Model, len(model.Objects))
	for o, object := range model.Objects {
		meshes := make([]*gfx.Mesh, len(object.Meshes))
		var maxLength float64
		for m, mesh := range object.Meshes {
			var vertices []gfx.Vec3
			var normals []gfx.Vec3
			var texCoords []gfx.TexCoord
			var indicies []uint32
			vertexLookup := make(map[gfx.Vec3]int64)
			for _, face := range mesh.Faces {
				if len(face.References) != 3 {
					return nil, errors.New("non triangle face in object.")
				}
				for _, ref := range face.References {
					// Get Vertex
					v := model.Vertices[ref.VertexIndex]
					vec3 := gfx.Vec3{
						X: float32(v.X),
						Y: float32(v.Y),
						Z: float32(v.Z),
					}
					length := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
					if length > maxLength {
						maxLength = length
					}
					idx, ok := vertexLookup[vec3]
					if !ok {
						idx = int64(len(vertices))
						vertices = append(vertices, vec3)
						vertexLookup[vec3] = idx
					}
					indicies = append(indicies, uint32(idx))
					// Get TexCoords
					if ref.TexCoordIndex != obj.UndefinedIndex {
						tc := model.TexCoords[ref.TexCoordIndex]
						texCoords = append(texCoords, gfx.TexCoord{
							U: float32(tc.U),
							V: float32(tc.V),
						})
					}
					// Get Normals
					if ref.NormalIndex != obj.UndefinedIndex {
						n := model.Normals[ref.NormalIndex]
						vec3 := gfx.Vec3{
							X: float32(n.X),
							Y: float32(n.Y),
							Z: float32(n.Z),
						}
						normals = append(normals, vec3)
					}
				}
			}
			meshes[m] = &gfx.Mesh{}
			meshes[m].Vertices = vertices
			for i := range vertices {
				meshes[m].Vertices[i].X /= float32(maxLength)
				meshes[m].Vertices[i].Y /= float32(maxLength)
				meshes[m].Vertices[i].Z /= float32(maxLength)
			}
			meshes[m].Indices = indicies
			if len(texCoords) > 0 {
				meshes[m].TexCoords = []gfx.TexCoordSet{{Slice: texCoords}}
			}
			meshes[m].Normals = normals
		}
		models[o] = &Model{
			Meshes: meshes,
		}
	}
	return models, nil
}
