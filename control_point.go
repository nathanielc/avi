package avi

type controlPoint struct {
	objectT
	points    float64
	influence float64
}

type controlPointConf struct {
	Mass      float64
	Radius    float64
	Position  []float64
	Points    float64
	Influence float64
}

func NewControlPoint(id int64, conf controlPointConf) (*controlPoint, error) {
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
