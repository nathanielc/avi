package avi

type Pilot interface {
	JoinFleet(fleet string)
	LinkParts([]ShipPartConf, PartSetConf) ([]Part, error)
	Tick(int64)
}

type pilotFactory func() Pilot

var registeredPilots = make(map[string]pilotFactory)

//Register a ship to make it available
func RegisterPilot(pilot string, pf pilotFactory) {
	registeredPilots[pilot] = pf
}

// Get a registered ship by pilot name
func getPilot(pilot string) Pilot {
	if pf, ok := registeredPilots[pilot]; ok {
		return pf()
	}
	return nil
}

func init() {
	RegisterPilot("dud", NewDud)
}

type DudPilot struct {
	GenericPilot
}

func NewDud() Pilot {
	return &DudPilot{}
}

func (*DudPilot) Tick(tick int64) {
}
