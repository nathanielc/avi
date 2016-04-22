package avi

import (
	"errors"
	"fmt"
	"io"
	"math"
	"runtime/pprof"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
)

const minSectorSize = 100
const TimePerTick = 1e-3
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
	radius     int64
	sectorSize int64
	//Number of ships alive from each fleet
	survivors map[string]int
	scores    map[string]float64
	maxScore  float64
	//Available parts
	availableParts *PartsConf
	//ID counter
	idCounter ID
	stream    Drawer
	deleted   []ID
	rate      int64
	mem       io.Writer
}

func NewSimulation(
	mp *MapConf,
	parts *PartsConf,
	fleets []*FleetConf,
	stream Drawer,
	maxTime time.Duration,
	fps int64,
	mem io.Writer,
) (*Simulation, error) {
	rate := int64(1.0 / (TimePerTick * float64(fps)))
	if rate < 1 {
		rate = 1
	}
	maxTicks := int64(float64(maxTime/time.Second) / float64(TimePerTick))
	sim := &Simulation{
		radius:         mp.Radius,
		availableParts: parts,
		survivors:      make(map[string]int),
		scores:         make(map[string]float64),
		maxScore:       mp.Rules.Score,
		rate:           rate,
		maxTicks:       maxTicks,
		stream:         stream,
		mem:            mem,
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
	sim.deleted = append(sim.deleted, ship.ID())

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

// Adds a fleet to the simulation based on a given fleet config
func (sim *Simulation) addFleet(center mgl64.Vec3, fleet *FleetConf, maxMass float64) error {

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

	fleet, score := sim.loop()

	glog.Infoln("All scores:", sim.scores)
	glog.Infof("%s win with %f @ tick: %d!!!", fleet, score, sim.tick)
}

func (sim *Simulation) loop() (string, float64) {
	cont := true
	score := 0.0
	maxTicks := sim.maxTicks
	var drawables []Drawable
	for cont && !(maxTicks > 0 && maxTicks < sim.tick+1) && (score < sim.maxScore) && len(sim.ships) > 0 {
		score, cont = sim.doTick()
		if sim.stream != nil && sim.tick%sim.rate == 0 {
			if sim.mem != nil {
				pprof.WriteHeapProfile(sim.mem)
			}
			l := len(sim.ships) + len(sim.projs) + len(sim.astds) + len(sim.ctlps)

			if cap(drawables) < l {
				drawables = make([]Drawable, l)
			}
			i := 0
			for _, d := range sim.ships {
				drawables[i] = d
				i++
			}
			//for _, d := range sim.projs {
			//	drawables[i] = d
			//	i++
			//}
			for _, d := range sim.astds {
				drawables[i] = d
				i++
			}
			for _, d := range sim.ctlps {
				drawables[i] = d
				i++
			}
			sim.stream.Draw(sim.scores, drawables[:i], sim.deleted)
			sim.deleted = sim.deleted[0:0]
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
			ship.Position(),
			ship.Velocity(),
			ship.Radius(),
		)
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

	glog.V(4).Infoln("Sector size", sim.sectorSize)
}

func (sim *Simulation) propagateObject(obj Object) {
	obj.setPosition(obj.Position().Add(obj.Velocity().Mul(TimePerTick)))
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
		if ship.Health() <= 0 || ship.Position().Len() > float64(sim.radius) {
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
		if proj.Health() < 0 || proj.Position().Len() > float64(sim.radius) {
			sim.projs = append(sim.projs[:i], sim.projs[i+1:]...)
			sim.deleted = append(sim.deleted, proj.ID())
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
	pos := obj.Position()
	radius := obj.Radius() + obj.Velocity().Len()

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
	staticPos := obj2.Position()
	dynamicPos := obj1.Position()
	dynamicVel := obj1.Velocity().Sub(obj2.Velocity()).Mul(TimePerTick)
	maxRange := dynamicVel.Len()

	delta := staticPos.Sub(dynamicPos)
	distance := delta.Len()

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
	v1 := obj1.Velocity().Mul(ratio * TimePerTick)
	v2 := obj2.Velocity().Mul(ratio * TimePerTick)

	obj1.setPosition(obj1.Position().Add(v1))
	obj2.setPosition(obj2.Position().Add(v2))

	//Resolve collision
	resolveCollision(obj1, obj2, cor)

	return true
}

func resolveCollision(obj1, obj2 Object, cor float64) {
	norm := obj1.Position().Sub(obj2.Position()).Normalize()

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
	impulse := norm.Mul(actual)

	damage := impulseToDamage * (elastic - actual)

	obj1.setHealth(obj1.Health() - damage)
	obj2.setHealth(obj2.Health() - damage)

	obj1.setVelocity(v1.Add(impulse.Mul(im1)))
	obj2.setVelocity(v2.Sub(impulse.Mul(im2)))
}
