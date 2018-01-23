package src

// Fohristiwhirl, November 2017.
// A sort of alternative starter kit.

import (
	"math"
)

func hasBit(n int, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
}

func Projection(x1, y1, distance, degrees float64) (x2, y2 float64) {

	// Given a coordinate, a distance and an Angle, find a new coordinate.

	if distance == 0 {
		return x1, y1
	}

	radians := DegToRad(float64(degrees))

	x2 = distance*math.Cos(radians) + x1
	y2 = distance*math.Sin(radians) + y1

	return x2, y2
}

func Angle(x1, y1, x2, y2 float64) float64 {

	rad := math.Atan2(y2-y1, x2-x1)
	deg := RadToDeg(rad)

	//deg_int := Round(deg)

	for deg < 0 {
		deg += 360
	}

	return math.Mod(deg, 360.0)
}

func DegToRad(d float64) float64 {
	return d / 180 * math.Pi
}

func RadToDeg(r float64) float64 {
	return r / math.Pi * 180
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func MaxFloatSlice(floats []float64) float64 {
	if len(floats) < 1 {
		panic("Can't max an empty slice")
	}
	max := floats[0]
	for _, float := range floats[1:] {
		max = MaxFloat(max, float)
	}
	return max
}

func MinFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func MinFloatSlice(floats []float64) float64 {
	if len(floats) < 1 {
		panic("Can't min an empty slice")
	}
	min := floats[0]
	for _, float := range floats[1:] {
		min = MinFloat(min, float)
	}
	return min
}

func Round(n float64) int {
	return int(math.Floor(n + 0.5))
}

func RoundToFloat(n float64) float64 {
	return math.Floor(n + 0.5)
}

func Dist2(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx + dy*dy
}

func Dist(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func ModInt(x, y int) int {
	return int(math.Mod(float64(x), float64(y)))
}

func SetOfShips(slice []*Ship) map[int]bool {
	set := make(map[int]bool)
	for _, ship := range slice {
		set[ship.GetId()] = true
	}
	return set
}

func ListOfShips(game Game, m map[int]bool) []*Ship {
	list := []*Ship{}
	for k, _ := range m {
		s, _ := game.GetShip(k)
		list = append(list, s)
	}
	return list
}

type ListSet struct {
	slice []int
	m     map[int]int
}

//
//func NewListSet(size int) ListSet {
//	return ListSet{slice: make([]int, size), m: map[int]int{}}
//}
//
//func (l *ListSet) Add(id int) {
//	i := len(l.slice)
//	l.slice = append(l.slice, id)
//	if _, ok := l.m[id]; !ok { panic("Unable to add to ListSet: key already exists") }
//	l.m[id] = i
//}
//
//func (l *ListSet) Populate(ints []int) {
//	for _, id := range ints { l.Add(id) }
//}
//
//func (l *ListSet) Contains(ship Ship) bool {
//	_, ok := l.m[ship.GetId()]
//	return ok
//}
//
//func (l *ListSet) Remove(id int) {
//	if i, ok := l.m[id]; !ok {
//		panic("Unable to remove from ListSet: key does not exists")
//	} else {
//		l.slice = append(l.slice[:i], l.slice[i+1:]...)
//		delete(l.m, id)
//	}
//}
