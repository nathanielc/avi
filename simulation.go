package avi

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	"math"
)

const minSectorSize = 100
const TimePerTick = 1e-2
const small = 1e-6

const impulseToDamage = 0.25

var maxTicks = flag.Int("ticks", -1, "Optional maximum ticks to simulate")
var streamRate = flag.Int("rate", 10, "Every 'rate' ticks emit a frame")

type Simulation struct {
	ships      []*shipT
	inrts      []Object
	projs      []*projectile
	ctlps      []*controlPoint
	astds      []*asteroid
	tick       int64
	radius     int64
	sectorSize int64
	//Number of ships alive from each fleet
	survivors map[string]int
	scores    map[string]float64
	maxScore  float64
	//Available parts
	availableParts *PartsConf
	//ID counter
	idCounter int64
	stream    *Stream
	rate      int64
}

func NewSimulation(mp *MapConf, parts *PartsConf, fleets []*FleetConf, stream *Stream) (*Simulation, error) {
	sim := &Simulation{
		radius:         mp.Radius,
		availableParts: parts,
		survivors:      make(map[string]int),
		scores:         make(map[string]float64),
		maxScore:       mp.Rules.Score,
		rate:           int64(*streamRate),
		stream:         stream,
	}
	// Add Control Points
	for _, cp := range mp.ControlPoints {
		sim.addControlPoint(cp)
	}
	// Add Asteroids
	for _, asteroid := range mp.Asteroids {
		sim.addAsteroid(asteroid)
	}
	for i, fleet := range fleets {

		fleetMass := 0.0

		if i == len(mp.StartingPoints) {
			err := errors.New(fmt.Sprintf("Too many fleets for the map, only %d fleets allowed", len(mp.StartingPoints)))
			return nil, err
		}
		center, err := sliceToVec(mp.StartingPoints[i])
		if err != nil {
			return nil, err
		}

		for _, shipConf := range fleet.Ships {

			glog.Infof("Adding ship with pilot %s for fleet %s", shipConf.Pilot, fleet.Name)
			pilot := getPilot(shipConf.Pilot)
			if pilot == nil {
				return nil, errors.New(fmt.Sprintf("Unknown pilot '%s'", shipConf.Pilot))
			}
			pilot.JoinFleet(fleet.Name)

			relativePos, err := sliceToVec(shipConf.Position)
			if err != nil {
				return nil, err
			}

			pos := center.Add(relativePos)
			ship, err := sim.AddShip(fleet.Name, pos, pilot, shipConf)
			if err != nil {
				return nil, err
			}

			fleetMass += ship.GetMass()
		}

		if fleetMass > mp.Rules.MaxFleetMass {
			err = errors.New(fmt.Sprintf("Mass for fleet '%s' is too large '%f' > '%f'", fleet.Name, fleetMass, mp.Rules.MaxFleetMass))
			glog.Errorln(err)
			return nil, err
		}

	}

	return sim, nil
}

func (sim *Simulation) getNextID() int64 {
	id := sim.idCounter
	sim.idCounter++
	return id
}

func (sim *Simulation) AddShip(fleet string, pos mgl64.Vec3, pilot Pilot, conf ShipConf) (*shipT, error) {
	ship, err := newShip(sim.getNextID(), sim, fleet, pos, pilot, conf)
	if err != nil {
		return nil, err
	}
	sim.ships = append(sim.ships, ship)

	sim.survivors[fleet]++
	return ship, nil
}
func (sim *Simulation) removeShip(i int) {

	ship := sim.ships[i]
	sim.ships = append(sim.ships[:i], sim.ships[i+1:]...)

	sim.survivors[ship.fleet]--
}

func (sim *Simulation) addProjectile(pos, vel mgl64.Vec3, mass, radius float64) {
	p := &projectile{
		objectT{
			id:       sim.getNextID(),
			position: pos,
			velocity: vel,
			mass:     mass,
			radius:   radius,
		},
	}
	sim.projs = append(sim.projs, p)
}

func (sim *Simulation) addControlPoint(cpConf controlPointConf) {

	cp, err := NewControlPoint(sim.getNextID(), cpConf)
	if err != nil {
		glog.Error(err)
		return
	}

	sim.inrts = append(sim.inrts, cp)
	sim.ctlps = append(sim.ctlps, cp)

}

func (sim *Simulation) addAsteroid(aConf asteroidConf) {

	as, err := NewAsteroid(sim.getNextID(), aConf)
	if err != nil {
		glog.Error(err)
		return
	}

	sim.inrts = append(sim.inrts, as)
	sim.astds = append(sim.astds, as)

}

func (sim *Simulation) Start() {
	glog.Infoln("Starting AVI Simulation")

	for fleet := range sim.survivors {
		sim.scores[fleet] = 0.0
	}

	fleet, score := sim.loop()

	glog.Infoln("All scores:", sim.scores)
	glog.Infof("%s win with %f @ tick: %d!!!", fleet, score, sim.tick)
}

func (sim *Simulation) loop() (string, float64) {
	cont := true
	score := 0.0
	maxTicks := int64(*maxTicks)
	for cont && !(maxTicks > 0 && maxTicks < sim.tick+1) && (score < sim.maxScore) {
		score, cont = sim.doTick()
		if sim.stream != nil && sim.tick%sim.rate == 0 {
			sim.stream.SendFrame(sim.scores, sim.ships, sim.projs, sim.astds, sim.ctlps)
		}
	}
	var fleet string
	score = 0.0
	for f, s := range sim.scores {
		if s > score {
			fleet = f
			score = s
		}
	}
	return fleet, score
}

func (sim *Simulation) doTick() (float64, bool) {

	score := sim.scoreFleets()
	sim.tickShips()
	sim.propagateObjects()
	sim.collideObjects()
	sim.tick++
	return score, true
}
func (sim *Simulation) scoreFleets() float64 {
	score := 0.0
	for _, cp := range sim.ctlps {
		for _, ship := range sim.ships {
			distance := cp.position.Sub(ship.position).Len()
			if distance < cp.influence {
				sim.scores[ship.fleet] += cp.points * TimePerTick
			}
			if s := sim.scores[ship.fleet]; s > score {
				score = s
			}
		}
	}
	return score
}

func (sim *Simulation) tickShips() {
	n := len(sim.ships)
	complete := make(chan int, n)
	for _, ship := range sim.ships {
		ship := ship
		go func() {
			ship.Energize()
			ship.Tick()
			complete <- 1
		}()
	}
	for i := 0; i < n; i++ {
		<-complete
	}
}

func (sim *Simulation) propagateObjects() {
	sim.sectorSize = minSectorSize
	for _, ship := range sim.ships {
		glog.V(4).Infoln("S: ",
			ship.GetPosition(),
			ship.GetVelocity(),
			ship.GetRadius(),
		)
		sim.propagateObject(ship)
		if r := int64(ship.GetRadius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}
	for _, inrt := range sim.inrts {
		sim.propagateObject(inrt)
		if r := int64(inrt.GetRadius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}
	for _, proj := range sim.projs {
		sim.propagateObject(proj)
		if r := int64(proj.GetRadius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}

	glog.V(4).Infoln("Sector size", sim.sectorSize)
}

func (sim *Simulation) propagateObject(obj Object) {
	obj.setPosition(obj.GetPosition().Add(obj.GetVelocity().Mul(TimePerTick)))
}

func (sim *Simulation) collideObjects() {

	const OO_COR = 0.7
	const PO_COR = 0.1

	othrSectors := make(map[int64][]Object)
	projSectors := make(map[int64][]Object)
	glog.V(4).Infoln("#ships", len(sim.ships))
	glog.V(4).Infoln("#projs", len(sim.projs))
	for i := 0; i < len(sim.ships); {
		ship := sim.ships[i]
		if ship.GetHealth() <= 0 || ship.GetPosition().Len() > float64(sim.radius) {
			sim.removeShip(i)
			continue
		}
		glog.V(4).Infoln(i, ship)
		sim.placeInSectors(ship, othrSectors)
		i++
	}
	for i := 0; i < len(sim.inrts); {
		inrt := sim.inrts[i]
		glog.V(4).Infoln(i, inrt)
		sim.placeInSectors(inrt, othrSectors)
		i++
	}
	for i := 0; i < len(sim.projs); {
		proj := sim.projs[i]
		if proj.GetHealth() < 0 || proj.GetPosition().Len() > float64(sim.radius) {
			sim.projs = append(sim.projs[:i], sim.projs[i+1:]...)
			continue
		}
		sim.placeInSectors(proj, projSectors)
		i++
	}
	glog.V(4).Infoln(othrSectors)
	glog.V(4).Infoln(projSectors)

	for _, sector := range othrSectors {
		l := len(sector)
	othrothr:
		for i := 0; i < l; i++ {
			othr1 := sector[i]
			for j := i + 1; j < l; j++ {
				othr2 := sector[j]
				if othr1 == othr2 {
					continue
				}
				if collide(othr1, othr2, OO_COR) {
					continue othrothr
				}
			}
		}
	}
	for sectorIndex, othrs := range othrSectors {
		projs := projSectors[sectorIndex]
		numothrs := len(othrs)
		numProj := len(projs)

		if numProj == 0 {
			continue
		}

	projothr:
		for i := 0; i < numothrs; i++ {
			othr := othrs[i]
			for j := 0; j < numProj; j++ {
				proj := projs[j].(*projectile)
				if collide(proj, othr, PO_COR) {
					continue projothr
				}
			}
		}
	}
}

func (sim *Simulation) placeInSectors(obj Object, sectors map[int64][]Object) {
	pos := obj.GetPosition()
	radius := obj.GetRadius() + obj.GetVelocity().Len()

	numSectors := sim.radius * 2 / sim.sectorSize
	numSectors2 := numSectors * numSectors
	glog.V(4).Infoln("#S:", numSectors)

	sectorsList := make(map[int64]bool)
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			for k := -1; k <= 1; k++ {
				x := int64(pos[0]+radius*float64(i)) / sim.sectorSize
				y := int64(pos[1]+radius*float64(j)) / sim.sectorSize
				z := int64(pos[2]+radius*float64(k)) / sim.sectorSize

				index := x + y*numSectors + z*numSectors2

				if !sectorsList[index] {
					sectorsList[index] = true
					sectors[index] = append(sectors[index], obj)
				}
			}
		}
	}

	glog.V(4).Infoln("Placed in sectors:", sectorsList)
}

func collide(obj1, obj2 Object, cor float64) bool {
	// Convert to the moving reference frame of obj2
	staticPos := obj2.GetPosition()
	dynamicPos := obj1.GetPosition()
	dynamicVel := obj1.GetVelocity().Sub(obj2.GetVelocity()).Mul(TimePerTick)
	maxRange := dynamicVel.Len()

	delta := staticPos.Sub(dynamicPos)
	distance := delta.Len()

	sumRadii := obj1.GetRadius() + obj2.GetRadius()
	distanceRadii := distance - sumRadii
	//Not close enough
	if maxRange < distanceRadii {
		return false
		//} else if distanceRadii < 0 {
		//	//We have a static collision
		//	resolveCollision(obj1, obj2, cor)
		//	return true
	}

	norm := dynamicVel.Normalize()

	direction := norm.Dot(delta)
	// Going the wrong direction
	if direction <= 0 {
		return false
	}

	f := (distance * distance) - (direction * direction)

	sumRadiiSquared := sumRadii * sumRadii
	// Still not close enough
	if f >= sumRadiiSquared {
		return false
	}

	t := sumRadiiSquared - f

	// Invalid geometry no collision
	if t < 0 {
		return false
	}

	travelDist := distance - math.Sqrt(t)

	// Didn't get close enough no collision
	if maxRange < travelDist {
		return false
	}

	// We have a collision determine the position of the collision
	ratio := travelDist / maxRange

	//Place object next to each other at point of collision
	v1 := obj1.GetVelocity().Mul(ratio * TimePerTick)
	v2 := obj2.GetVelocity().Mul(ratio * TimePerTick)

	obj1.setPosition(obj1.GetPosition().Add(v1))
	obj2.setPosition(obj2.GetPosition().Add(v2))

	//Resolve collision
	resolveCollision(obj1, obj2, cor)

	return true
}

func resolveCollision(obj1, obj2 Object, cor float64) {
	norm := obj1.GetPosition().Sub(obj2.GetPosition()).Normalize()

	// inverse mass
	im1 := 1.0 / obj1.GetMass()
	im2 := 1.0 / obj2.GetMass()

	// impact speed
	v1 := obj1.GetVelocity()
	v2 := obj2.GetVelocity()

	v := v1.Sub(v2)
	vn := v.Dot(norm)

	actual := (-(1.0 + cor) * vn) / (im1 + im2)
	elastic := (-2.0 * vn) / (im1 + im2)
	impulse := norm.Mul(actual)

	damage := impulseToDamage * (elastic - actual)

	obj1.setHealth(obj1.GetHealth() - damage)
	obj2.setHealth(obj2.GetHealth() - damage)

	obj1.setVelocity(v1.Add(impulse.Mul(im1)))
	obj2.setVelocity(v2.Sub(impulse.Mul(im2)))
}
