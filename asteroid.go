package avi

const asteroidTexture = "asteroid"

type asteroid struct {
	objectT
	texture string
}

type AsteroidConf struct {
	Mass     float64   `yaml:"mass" json:"mass"`
	Radius   float64   `yaml:"radius" json:"radius"`
	Position []float64 `yaml:"position" json:"position"`
	Texture  string    `yaml:"texture" json:"texture"`
}

func NewAsteroid(id ID, conf AsteroidConf) (*asteroid, error) {
	pos, err := sliceToVec(conf.Position)
	if err != nil {
		return nil, err
	}
	texture := conf.Texture
	if texture == "" {
		texture = asteroidTexture
	}
	return &asteroid{
		objectT: objectT{
			id:       id,
			position: pos,
			mass:     conf.Mass,
			radius:   conf.Radius,
		},
		texture: texture,
	}, nil
}

func (a asteroid) Texture() string {
	return a.texture
}
