package avi

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"azul3d.org/engine/lmath"
	"github.com/golang/glog"
)

const minSectorSize = 100
const SecondsPerTick = 1e-3
const small = 1e-6

const impulseToDamage = 0.25

type Simulation struct {
	ships      []*shipT
	inrts      []Object
	projs      []*projectile
	ctlps      []*controlPoint
	astds      []*asteroid
	tick       int64
	maxTicks   int64
	radius     float64
	sectorSize int64
	//Number of ships alive from each fleet
	survivors map[string]int
	scores    map[string]float64
	maxScore  float64
	//Available parts
	availableParts PartSetConf
	//ID counter
	idCounter ID
	stream    Drawer

	added   map[ID]Drawable
	deleted []ID

	rate int64

	mu sync.Mutex
	// Ships tick wait group
	shipWG sync.WaitGroup
}

func NewSimulation(
	mp MapConf,
	parts PartSetConf,
	fleets []FleetConf,
	stream Drawer,
	maxTime time.Duration,
	fps int64,
) (*Simulation, error) {
	rate := int64(1.0 / (SecondsPerTick * float64(fps)))
	if rate < 1 {
		rate = 1
	}
	maxTicks := int64(float64(maxTime/time.Second) / float64(SecondsPerTick))
	sim := &Simulation{
		radius:         float64(mp.Radius),
		availableParts: parts,
		survivors:      make(map[string]int),
		scores:         make(map[string]float64),
		maxScore:       mp.Rules.Score,
		rate:           rate,
		maxTicks:       maxTicks,
		stream:         stream,
		added:          make(map[ID]Drawable),
	}
	// Add Control Points
	for _, cp := range mp.ControlPoints {
		sim.addControlPoint(cp)
	}
	// Add Asteroids
	for _, asteroid := range mp.Asteroids {
		sim.addAsteroid(asteroid)
	}
	// Add Fleets
	for i, fleet := range fleets {
		if i == len(mp.StartingPoints) {
			err := errors.New(fmt.Sprintf("Too many fleets for the map, only %d fleets allowed", len(mp.StartingPoints)))
			return nil, err
		}
		center, err := sliceToVec(mp.StartingPoints[i])
		if err != nil {
			return nil, err
		}
		err = sim.addFleet(center, fleet, mp.Rules.MaxFleetMass)
		if err != nil {
			return nil, err
		}

	}
	return sim, nil
}

func (sim *Simulation) getNextID() ID {
	id := sim.idCounter
	sim.idCounter++
	return id
}

func (sim *Simulation) AddShip(fleet string, pos lmath.Vec3, pilot Pilot, conf ShipConf) (*shipT, error) {
	ship, err := newShip(sim.getNextID(), sim, fleet, pos, pilot, conf)
	if err != nil {
		return nil, err
	}
	sim.ships = append(sim.ships, ship)
	sim.added[ship.id] = ship

	sim.survivors[fleet]++
	return ship, nil
}
func (sim *Simulation) removeShip(i int) {

}

func (sim *Simulation) addProjectile(pos, vel lmath.Vec3, mass, radius float64) {
	p := &projectile{
		objectT{
			id:       sim.getNextID(),
			position: pos,
			velocity: vel,
			mass:     mass,
			radius:   radius,
		},
	}
	sim.mu.Lock()
	sim.projs = append(sim.projs, p)
	sim.added[p.id] = p
	sim.mu.Unlock()
}

func (sim *Simulation) addControlPoint(cpConf ControlPointConf) {

	cp, err := NewControlPoint(sim.getNextID(), cpConf)
	if err != nil {
		glog.Error(err)
		return
	}

	sim.inrts = append(sim.inrts, cp)
	sim.ctlps = append(sim.ctlps, cp)
	sim.added[cp.id] = cp
}

func (sim *Simulation) addAsteroid(aConf AsteroidConf) {

	as, err := NewAsteroid(sim.getNextID(), aConf)
	if err != nil {
		glog.Error(err)
		return
	}

	sim.inrts = append(sim.inrts, as)
	sim.astds = append(sim.astds, as)
	sim.added[as.id] = as
}

// Adds a fleet to the imulation based on a given fleet config
func (sim *Simulation) addFleet(center lmath.Vec3, fleet FleetConf, maxMass float64) error {

	fleetMass := 0.0

	for _, shipConf := range fleet.Ships {

		glog.Infof("Adding ship with pilot %s for fleet %s", shipConf.Pilot, fleet.Name)
		pilot := getPilot(shipConf.Pilot)
		if pilot == nil {
			return errors.New(fmt.Sprintf("Unknown pilot '%s'", shipConf.Pilot))
		}
		pilot.JoinFleet(fleet.Name)

		relativePos, err := sliceToVec(shipConf.Position)
		if err != nil {
			return err
		}

		pos := center.Add(relativePos)
		ship, err := sim.AddShip(fleet.Name, pos, pilot, shipConf)
		if err != nil {
			completeErr := errors.New(fmt.Sprintf("Error adding ship '%s' to fleet '%s': %s", shipConf.Pilot, fleet.Name, err.Error()))
			return completeErr
		}

		fleetMass += ship.Mass()
	}

	if fleetMass > maxMass {
		err := errors.New(fmt.Sprintf("Mass for fleet '%s' is too large '%f' > '%f'", fleet.Name, fleetMass, maxMass))
		glog.Errorln(err)
		return err
	}
	glog.Infof("Fleet mass '%s' is %f", fleet.Name, fleetMass)

	return nil
}

func (sim *Simulation) Start() {
	glog.Infoln("Starting AVI Simulation")

	for fleet := range sim.survivors {
		sim.scores[fleet] = 0.0
	}

	c := sim.loop()

	glog.Infoln("All scores:", sim.scores)
	glog.Infof("%s win with %f beacuse %s, @ tick: %d!!!", strings.Join(c.Winners, ", "), c.Score, c.Reason, sim.tick)
}

type Condition struct {
	// Winners reports the names of the winning fleets.
	Winners []string
	Score   float64
	// Reason contains the reason for game end.
	Reason string
}

func (sim *Simulation) checkEndConditions() (Condition, bool) {
	bestFleets, score := sim.bestFleets()
	end := false
	var reason string
	switch {
	case sim.tick >= sim.maxTicks:
		reason = "max ticks reached"
		end = true
	case len(sim.ships) == 0:
		reason = "all ships have been destroyed"
		end = true
	case score > sim.maxScore:
		reason = "max score reached"
		end = true
	default:
		// Check if last survivor has best score
		var numSurvivors int
		var survivor string
		for fleet, n := range sim.survivors {
			if n > 0 {
				numSurvivors++
				survivor = fleet
			}
		}
		if numSurvivors == 1 {
			if len(bestFleets) == 1 && bestFleets[0] == survivor {
				reason = "last surviving fleet has best score"
				end = true
			}
		}
	}
	return Condition{
		Winners: bestFleets,
		Score:   score,
		Reason:  reason,
	}, end
}

func (sim *Simulation) bestFleets() (bestFleets []string, bestScore float64) {
	// Find the highest score
	for _, score := range sim.scores {
		if score > bestScore {
			bestScore = score
		}
	}
	// Find best fleets
	for fleet, score := range sim.scores {
		if score == bestScore {
			bestFleets = append(bestFleets, fleet)
		}
	}
	return
}

func (sim *Simulation) loop() Condition {
	glog.Infoln("MaxTicks", sim.maxTicks)
	tickPerSecond := int64(1 / float64(SecondsPerTick))
	var added, existing []Drawable
	for {
		if sim.tick%tickPerSecond == 0 {
			glog.Infoln("TICK:", sim.tick)
		}
		sim.doTick()
		if sim.stream != nil && sim.tick%sim.rate == 0 {
			for _, d := range sim.ships {
				if _, ok := sim.added[d.id]; !ok {
					existing = append(existing, d)
				}
			}
			for _, d := range sim.projs {
				if _, ok := sim.added[d.id]; !ok {
					existing = append(existing, d)
				}
			}
			for _, d := range sim.astds {
				if _, ok := sim.added[d.id]; !ok {
					existing = append(existing, d)
				}
			}
			for _, d := range sim.ctlps {
				if _, ok := sim.added[d.id]; !ok {
					existing = append(existing, d)
				}
			}
			// collect added
			for id, d := range sim.added {
				added = append(added, d)
				delete(sim.added, id)
			}
			sim.stream.Draw(float64(sim.tick)*SecondsPerTick, sim.scores, added, existing, sim.deleted)
			sim.deleted = sim.deleted[0:0]
			added = added[0:0]
			existing = existing[0:0]
		}
		// Check game end conditions
		if c, end := sim.checkEndConditions(); end {
			return c
		}
	}
}

func (sim *Simulation) doTick() (float64, bool) {

	score := sim.scoreFleets()
	sim.tickShips()
	sim.propagateObjects()
	sim.collideObjects()
	sim.destroyShips()
	sim.tick++
	return score, true
}
func (sim *Simulation) scoreFleets() float64 {
	score := 0.0
	for _, cp := range sim.ctlps {
		influence2 := cp.influence * cp.influence
		for _, ship := range sim.ships {
			distance2 := cp.position.Sub(ship.position).LengthSq()
			if distance2 < influence2 {
				sim.scores[ship.fleet] += cp.points * SecondsPerTick
			}
			if s := sim.scores[ship.fleet]; s > score {
				score = s
			}
		}
	}
	return score
}

func (sim *Simulation) tickShips() {
	sim.shipWG.Add(len(sim.ships))
	for _, ship := range sim.ships {
		ship := ship
		go func() {
			ship.Energize()
			ship.Tick()
			sim.shipWG.Done()
		}()
	}
	sim.shipWG.Wait()
}

func (sim *Simulation) propagateObjects() {
	sim.sectorSize = minSectorSize
	for _, ship := range sim.ships {
		if glog.V(4) {
			glog.Infoln("S: ",
				ship.Position(),
				ship.Velocity(),
				ship.Radius(),
			)
		}
		sim.propagateObject(ship)
		if r := int64(ship.Radius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}
	for _, inrt := range sim.inrts {
		sim.propagateObject(inrt)
		if r := int64(inrt.Radius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}
	for _, proj := range sim.projs {
		sim.propagateObject(proj)
		if r := int64(proj.Radius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}

	if glog.V(4) {
		glog.Infoln("Sector size", sim.sectorSize)
	}
}

func (sim *Simulation) propagateObject(obj Object) {
	if obj != nil {
		obj.setPosition(obj.Position().Add(obj.Velocity().MulScalar(SecondsPerTick)))
	}
}

func (sim *Simulation) collideObjects() {

	const OO_COR = 0.7
	const PO_COR = 0.1

	for _, ship0 := range sim.ships {
		// Collide ships with ships
		for _, ship1 := range sim.ships {
			collide(ship0, ship1, OO_COR)
		}
		// Collide ships with interts
		for _, inrt := range sim.inrts {
			collide(ship0, inrt, OO_COR)
		}
	}
	// Collide inerts with inerts
	for _, i0 := range sim.inrts {
		for _, i1 := range sim.inrts {
			collide(i0, i1, OO_COR)
		}
	}
	if glog.V(4) {
		glog.Infoln("Colliding projectiles", len(sim.projs))
	}

	// Projectiles can only collide once
	// So filter them out if they do
	projs := sim.projs[0:0]
projectiles:
	for _, p := range sim.projs {
		// Collide projectiles with ships
		for _, ship := range sim.ships {
			if collide(p, ship, PO_COR) {
				sim.deleted = append(sim.deleted, p.ID())
				continue projectiles
			}
		}
		// Collide projectiles with inerts
		for _, inrt := range sim.inrts {
			if collide(p, inrt, PO_COR) {
				sim.deleted = append(sim.deleted, p.ID())
				continue projectiles
			}
		}
		// Projectile didn't collide so keep it around
		projs = append(projs, p)
	}
	sim.projs = projs
}

func collide(obj1, obj2 Object, cor float64) bool {
	if obj1 == obj2 {
		return false
	}
	// Convert to the moving reference frame of obj2
	staticPos := obj2.Position()
	dynamicPos := obj1.Position()
	dynamicVel := obj1.Velocity().Sub(obj2.Velocity()).MulScalar(SecondsPerTick)
	maxRange := dynamicVel.Length()

	delta := staticPos.Sub(dynamicPos)
	distance := delta.Length()

	sumRadii := obj1.Radius() + obj2.Radius()
	distanceRadii := distance - sumRadii
	//Not close enough
	if maxRange < distanceRadii {
		return false
		//} else if distanceRadii < 0 {
		//	//We have a static collision
		//	resolveCollision(obj1, obj2, cor)
		//	return true
	}

	norm, _ := dynamicVel.Normalized()

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
	v1 := obj1.Velocity().MulScalar(ratio * SecondsPerTick)
	v2 := obj2.Velocity().MulScalar(ratio * SecondsPerTick)

	obj1.setPosition(obj1.Position().Add(v1))
	obj2.setPosition(obj2.Position().Add(v2))

	//Resolve collision
	resolveCollision(obj1, obj2, cor)

	return true
}

func resolveCollision(obj1, obj2 Object, cor float64) {
	norm, _ := obj1.Position().Sub(obj2.Position()).Normalized()

	// inverse mass
	im1 := 1.0 / obj1.Mass()
	im2 := 1.0 / obj2.Mass()

	// impact speed
	v1 := obj1.Velocity()
	v2 := obj2.Velocity()

	v := v1.Sub(v2)
	vn := v.Dot(norm)

	actual := (-(1.0 + cor) * vn) / (im1 + im2)
	elastic := (-2.0 * vn) / (im1 + im2)
	impulse := norm.MulScalar(actual)

	damage := impulseToDamage * (elastic - actual)

	obj1.setHealth(obj1.Health() - damage)
	obj2.setHealth(obj2.Health() - damage)

	obj1.setVelocity(v1.Add(impulse.MulScalar(im1)))
	obj2.setVelocity(v2.Sub(impulse.MulScalar(im2)))
}

func (sim *Simulation) destroyShips() {
	ships := sim.ships[0:0]
	for _, ship := range sim.ships {
		if ship.Health() <= 0 || ship.Position().Length() > sim.radius {
			sim.deleted = append(sim.deleted, ship.ID())
			sim.survivors[ship.fleet]--
		} else {
			ships = append(ships, ship)
		}
	}
	sim.ships = ships
}
