package avi

const controlPointTexture = "control_point"

type controlPoint struct {
	objectT
	points    float64
	influence float64
}

type ControlPointConf struct {
	Mass      float64   `yaml:"mass" json:"mass"`
	Radius    float64   `yaml:"radius" json:"radius"`
	Position  []float64 `yaml:"position" json:"position"`
	Points    float64   `yaml:"points" json:"points"`
	Influence float64   `yaml:"influence" json:"influence"`
}

func NewControlPoint(id ID, conf ControlPointConf) (*controlPoint, error) {
	pos, err := sliceToVec(conf.Position)
	if err != nil {
		return nil, err
	}
	return &controlPoint{
		objectT: objectT{
			id:       id,
			position: pos,
			mass:     conf.Mass,
			radius:   conf.Radius,
		},
		points:    conf.Points,
		influence: conf.Influence,
	}, nil
}

func (controlPoint) Texture() string {
	return controlPointTexture
}
