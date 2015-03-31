package avi

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
)

const minSectorSize = 100
const timePerTick = 1e-2

const impulseToDamage = 1.1

var maxTicks = flag.Int("ticks", -1, "Optional maximum ticks to simulate")
var maxScore = flag.Int("score", 1e6, "Winning score. First fleet to achieve this score wins.")
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
	score     map[string]float64
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
		score:          make(map[string]float64),
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
	for _, fleet := range fleets {

		center, err := sliceToVec(fleet.Center)
		if err != nil {
			return nil, err
		}

		for _, shipConf := range fleet.Ships {

			glog.Infof("Adding ship %s for fleet %s", shipConf.Pilot, fleet.Name)
			pilot := getPilot(shipConf.Pilot)
			if pilot == nil {
				return nil, errors.New(fmt.Sprintf("Unknown pilot '%s'", shipConf.Pilot))
			}

			relativePos, err := sliceToVec(shipConf.Position)
			if err != nil {
				return nil, err
			}

			pos := center.Add(relativePos)
			err = sim.AddShip(fleet.Name, pos, pilot, shipConf)
			if err != nil {
				return nil, err
			}
		}
	}

	return sim, nil
}

func (sim *Simulation) getNextID() int64 {
	id := sim.idCounter
	sim.idCounter++
	return id
}

func (sim *Simulation) AddShip(fleet string, pos mgl64.Vec3, pilot Pilot, conf ShipConf) error {
	ship, err := newShip(sim.getNextID(), sim, fleet, pos, pilot, conf)
	if err != nil {
		return err
	}
	sim.ships = append(sim.ships, ship)

	sim.survivors[fleet]++
	return nil
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
		sim.score[fleet] = 0.0
	}

	fleet, score := sim.loop()

	glog.Infoln("All scores:", sim.score)
	glog.Infof("%s win with %f @ tick: %d!!!", fleet, score, sim.tick)
}

func (sim *Simulation) loop() (string, float64) {
	cont := true
	score := 0.0
	maxScore := float64(*maxScore)
	maxTicks := int64(*maxTicks)
	for cont && !(maxTicks > 0 && maxTicks < sim.tick+1) && (score < maxScore) {
		score, cont = sim.doTick()
		if sim.stream != nil && sim.tick%sim.rate == 0 {
			sim.stream.SendFrame(sim.ships, sim.projs, sim.astds, sim.ctlps)
		}
	}
	var fleet string
	score = 0.0
	for f, s := range sim.score {
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
				sim.score[ship.fleet] += cp.points * timePerTick
			}
			if s := sim.score[ship.fleet]; s > score {
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
		ship.position = ship.position.Add(ship.velocity.Mul(timePerTick))
		if r := int64(ship.GetRadius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}
	for _, inrt := range sim.inrts {
		inrt.setPosition(inrt.GetPosition().Add(inrt.GetVelocity().Mul(timePerTick)))
		if r := int64(inrt.GetRadius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}
	for _, proj := range sim.projs {
		proj.position = proj.position.Add(proj.velocity.Mul(timePerTick))
		if r := int64(proj.GetRadius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}

	glog.V(4).Infoln("Sector size", sim.sectorSize)
}

func (sim *Simulation) collideObjects() {
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

	// Group collisions
	othrOthr := make(map[Object]Object)
	projOthr := make(map[*projectile]Object)

	// othr-othr
	for _, sector := range othrSectors {
		l := len(sector)
		for i := 0; i < l; i++ {
			othr1 := sector[i]
			for j := i + 1; j < l; j++ {
				othr2 := sector[j]
				if collide(othr1, othr2) {
					glog.V(4).Infoln("SS")
					othrOthr[othr1] = othr2
				}
			}
		}
	}
	// proj-othr
	for sectorIndex, othrs := range othrSectors {
		projs := projSectors[sectorIndex]
		numothrs := len(othrs)
		numProj := len(projs)

		if numProj == 0 {
			continue
		}

		for i := 0; i < numothrs; i++ {
			othr := othrs[i]
			for j := 0; j < numProj; j++ {
				proj := projs[j].(*projectile)
				if collide(proj, othr) {
					glog.V(4).Infoln("PS")
					projOthr[proj] = othr
				}
			}
		}
	}

	for othr1, othr2 := range othrOthr {
		sim.doOthrOthrCollision(othr1, othr2)
	}

	for proj, othr := range projOthr {
		sim.doProjOthrCollision(proj, othr)
	}
}

func (sim *Simulation) placeInSectors(obj Object, sectors map[int64][]Object) {
	pos := obj.GetPosition()
	radius := obj.GetRadius()

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

func collide(obj1, obj2 Object) bool {
	distanceApart := obj1.GetPosition().Sub(obj2.GetPosition()).Len()
	radii := obj1.GetRadius() + obj2.GetRadius()
	if distanceApart < radii {
		glog.V(4).Infoln("Collision", obj1.GetPosition(), obj2.GetPosition())
		return true
	}
	return false
}

// Get the collision plane between to spheres
// This is also the translation distance between the spheres
func collisionPlane(obj1, obj2 Object) (mgl64.Vec3, mgl64.Vec3) {
	delta := obj1.GetPosition().Sub(obj2.GetPosition())
	distanceApart := delta.Len()
	radii := obj1.GetRadius() + obj2.GetRadius()
	return delta.Mul((radii - distanceApart) / distanceApart), delta.Normalize()
}

func (sim *Simulation) doOthrOthrCollision(othr1, othr2 Object) {

	const cor = 0.7
	mtd, normal := collisionPlane(othr1, othr2)
	resolveIntersection(othr1, othr2, mtd)
	resolveCollision(othr1, othr2, normal, cor)

}

func (sim *Simulation) doProjOthrCollision(proj *projectile, othr Object) {

	glog.V(4).Infoln("doPSC")
	const cor = 0.1
	mtd, normal := collisionPlane(othr, proj)
	resolveIntersection(proj, othr, mtd)
	resolveCollision(proj, othr, normal, cor)
}

func resolveIntersection(obj1, obj2 Object, mtd mgl64.Vec3) {
	// inverse mass
	im1 := 1.0 / obj1.GetMass()
	im2 := 1.0 / obj2.GetMass()

	pos1 := obj1.GetPosition()
	pos2 := obj2.GetPosition()

	obj1.setPosition(pos1.Add(mtd.Mul(im1 / (im1 + im2))))
	obj2.setPosition(pos2.Sub(mtd.Mul(im2 / (im1 + im2))))
}

func resolveCollision(obj1, obj2 Object, normal mgl64.Vec3, cor float64) {

	// inverse mass
	im1 := 1.0 / obj1.GetMass()
	im2 := 1.0 / obj2.GetMass()

	// impact speed
	v1 := obj1.GetVelocity()
	v2 := obj2.GetVelocity()

	v := v1.Sub(v2)
	vn := v.Dot(normal)

	if vn > 0 {
		glog.V(4).Infoln("sphere intersecting but moving away from each other already")
		return
	}

	i := (-(1.0 + cor) * vn) / (im1 + im2)
	ti := (-2.0 * vn) / (im1 + im2)
	impulse := normal.Mul(i)

	damage := impulseToDamage * (ti - i)

	glog.V(4).Infoln("Damage:", v1, v2, v, ti, i, damage)

	obj1.setHealth(obj1.GetHealth() - damage)
	obj2.setHealth(obj2.GetHealth() - damage)

	obj1.setVelocity(v1.Add(impulse.Mul(im1)))
	obj2.setVelocity(v2.Sub(impulse.Mul(im2)))
}
