package avi

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/nvcook42/avi/logger"
)

const minSectorSize = 100
const timePerTick = 1e-2

const impulseToDamage = 1.0

var maxTicks = flag.Int("ticks", -1, "Optional maximum ticks to simulate")
var streamRate = flag.Int("rate", 10, "Every 'rate' ticks emit a frame")

type Simulation struct {
	ships      []*shipT
	projs      []*projectile
	tick       int64
	radius     int64
	sectorSize int64
	//Number of ships alive from each fleet
	survivors map[string]int
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
		rate:           int64(*streamRate),
		stream:         stream,
	}
	for _, fleet := range fleets {

		center, err := sliceToVec(fleet.Center)
		if err != nil {
			return nil, err
		}

		for _, shipConf := range fleet.Ships {

			logger.Info.Printf("Adding ship %s for fleet %s", shipConf.Name, fleet.Name)
			ship := getShipByName(shipConf.Name)
			if ship == nil {
				return nil, errors.New(fmt.Sprintf("Unknown ship name '%s'", shipConf.Name))
			}

			relativePos, err := sliceToVec(shipConf.Position)
			if err != nil {
				return nil, err
			}

			pos := center.Add(relativePos)
			err = sim.AddShip(fleet.Name, pos, ship, shipConf)
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

func (sim *Simulation) AddShip(fleet string, pos mgl64.Vec3, ship Ship, conf ShipConf) error {
	s, err := newShip(sim.getNextID(), sim, fleet, pos, ship, conf)
	if err != nil {
		return err
	}
	sim.ships = append(sim.ships, s)

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

func (sim *Simulation) Start() {
	logger.Info.Println("Starting AVI Simulation")

	fleet := sim.loop()

	logger.Info.Printf("%s are the only survivors @ tick: %d!!!", fleet, sim.tick)
}

func (sim *Simulation) loop() string {
	var fleet string
	cont := true
	maxTicks := int64(*maxTicks)
	for cont && !(maxTicks > 0 && maxTicks < sim.tick+1) {
		fleet, cont = sim.doTick()
		//logger.Debug.Println(sim.tick)
		if sim.stream != nil && sim.tick%sim.rate == 0 {
			sim.stream.SendFrame(sim.ships, sim.projs)
		}
	}
	return fleet
}

func (sim *Simulation) doTick() (string, bool) {

	numberAlive := 0
	var fleet string
	for f, count := range sim.survivors {
		if count > 0 {
			numberAlive++
			fleet = f
		}
	}

	if numberAlive == 1 {
		return fleet, false
	}
	sim.tickShips()
	sim.propagateObjects()
	sim.collideObjects()
	sim.tick++
	return fleet, true
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
		logger.Debug.Println("S: ",
			ship.GetPosition(),
			ship.GetVelocity(),
			ship.GetRadius(),
		)
		ship.position = ship.position.Add(ship.velocity.Mul(timePerTick))
		if r := int64(ship.GetRadius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}
	if len(sim.projs) > 0 {
		logger.Debug.Println("P: ",
			sim.projs[0].GetPosition(),
			sim.projs[0].GetVelocity(),
		)
	}
	for _, proj := range sim.projs {
		proj.position = proj.position.Add(proj.velocity.Mul(timePerTick))
		if r := int64(proj.GetRadius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}

	logger.Debug.Println("Sector size", sim.sectorSize)
}

func (sim *Simulation) collideObjects() {
	shipSectors := make(map[int64][]Object)
	projSectors := make(map[int64][]Object)
	logger.Debug.Println("#ships", len(sim.ships))
	logger.Debug.Println("#projs", len(sim.projs))
	for i := 0; i < len(sim.ships); {
		ship := sim.ships[i]
		if ship.GetHealth() <= 0 {
			sim.removeShip(i)
			continue
		}
		logger.Debug.Println(i, ship)
		sim.placeInSectors(ship, shipSectors)
		i++
	}
	for i := 0; i < len(sim.projs); {
		proj := sim.projs[i]
		if proj.GetHealth() < 0 {
			sim.projs = append(sim.projs[:i], sim.projs[i+1:]...)
			continue
		}
		sim.placeInSectors(proj, projSectors)
		i++
	}
	logger.Debug.Println(shipSectors)
	logger.Debug.Println(projSectors)

	// Group collisions
	shipShip := make(map[*shipT]*shipT)
	projShip := make(map[*projectile]*shipT)
	// ship-ship
	for _, sector := range shipSectors {
		l := len(sector)
		for i := 0; i < l; i++ {
			ship1 := sector[i].(*shipT)
			for j := i + 1; j < l; j++ {
				ship2 := sector[j].(*shipT)
				if collide(ship1, ship2) {
					logger.Debug.Println("SS")
					shipShip[ship1] = ship2
				}
			}
		}
	}
	// proj-ship
	for sectorIndex, ships := range shipSectors {
		projs := projSectors[sectorIndex]
		numShips := len(ships)
		numProj := len(projs)

		if numProj == 0 {
			continue
		}

		for i := 0; i < numShips; i++ {
			ship := ships[i].(*shipT)
			for j := 0; j < numProj; j++ {
				proj := projs[j].(*projectile)
				if collide(proj, ship) {
					logger.Debug.Println("PS")
					projShip[proj] = ship
				}
			}
		}
	}

	for ship1, ship2 := range shipShip {
		sim.doShipShipCollision(ship1, ship2)
	}

	for proj, ship := range projShip {
		sim.doProjShipCollision(proj, ship)
	}
}

func (sim *Simulation) placeInSectors(obj Object, sectors map[int64][]Object) {
	pos := obj.GetPosition()
	radius := obj.GetRadius()

	numSectors := sim.radius * 2 / sim.sectorSize
	numSectors2 := numSectors * numSectors
	logger.Debug.Println("#S:", numSectors)

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

	logger.Debug.Println("Placed in sectors:", sectorsList)
}

func collide(obj1, obj2 Object) bool {
	distanceApart := obj1.GetPosition().Sub(obj2.GetPosition()).Len()
	radii := obj1.GetRadius() + obj2.GetRadius()
	if distanceApart < radii {
		logger.Debug.Println("Collision", obj1.GetPosition(), obj2.GetPosition())
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

func (sim *Simulation) doShipShipCollision(ship1, ship2 *shipT) {

	const cor = 0.7
	mtd, normal := collisionPlane(ship1, ship2)
	resolveIntersection(ship1, ship2, mtd)
	resolveCollision(ship1, ship2, normal, cor)

}

func (sim *Simulation) doProjShipCollision(proj *projectile, ship *shipT) {

	logger.Debug.Println("doPSC")
	const cor = 0.1
	mtd, normal := collisionPlane(ship, proj)
	resolveIntersection(proj, ship, mtd)
	resolveCollision(proj, ship, normal, cor)

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
		logger.Debug.Println("sphere intersecting but moving away from each other already")
		return
	}

	i := (-(1.0 + cor) * vn) / (im1 + im2)
	ti := (-2.0 * vn) / (im1 + im2)
	impulse := normal.Mul(i)

	damage := impulseToDamage * (ti - i)

	logger.Debug.Println("Damage:", v1, v2, v, ti, i, damage)

	obj1.setHealth(obj1.GetHealth() - damage)
	obj2.setHealth(obj2.GetHealth() - damage)

	obj1.setVelocity(v1.Add(impulse.Mul(im1)))
	obj2.setVelocity(v2.Sub(impulse.Mul(im2)))
}
