package avi

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/nvcook42/avi/logger"
	"errors"
	"fmt"
)

const minSectorSize = 100
const timePerTick = 1e-3

type Simulation struct {
	ships       []*shipT
	projectiles []*projectile
	tick        int64
	size        int64
	sectors     sectorMap
	sectorSize  int64
	collisions  []collision
	//Number of ships alive from each fleet
	survivors   map[string]int
	//Available parts
	availableParts *PartsConf
}

type sectorMapZ map[int64][]object
type sectorMapY map[int64]sectorMapZ
type sectorMap map[int64]sectorMapY

type collision struct {
	obj1 object
	obj2 object
}

func NewSimulation(mp *MapConf, parts *PartsConf, fleets []*FleetConf) (*Simulation, error) {
	sim := &Simulation{
		size: mp.Size,
		availableParts: parts,
		survivors: make(map[string]int),
	}

	for _, fleet := range fleets {
		
		center, err := sliceToVec(fleet.Center)
		if err != nil {
			return nil, err
		}

		for _, shipConf := range fleet.Ships {

			ship := getShipByName(shipConf.Name)
			if ship == nil {
				return nil, errors.New(fmt.Sprintf("Unknown ship name '%s'", shipConf.Name))
			}


			relativePos, err := sliceToVec(shipConf.Position)
			if err != nil {
				return nil, err
			}

			pos := center.Add(relativePos)
			err = sim.AddShip(fleet.Name, pos, ship, shipConf.Parts)
			if err != nil {
				return nil, err
			}
		}
	}

	return sim, nil
}

func (sim *Simulation) AddShip(fleet string, pos mgl64.Vec3, ship Ship, parts []ShipPartConf) error {
	s, err := newShip(sim, fleet, pos, ship, parts)
	if err != nil {
		return err
	}
	logger.Debugln("Added new ship", s.GetMass())
	sim.ships = append(sim.ships, s)

	sim.survivors[fleet]++
	return nil
}

func (sim *Simulation) addProjectile(p *projectile) {
	sim.projectiles = append(sim.projectiles, p)
}

func (sim *Simulation) Start() {
	logger.Infoln("Starting AVI Simulation")

	sim.loop()
}

func (sim *Simulation) loop() {
	for {
		sim.doTick()
	}
}

func (sim *Simulation) doTick() {
	sim.tickShips()
	sim.propagateObjects()
	sim.collideObjects()
	sim.tick++
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
		ship.Move()
		if r := int64(ship.GetRadius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}
	for _, p := range sim.projectiles {
		p.Move()
		if r := int64(p.GetRadius() * 2); r > sim.sectorSize {
			sim.sectorSize = r
		}
	}
}

func (sim *Simulation) collideObjects() {
	sim.sectors = make(sectorMap)
	for _, ship := range sim.ships {
		sim.collideInSectors(ship)
	}
	for _, p := range sim.projectiles {
		sim.collideInSectors(p)
	}

}

func (sim *Simulation) collideInSectors(obj object) {
	pos := obj.GetPosition()
	radius := obj.GetRadius()
	x := int64(pos[0]) / sim.sectorSize
	y := int64(pos[1]) / sim.sectorSize
	z := int64(pos[2]) / sim.sectorSize

	sectorSubspace := make([]bool, 27)
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			for k := -1; k <= 1; k++ {
				edgeX := int64(pos[0]+radius*float64(i)) / sim.sectorSize
				edgeY := int64(pos[1]+radius*float64(j)) / sim.sectorSize
				edgeZ := int64(pos[2]+radius*float64(k)) / sim.sectorSize
				subX := edgeX - x + 1
				subY := edgeY - y + 1
				subZ := edgeZ - z + 1

				subspace := calcSubspaceIndex(subX, subY, subZ)
				if !sectorSubspace[subspace] {
					sim.collideInSector(obj, edgeX, edgeY, edgeZ)
					sectorSubspace[subspace] = true
				}
			}
		}
	}
}

func calcSubspaceIndex(x, y, z int64) int8 {
	return int8(x + y*3 + z*9)
}

func (sim *Simulation) collideInSector(obj object, x, y, z int64) {
	var ymap sectorMapY
	var zmap sectorMapZ
	var objs []object
	var ok bool

	ymap, ok = sim.sectors[x]
	if !ok {
		ymap = make(sectorMapY)
		sim.sectors[x] = ymap
	}

	zmap, ok = ymap[y]
	if !ok {
		zmap = make(sectorMapZ)
		ymap[y] = zmap
	}

	objs, ok = zmap[z]
	if !ok {
		objs = make([]object, 0)
	}

	for _, other := range objs {
		distanceApart := obj.GetPosition().Sub(other.GetPosition()).Len()
		radii := obj.GetRadius() + other.GetRadius()
		if distanceApart < radii {
			logger.Debugln("Collision", obj.GetPosition(), other.GetPosition())
			sim.collisions = append(sim.collisions, collision{
				obj,
				other,
			})
		}
	}

	objs = append(objs, obj)
	zmap[z] = objs
}
