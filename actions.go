package avi

type Opertation uint32

type Order struct {
	Actions []Action
}

type OrderResult struct {
	Actions []bool
}

type Action struct {
	PartID     string
	Opertation Opertation
	Args       interface{}
}
