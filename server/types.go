package server

import "github.com/nathanielc/gdvariant"

type NewObject struct {
	ID       uint32
	Position gdvariant.Vector3
	Radius   float32
	Model    string
}

type ObjectUpdate struct {
	ID       uint32
	Position gdvariant.Vector3
}

type Frame struct {
	Time           float32
	Scores         map[string]float32
	NewObjects     []NewObject
	ObjectUpdates  []ObjectUpdate
	DeletedObjects []uint32
}
