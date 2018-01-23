package src

// Fohristiwhirl, November 2017.
// A sort of alternative starter kit.

import (
	"fmt"
	"math"
	"sort"

	//	"github.com/hongshibao/go-kdtree"
)

// ------------------------------------------------------

type Entity interface {
	//	kdtree.Point
	Type() EntityType
	GetId() int
	GetX() float64
	GetY() float64
	GetRadius() float64
	SetRadius(r float64)
	GetSortDistance() float64
	SetSortDistance(dist float64)
	Angle(other Entity) float64
	Dist(other Entity) float64
	Dist2(other Entity) float64
	Alive() bool
	String() string
	OffsetPolar(dist, angle float64) Point
	OffsetPolarInt(dist, angle int) Point
	OffsetTowards(other Entity, dist float64) Point
	ClosestPoint(other Entity, minDist float64) Point
	AsRadius(r float64) Point
}

func EntIndex(ents []Entity, e Entity) int {
	for i, ent := range ents {
		if e == ent {
			return i
		}
	}
	return -1
}

func Furthest(points []Entity, e Entity) Entity {
	if len(points) < 1 {
		panic("")
	}
	return SortTo(points, e)[len(points)-1]
}

func Nearest(points []Entity, e Entity) Entity {
	if len(points) < 1 {
		panic("can't find the nearest of an empty list of entities!")
	}
	return SortTo(points, e)[0]
}

func SortTo(points []Entity, e Entity) []Entity {
	if len(points) < 1 {
		panic("")
	}
	newPoints := make([]Entity, len(points))
	for i, point := range points {
		newPoints[i] = point
	}
	sort.SliceStable(newPoints, func(i, j int) bool {
		return points[i].Dist(e) < points[j].Dist(e)
	})
	return newPoints
}

func PointsToEntities(ents []Point) []Entity {
	entities := make([]Entity, len(ents))
	for i, v := range ents {
		entities[i] = Entity(v)
	}
	return entities
}

func PlanetsToEntities(ents []*Planet) []Entity {
	entities := make([]Entity, len(ents))
	for i, v := range ents {
		entities[i] = Entity(v)
	}
	return entities
}

func ShipsToEntities(ents []*Ship) []Entity {
	entities := make([]Entity, len(ents))
	for i, v := range ents {
		entities[i] = Entity(v)
	}
	return entities
}

func EntitiesToShips(entities []Entity) []*Ship {
	ships := make([]*Ship, len(entities))
	for i, ent := range entities {
		ships[i] = ent.(*Ship)
	}
	return ships
}

func EntitiesToPlanets(entities []Entity) []*Planet {
	planets := make([]*Planet, len(entities))
	for i, ent := range entities {
		planets[i] = ent.(*Planet)
	}
	return planets
}

func EntitiesToPoints(entities []Entity) []Point {
	points := make([]Point, len(entities))
	for i, ent := range entities {
		points[i] = ent.(Point)
	}
	return points
}

func Filter(vs []Entity, f func(entity Entity) bool) []Entity {
	vsf := make([]Entity, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func EntitiesDist(a, b Entity) float64 {
	if a.Type() == NOTHING || b.Type() == NOTHING {
		panic("EntitiesDist() called with NOTHING entity")
	}
	return Dist(a.GetX(), a.GetY(), b.GetX(), b.GetY())
}

func EntitiesDist2(a, b Entity) float64 {
	if a.Type() == NOTHING || b.Type() == NOTHING {
		panic("EntitiesDist() called with NOTHING entity")
	}
	return Dist2(a.GetX(), a.GetY(), b.GetX(), b.GetY())
}

func EntitiesAngle(a, b Entity) float64 {
	if a.Type() == NOTHING || b.Type() == NOTHING {
		panic("EntitiesAngle() called with NOTHING entity")
	}
	return Angle(a.GetX(), a.GetY(), b.GetX(), b.GetY())
}

func EntitiesOffsetPolar(e Entity, distance float64, degrees float64) Point {
	var ret Point
	ret.X = e.GetX() + math.Cos(DegToRad(degrees))*distance
	ret.Y = e.GetY() + math.Sin(DegToRad(degrees))*distance
	return ret
}

func EntitiesOffsetTowards(e, other Entity, distance float64) Point {
	return e.OffsetPolar(distance, e.Angle(other))
}

func EntitiesClosestPoint(e, other Entity, minDist float64) Point {
	if minDist <= 0 {
		minDist = 3
	}
	radius := other.GetRadius() + minDist
	return other.OffsetTowards(e, radius)
}

// ------------------------------------------------------

type Planet struct {
	Id                int
	X                 float64
	Y                 float64
	HP                int
	Radius            float64
	DockingSpots      int
	CurrentProduction int
	Owned             bool
	Owner             int // Protocol will send 0 if not owned at all, but we "correct" this to -1
	DockedShips       int // The ships themselves can be accessed via game.dockMap[]
	DistanceTo        float64
	Value             float64
	ObjectiveMet      bool
	usedShips         map[int]bool
	SpawnPoint        Point
}

func (p Planet) CalcSpawnPoint(game Game) Point {
	best_location := p.AsRadius(0)
	best_distance := math.MaxFloat64
	max_delta := float64(SPAWN_RADIUS)
	for dx := -max_delta; dx <= max_delta; dx++ {
		for dy := -max_delta; dy <= max_delta; dy++ {
			offset_angle := math.Atan2(dy, dx)
			offset_x := dx + p.Radius*math.Cos(offset_angle)
			offset_y := dy + p.Radius*math.Sin(offset_angle)

			location := Point{X: p.GetX() + offset_x, Y: p.GetY() + offset_y}

			distance := location.Dist(game.Center())
			if distance < best_distance {
				best_distance = distance
				best_location = location
			}
		}
	}
	return best_location
}

func (p Planet) IsSafeExpandable(game Game) bool {
	if !((!p.Owned || p.Owner == game.pid) && p.OpenSpots() > 1 && len(game.GetAttackers(p, -1, -1)) < 1) {
		return false
	}
	//	game.Log("Checking nearest planet owner")
	ownedPs := Filter(PlanetsToEntities(game.AllPlanets()), func(e Entity) bool {
		return e.(*Planet).Owned
	})
	if len(ownedPs) <= 0 {
		return true
	}

	//	game.Log("Finding nearest planet")
	n := Nearest(ownedPs, p)

	//	game.Log("Returning owner check of nearest planet")
	return n.(*Planet).Owner == game.Pid() || n.(*Planet).Owner == -2 // || n.(*Planet).DockingSpots < p.DockingSpots
}

func (p Planet) OpenSpots() int {
	return p.DockingSpots - p.DockedShips
}

func (p Planet) IsFull() bool {
	return p.DockedShips >= p.DockingSpots
}

// ------------------------------------------------------

type Ship struct {
	Id              int
	Owner           int
	X               float64
	Y               float64
	Radius          float64
	HP              int
	DockedStatus    DockedStatus
	DockedPlanet    int
	DockingProgress int

	Birth        int // Turn this *Ship was first seen
	DistanceTo   float64
	Objectives   *[]Entity
	NextSpeed    int
	NextAngle    int
	LastSpeed    int
	LastAngle    int
	LastLocation Point
	IsKite       bool
	Combatants   []*Ship
	Original     **Ship
}

func (e *Ship) Projection() *Ship {
	if e.Original != nil {
		return e
	} else {
		clone := Ship(*e)
		clone.Original = &e
		clone.NextSpeed = 0
		clonePoint := e.OffsetPolar(float64(e.NextSpeed), float64(e.NextAngle))
		clone.X = clonePoint.GetX()
		clone.Y = clonePoint.GetY()
		return &clone
	}

}

func (e *Ship) Unprojection(game Game) *Ship {

	if e.Original != nil {
		return *e.Original
	} else {
		return e
	}
}

func (e *Ship) CanDock(p Planet) bool {
	if e.Alive() && p.Alive() && p.IsFull() == false && (p.Owned == false || p.Owner == e.Owner) {
		return e.Dist(p)-p.Radius < DOCKING_RADIUS+SHIP_RADIUS
	}
	return false
}

func (e *Ship) IsUndocked() bool {
	return e.DockedStatus == UNDOCKED
}

func (e *Ship) HasMovePlanned() bool {
	return e.NextSpeed > 0
}

// ------------------------------------------------------

type Point struct {
	X              float64
	Y              float64
	DistanceTo     float64
	Radius         float64
	ShipIDRef      int
	Speed, Degrees int // These are for choices, possible thrust commands attached to the result Point
}

func (e Point) IsSameLoc(other Point) bool {
	return e.Dist(other) < POINT_MATCH_TOLERANCE
}

type Nothing struct{}

// ------------------------------------------------------

// Interface satisfiers....

func (e *Ship) Type() EntityType   { return SHIP }
func (e Point) Type() EntityType   { return POINT }
func (e Planet) Type() EntityType  { return PLANET }
func (e Nothing) Type() EntityType { return NOTHING }

func (e *Ship) GetId() int   { return e.Id }
func (e Point) GetId() int   { return -1 }
func (e Planet) GetId() int  { return e.Id }
func (e Nothing) GetId() int { panic("GetId() called on NOTHING entity") }

func (e *Ship) GetX() float64   { return (*e).X }
func (e Point) GetX() float64   { return e.X }
func (e Planet) GetX() float64  { return e.X }
func (e Nothing) GetX() float64 { panic("GetX() called on NOTHING entity") }

func (e *Ship) GetY() float64   { return e.Y }
func (e Point) GetY() float64   { return e.Y }
func (e Planet) GetY() float64  { return e.Y }
func (e Nothing) GetY() float64 { panic("GetY() called on NOTHING entity") }

func (e *Ship) GetRadius() float64   { return e.Radius }
func (e Point) GetRadius() float64   { return e.Radius }
func (e Planet) GetRadius() float64  { return e.Radius }
func (e Nothing) GetRadius() float64 { return 0 }

func (e *Ship) SetRadius(r float64)   { e.Radius = r }
func (e Point) SetRadius(r float64)   { e.Radius = r }
func (e Planet) SetRadius(r float64)  { e.Radius = r }
func (e Nothing) SetRadius(r float64) { panic("SetRadius() call on NOTHING") }

func (e *Ship) Angle(other Entity) float64   { return EntitiesAngle(e, other) }
func (e Point) Angle(other Entity) float64   { return EntitiesAngle(e, other) }
func (e Planet) Angle(other Entity) float64  { return EntitiesAngle(e, other) }
func (e Nothing) Angle(other Entity) float64 { return EntitiesAngle(e, other) } // Will panic

func (e *Ship) Dist(other Entity) float64   { return EntitiesDist(e, other) }
func (e Point) Dist(other Entity) float64   { return EntitiesDist(e, other) }
func (e Planet) Dist(other Entity) float64  { return EntitiesDist(e, other) }
func (e Nothing) Dist(other Entity) float64 { return EntitiesDist(e, other) } // Will panic

func (e *Ship) Dist2(other Entity) float64   { return EntitiesDist2(e, other) }
func (e Point) Dist2(other Entity) float64   { return EntitiesDist2(e, other) }
func (e Planet) Dist2(other Entity) float64  { return EntitiesDist2(e, other) }
func (e Nothing) Dist2(other Entity) float64 { return EntitiesDist2(e, other) } // Will panic

func (e *Ship) OffsetPolar(dist, angle float64) Point   { return EntitiesOffsetPolar(e, dist, angle) }
func (e Point) OffsetPolar(dist, angle float64) Point   { return EntitiesOffsetPolar(e, dist, angle) }
func (e Planet) OffsetPolar(dist, angle float64) Point  { return EntitiesOffsetPolar(e, dist, angle) }
func (e Nothing) OffsetPolar(dist, angle float64) Point { return EntitiesOffsetPolar(e, dist, angle) }

func (e *Ship) OffsetPolarInt(dist, angle int) Point {
	return EntitiesOffsetPolar(e, float64(dist), float64(angle))
}
func (e Point) OffsetPolarInt(dist, angle int) Point {
	return EntitiesOffsetPolar(e, float64(dist), float64(angle))
}
func (e Planet) OffsetPolarInt(dist, angle int) Point {
	return EntitiesOffsetPolar(e, float64(dist), float64(angle))
}
func (e Nothing) OffsetPolarInt(dist, angle int) Point {
	return EntitiesOffsetPolar(e, float64(dist), float64(angle))
}

func (e *Ship) OffsetTowards(other Entity, dist float64) Point {
	return EntitiesOffsetTowards(e, other, dist)
}
func (e Point) OffsetTowards(other Entity, dist float64) Point {
	return EntitiesOffsetTowards(e, other, dist)
}
func (e Planet) OffsetTowards(other Entity, dist float64) Point {
	return EntitiesOffsetTowards(e, other, dist)
}
func (e Nothing) OffsetTowards(other Entity, dist float64) Point {
	return EntitiesOffsetTowards(e, other, dist)
}

func (e *Ship) GetSortDistance() float64   { return e.DistanceTo }
func (e Point) GetSortDistance() float64   { return e.DistanceTo }
func (e Planet) GetSortDistance() float64  { return e.DistanceTo }
func (e Nothing) GetSortDistance() float64 { return 0 } // Will panic

func (e *Ship) SetSortDistance(d float64)   { e.DistanceTo = d }
func (e Point) SetSortDistance(d float64)   { e.DistanceTo = d }
func (e Planet) SetSortDistance(d float64)  { e.DistanceTo = d }
func (e Nothing) SetSortDistance(d float64) { panic("SetSortDistance() called on NOTHING entity") } // Will panic

func (e *Ship) AsRadius(d float64) Point   { return Point{X: e.X, Y: e.Y, Radius: d} }
func (e Point) AsRadius(d float64) Point   { return Point{X: e.X, Y: e.Y, Radius: d} }
func (e Planet) AsRadius(d float64) Point  { return Point{X: e.X, Y: e.Y, Radius: d} }
func (e Nothing) AsRadius(d float64) Point { panic("Something trie dto get NOTHING as a radius") }

func (e *Ship) Alive() bool   { return e.HP > 0 }
func (e Point) Alive() bool   { return true }
func (e Planet) Alive() bool  { return e.HP > 0 }
func (e Nothing) Alive() bool { return false }

func (e *Ship) String() string { return fmt.Sprintf("Ship %d [%.4f,%.4f]-%v", e.Id, e.X, e.Y, e.Radius) }
func (e Point) String() string {
	return fmt.Sprintf("(%v)Point [%.4f,%.4f]-%v", e.ShipIDRef, e.X, e.Y, e.Radius)
}
func (e Planet) String() string {
	return fmt.Sprintf("Planet %d [%.4f,%.4f]-%v", e.Id, e.X, e.Y, e.Radius)
}
func (e Nothing) String() string { return "null entity" }

func (e *Ship) ClosestPoint(other Entity, minDist float64) Point {
	return EntitiesClosestPoint(e, other, minDist)
}
func (e Point) ClosestPoint(other Entity, minDist float64) Point {
	return EntitiesClosestPoint(e, other, minDist)
}
func (e Planet) ClosestPoint(other Entity, minDist float64) Point {
	return EntitiesClosestPoint(e, other, minDist)
}
func (e Nothing) ClosestPoint(other Entity, minDist float64) Point {
	return EntitiesClosestPoint(e, other, minDist)
}

//func (e *Ship) Dim() int { return 2 }
//func (e *Ship) GetValue(dim int) float64 {
//	if dim == 0 {
//		return e.GetX()
//	} else {
//		return e.GetY()
//	}
//}

//func (e *Ship) Distance(other kdtree.Point) float64 {
//	var ret float64
//	for i := 0; i < e.Dim(); i++ {
//		tmp := e.GetValue(i) - other.GetValue(i)
//		ret += tmp * tmp
//	}
//	return ret
//}

//func (e *Ship) PlaneDistance(val float64, dim int) float64 {
//	tmp := e.GetValue(dim) - val
//	return tmp * tmp
//}

//func (e Planet) Dim() int { return 2 }
//func (e Planet) GetValue(dim int) float64 {
//	if dim == 0 {
//		return e.GetX()
//	} else {
//		return e.GetY()
//	}
//}

//func (e Planet) Distance(other kdtree.Point) float64 {
//	var ret float64
//	for i := 0; i < e.Dim(); i++ {
//		tmp := e.GetValue(i) - other.GetValue(i)
//		ret += tmp * tmp
//	}
//	return ret
//}

//func (e Planet) PlaneDistance(val float64, dim int) float64 {
//	tmp := e.GetValue(dim) - val
//	return tmp * tmp
//}

//func (e Point) Dim() int { return 2 }
//func (e Point) GetValue(dim int) float64 {
//	if dim == 0 {
//		return e.GetX()
//	} else {
//		return e.GetY()
//	}
//}

//func (e Point) Distance(other kdtree.Point) float64 {
//	var ret float64
//	for i := 0; i < e.Dim(); i++ {
//		tmp := e.GetValue(i) - other.GetValue(i)
//		ret += tmp * tmp
//	}
//	return ret
//}

//func (e Point) PlaneDistance(val float64, dim int) float64 {
//	tmp := e.GetValue(dim) - val
//	return tmp * tmp
//}

//func (e Nothing) Dim() int { return 2 }
//func (e Nothing) GetValue(dim int) float64 {
//	if dim == 0 {
//		return e.GetX()
//	} else {
//		return e.GetY()
//	}
//}

//func (e Nothing) Distance(other kdtree.Point) float64 {
//	var ret float64
//	for i := 0; i < e.Dim(); i++ {
//		tmp := e.GetValue(i) - other.GetValue(i)
//		ret += tmp * tmp
//	}
//	return ret
//}

//func (e Nothing) PlaneDistance(val float64, dim int) float64 {
//	tmp := e.GetValue(dim) - val
//	return tmp * tmp
//}
