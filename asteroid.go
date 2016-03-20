package avi

const asteroidTexture = "asteroid"

type asteroid struct {
	objectT
}

type asteroidConf struct {
	Mass     float64
	Radius   float64
	Position []float64
}

func NewAsteroid(id ID, conf asteroidConf) (*asteroid, error) {
	pos, err := sliceToVec(conf.Position)
	if err != nil {
		return nil, err
	}
	return &asteroid{
		objectT: objectT{
			id:       id,
			position: pos,
			mass:     conf.Mass,
			radius:   conf.Radius,
		},
	}, nil
}
func (asteroid) Texture() string {
	return asteroidTexture
}
