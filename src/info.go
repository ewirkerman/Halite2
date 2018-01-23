package src

// Fohristiwhirl, November 2017.
// A sort of alternative starter kit.

import (
	//	"math"
	"sort"

	//	"github.com/hongshibao/go-kdtree"
)

// ----------------------------------------------

func (g *Game) GetShip(sid int) (*Ship, bool) {
	ret, ok := g.shipMap[sid]
	return ret, ok
}

func (g *Game) GetShips(sids []int) []*Ship {
	ships := make([]*Ship, len(sids))
	for i, sid := range sids {
		s, _ := g.GetShip(sid)
		ships[i] = s
	}
	return ships
}

func (g *Game) GetPlanet(plid int) (*Planet, bool) {
	ret, ok := g.planetMap[plid]
	return ret, ok
}

// ----------------------------------------------

func (g *Game) AllShips() []*Ship {
	ret := make([]*Ship, len(g.all_ships_cache))
	copy(ret, g.all_ships_cache)
	return ret
}

func (g *Game) AllPlanets() []*Planet {
	ret := make([]*Planet, len(g.all_planets_cache))
	copy(ret, g.all_planets_cache)
	return ret
}

func (g *Game) PlanetsOwnedBy(pid int) []*Planet {
	planets := g.AllPlanets()
	var ret []*Planet
	for _, p := range planets {
		if p.Owner == pid {
			ret = append(ret, p)
		}
	}
	return ret
}

func (g *Game) WeakestEnemy() int {
	weakest := -1
	ships := 100000
	for pid := 0; pid < g.InitialPlayers(); pid++ {
		if len(g.PlanetsOwnedBy(pid)) < 1 {
			continue
		}
		if pid == g.pid {
			continue
		}
		if len(g.ShipsOwnedBy(pid)) < ships {
			weakest = pid
			ships = len(g.ShipsOwnedBy(pid))
		}
	}
	return weakest
}

func (g *Game) EnemyPlanetsOf(pid int) []*Planet {
	planets := g.AllPlanets()
	var ret []*Planet
	for _, p := range planets {
		if p.Owned && p.Owner != pid && p.Owner != -2 {
			ret = append(ret, p)
		}
	}
	return ret
}

// ----------------------------------------------

func (g *Game) ShipsOwnedBy(pid int) []*Ship {
	ret := make([]*Ship, len(g.playershipMap[pid]))
	copy(ret, g.playershipMap[pid])
	return ret
}

func (g *Game) ShipsUndockedOwnedBy(pid int) []*Ship {
	var ships []*Ship
	for _, ship := range g.ShipsOwnedBy(pid) {
		if ship.IsUndocked() {
			ships = append(ships, ship)
		}
	}
	return ships
}

func (g *Game) ShipsDockedOwnedBy(pid int) []*Ship {
	var ships []*Ship
	for _, ship := range g.ShipsOwnedBy(pid) {
		if !ship.IsUndocked() {
			ships = append(ships, ship)
		}
	}
	return ships
}

func (g *Game) MyShips() []*Ship {
	return g.ShipsOwnedBy(g.pid)
}

func (g *Game) MyShipsUndocked() []*Ship {
	return g.ShipsUndockedOwnedBy(g.pid)
}

func (g *Game) MyShipsDocked() []*Ship {
	return g.ShipsDockedOwnedBy(g.pid)
}

func (g *Game) MyPlanets() []*Planet {
	return g.PlanetsOwnedBy(g.pid)
}

func (g *Game) EnemyShipsUndocked() []*Ship {
	var ships []*Ship
	for _, ship := range g.enemy_ships_cache {
		if ship.IsUndocked() {
			ships = append(ships, ship)
		}
	}
	return ships
}

func (g *Game) EnemyShipsUndockedOf(pid int) []*Ship {
	if pid == g.pid {
		return g.EnemyShipsUndocked()
	}
	var ships []*Ship
	for p := 0; p < g.currentPlayers-1; p++ {
		if p == pid {
			continue
		}
		for _, ship := range g.ShipsOwnedBy(p) {
			if ship.IsUndocked() {
				ships = append(ships, ship)
			}
		}
	}
	return ships
}

func (g *Game) EnemyShips() []*Ship {
	ret := make([]*Ship, len(g.enemy_ships_cache))
	copy(ret, g.enemy_ships_cache)
	return ret
}

// ----------------------------------------------

func (g *Game) MyNewShipIDs() []int { // My ships born this turn.
	var ret []int
	for sid, _ := range g.shipMap {
		ship := g.shipMap[sid]
		if ship.Birth == g.turn && ship.Owner == g.pid {
			ret = append(ret, ship.Id)
		}
	}
	sort.Slice(ret, func(a, b int) bool {
		return ret[a] < ret[b]
	})
	return ret
}

func (g *Game) ShipsDockedAt(planet Planet) []*Ship {
	ret := make([]*Ship, len(g.dockMap[planet.Id]))
	copy(ret, g.dockMap[planet.Id])
	return ret
}

func (g *Game) RawWorld() string {
	return g.raw
}

func (g Game) MeanPoint(ents []Entity) Point {
	X := 0.0
	Y := 0.0
	for _, ent := range ents {
		X += ent.GetX()
		Y += ent.GetY()
	}
	mean := Point{X: X / float64(len(ents)), Y: Y / float64(len(ents))}
	return mean
}

func (g Game) Center() Point {
	return g.MeanPoint([]Entity{Point{X: 0, Y: 0}, Point{X: float64(g.Width()), Y: float64(g.Height())}})
}

// NearestEntities orders all points based on their proximity
// to a given *Ship from nearest for farthest

// Depending
//func (g *Game) NearestEntitiesKDT(ent Entity, population []Entity, maxCount int, maxDist float64) []Entity {
//	if maxCount < 1 {
//		maxCount = len(population)
//	}

//	if maxCount < 1 {
//		return []Entity{}
//	}

//	g.Log("pop size = %v", len(population))
//	KDT_ents := EntitiesToKDTPoints(population)
//	g.Log("pop size = %v", len(KDT_ents))
//	tree := kdtree.NewKDTree(KDT_ents)
//	g.Log("finding %v neighbors in a tree of %v...", maxCount, tree)

//	neighbors := tree.KNN(kdtree.Point(ent), maxCount)
//	g.Log("neighbors size = %v", len(neighbors))

//	if maxDist > 0 {
//		for i, neighbor := range neighbors {
//			if !WithinDistance(ent, neighbor.(Entity), maxDist) {
//				neighbors = neighbors[:int(math.Max(0, float64(i)-1.0))]
//			}
//		}
//	}
//	KDTPoints := KDTPointsToEntities(neighbors)
//	g.Log("KDTPoints size = %v", len(KDTPoints))
//	return KDTPoints
//}

func (g *Game) NearestEntities(ent Entity, population []Entity, maxCount int, maxDist float64) []Entity {

	//g.Log("OrigPop: %v", population)
	var newPop []Entity
	for _, pop_ent := range population {
		//g.Log("Distance from ent %v to %v: %v vs %v", ent, pop_ent, EntitiesDist(ent, pop_ent), maxDist)
		if pop_ent != ent && (maxDist <= 0 || EntitiesDist(ent, pop_ent) < maxDist) {
			//			g.Log("Including %v", pop_ent)
			newPop = append(newPop, pop_ent)
		}
	}

	if len(newPop) > 1 {
		//		g.Log("Including %v", newPop)
		sort.SliceStable(newPop, func(i, j int) bool { return EntitiesDist(ent, newPop[i]) < EntitiesDist(ent, newPop[j]) })
	}

	//sort.Sort(byDist(newPop))
	//g.Log("NewPop: %v", newPop)
	if maxCount > 0 && len(newPop) > maxCount {
		newPop = newPop[:maxCount]
	}

	return newPop
}

func (g *Game) ShipsProducedBy(i int) int {
	return len(g.playerShipsetMap[i])
}

func (g *Game) CurrentStandings() []int {
	pids := make([]int, g.InitialPlayers())
	for i := 0; i < len(pids); i++ {
		pids[i] = i
	}
	sort.Slice(pids, func(i, j int) bool {
		return len(g.ShipsOwnedBy(pids[i])) <= 0 && g.ShipsProducedBy(pids[i]) < g.ShipsProducedBy(pids[j])
	})
	return pids
}

func (g *Game) GetAttackers(entity Entity, turnsAway int, maxDist float64) []*Ship {
	if maxDist < 0 {
		if turnsAway < 1 {
			turnsAway = 5
		}
		maxDist = MAX_SPEED * float64(turnsAway)
	}

	maxDist += entity.GetRadius()
	return EntitiesToShips(g.NearestEntities(entity, ShipsToEntities(g.EnemyShipsUndocked()), -1, maxDist))
}

func WithinDistanceAll(projection Entity, ships []Entity, radius float64) bool {
	for _, ship := range ships {
		if !WithinDistance(ship, projection, radius) {
			return false
		}
	}
	return true
}

func WithinDistanceAny(projection Entity, ships []Entity, radius float64) bool {
	for _, ship := range ships {
		if projection != ship && WithinDistance(ship, projection, radius) {
			return true
		}
	}
	return false
}

func WithinDistance(projection, ship Entity, radius float64) bool {
	return ship.Dist(projection) <= radius
}

type byDist []Entity

func (a byDist) Len() int           { return len(a) }
func (a byDist) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byDist) Less(i, j int) bool { return a[i].GetSortDistance() < a[j].GetSortDistance() }

// By is the type of a "less" function that defines the ordering of its Planet arguments.
type By func(p1, p2 *Entity) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(entities []Entity) {
	ps := &sorter{
		entities: entities,
		by:       by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// sorter joins a By function and a slice of Planets to be sorted.
type sorter struct {
	entities []Entity
	by       func(p1, p2 *Entity) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *sorter) Len() int           { return len(s.entities) }
func (s *sorter) Swap(i, j int)      { s.entities[i], s.entities[j] = s.entities[j], s.entities[i] }
func (s *sorter) Less(i, j int) bool { return s.by(&s.entities[i], &s.entities[j]) }

// By is the type of a "less" function that defines the ordering of its Planet arguments.
type PointsBy func(p1, p2 *Point) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by PointsBy) Sort(points []Point) {
	ps := &pointsSorter{
		points: points,
		by:     by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// sorter joins a By function and a slice of Planets to be sorted.
type pointsSorter struct {
	points []Point
	by     func(p1, p2 *Point) bool // Closure used in the Less method.
}

func (s *pointsSorter) Len() int      { return len(s.points) }
func (s *pointsSorter) Swap(i, j int) { s.points[i], s.points[j] = s.points[j], s.points[i] }

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *pointsSorter) Less(i, j int) bool { return s.by(&s.points[i], &s.points[j]) }
