package src

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

func make_circle(points []Entity) Point {
	// Convert to float and randomize order
	shuffledPoints := make([]Point, len(points))
	for i, position := range rand.Perm(len(points)) {
		shuffledPoints[i] = points[position].AsRadius(0)
	}

	var c *Point
	for i, p := range shuffledPoints {
		if c == nil || !WithinDistance(c, p, c.GetRadius()) {
			c = _make_circle_one_point(shuffledPoints[:i+1], p)
		}
	}

	return *c
}

func make_circumcircle(p0, p1, p2 Point) *Point {
	// Mathematical algorithm from Wikipedia: Circumscribed circle
	ax, ay := p0.X, p0.Y
	bx, by := p1.X, p1.Y
	cx, cy := p2.X, p2.Y
	ox := (MinFloatSlice([]float64{ax, bx, cx}) + MaxFloatSlice([]float64{ax, bx, cx})) / 2.0
	oy := (MinFloatSlice([]float64{ay, by, cy}) + MaxFloatSlice([]float64{ay, by, cy})) / 2.0
	ax -= ox
	ay -= oy
	bx -= ox
	by -= oy
	cx -= ox
	cy -= oy
	d := (ax*(by-cy) + bx*(cy-ay) + cx*(ay-by)) * 2.0
	if d == 0.0 {
		return nil
	}
	x := ox + ((ax*ax+ay*ay)*(by-cy)+(bx*bx+by*by)*(cy-ay)+(cx*cx+cy*cy)*(ay-by))/d
	y := oy + ((ax*ax+ay*ay)*(cx-bx)+(bx*bx+by*by)*(ax-cx)+(cx*cx+cy*cy)*(bx-ax))/d
	r := Point{X: x, Y: y}
	ra := r.Dist(p0)
	rb := r.Dist(p1)
	rc := r.Dist(p2)
	c := Point{X: x, Y: y, Radius: MaxFloatSlice([]float64{ra, rb, rc})}
	return &c
}

func make_diameter(p0, p1 Point) Point {
	cx := (p0.X + p1.X) / 2.0
	cy := (p0.Y + p1.Y) / 2.0
	cp := Point{X: cx, Y: cy}
	r0 := cp.Dist(p0)
	r1 := cp.Dist(p1)
	c := Point{X: cx, Y: cy, Radius: MaxFloat(r0, r1)}
	return c
}

// One boundary point known
func _make_circle_one_point(points []Point, p Point) *Point {
	c := p.AsRadius(0)
	for i, q := range points {
		if !WithinDistance(c, q, c.GetRadius()) {
			if c.GetRadius() == 0.0 {
				c = make_diameter(p, q)
			} else {
				c = _make_circle_two_points(points[:i+1], p, q)
			}
		}
	}
	return &c
}

// Two boundary points known
func _make_circle_two_points(points []Point, p, q Point) Point {
	circ := make_diameter(p, q)
	var left *Point
	var right *Point
	px, py := p.X, p.Y
	qx, qy := q.X, q.Y

	// For each point not in the two-point circle
	for _, r := range points {
		if WithinDistance(circ, r, circ.GetRadius()) {
			continue
		}

		// Form a circumcircle and classify it on left or right side
		cross := _cross_product(px, py, qx, qy, r.X, r.Y)
		c := make_circumcircle(p, q, r)
		if c == nil {
			continue
		} else if cross > 0.0 && (left == nil || _cross_product(px, py, qx, qy, c.X, c.Y) > _cross_product(px, py, qx, qy, left.X, left.Y)) {
			left = c
		} else if cross < 0.0 && (right == nil || _cross_product(px, py, qx, qy, c.X, c.Y) < _cross_product(px, py, qx, qy, right.X, right.Y)) {
			right = c
		}
	}

	// Select which circle to return
	if left == nil && right == nil {
		return circ
	} else if left == nil {
		return *right
	} else if right == nil {
		return *left
	} else if left.GetRadius() <= right.GetRadius() {
		return *left
	} else {
		return *right
	}
}

func _cross_product(x0, y0, x1, y1, x2, y2 float64) float64 {
	return (x1-x0)*(y2-y0) - (y1-y0)*(x2-x0)
}

func GenCircle(game Game, points []Entity) Point {
	var circ Point
	if len(points) <= 2 {
		circ = make_diameter(points[0].AsRadius(0), points[1].AsRadius(0))
	} else if len(points) <= 3 {
		game.Log("Making circle from %v", points)
		maybeCirc := make_circumcircle(points[0].AsRadius(0), points[1].AsRadius(0), points[2].AsRadius(0))
		if maybeCirc == nil {
			maybeCirc = make_circumcircle(points[0].OffsetPolar(.001, 0), points[1].AsRadius(0), points[2].AsRadius(0))
		}
		circ = *(maybeCirc)
	}
	return circ
}

func arc_share(game Game, points []Entity) (Point, []Arc) {
	if len(points) > 3 || len(points) < 1 {
		panic(fmt.Sprintf("Cannot take find the arc share of %v", points))
	}

	if len(points) <= 1 {
		return points[0].AsRadius(0), []Arc{Arc{lower: 0, upper: 360, obstacle: points[0]}}
	}

	game.Log("Finding arc_share of enemies")
	circ := GenCircle(game, points)

	colors := []Color{RED, BLUE, GREEN, ORANGE}
	// sort them from lowest to highest so we can construct the arcs correctly
	sort.Slice(points, func(i, j int) bool { return circ.Angle(points[i]) < circ.Angle(points[j]) })

	arcs := make([]Arc, len(points))
	lastMid := 0.0
	nextMid := 0.0
	points = append(points, points[0])
	for i := 0; i < len(points)-1; i++ {
		nextMid = circ.Angle(points[i].OffsetTowards(points[i+1], points[i].Dist(points[i+1])/2.0))

		// if the angle diff is over 180
		if circ.Angle(points[i+1])-circ.Angle(points[i]) > 180 {
			nextMid += 180.0
			nextMid = math.Mod(nextMid, 360)
			// if this is the last one and the wrapped angle is over 180
		} else if i == len(points)-2 && 360-circ.Angle(points[i])+circ.Angle(points[i+1]) > 180 {
			nextMid += 180.0
			nextMid = math.Mod(nextMid, 360)
		}
		arc := Arc{lower: lastMid, upper: nextMid, obstacle: points[i]}
		arcs[i] = arc
		lastMid = nextMid
	}

	// a little clean up here to glue the very last arcs section to the first
	arcs[0] = Arc{lower: lastMid, upper: arcs[0].upper, obstacle: points[0]}

	i := 0
	for _, arc := range arcs {
		game.Log("Drawing %v in %v", arc, colors[i])
		game.DrawArc(circ, circ.GetRadius(), arc.lower, arc.upper, colors[i], 2, NAV_DISPLAY)
		i++
	}

	return circ, arcs
}

func InArcArc(game Game, arc, testArc Arc) bool {
	return InArc(game, arc, testArc.upper) && InArc(game, arc, testArc.lower)
}

func InArcCircle(game Game, circle Entity, arc Arc, testEnt Entity) bool {
	dist := circle.Dist(testEnt)
	base := circle.Angle(testEnt)

	theta := RadToDeg(math.Asin(testEnt.GetRadius() / dist))

	a := Arc{lower: base - theta, upper: base + theta, obstacle: testEnt}

	return InArcArc(game, arc, a)
}

func MakeOrbShip(game Game, myShips []*Ship, orbSize float64) *Ship {
	if orbSize < 0 {
		orbSize = 3 * SHIP_RADIUS
	}

	c := make_circle(ShipsToEntities(myShips))
	if c.GetRadius() < orbSize {
		c.Radius = orbSize
	}
	c.Radius += SHIP_RADIUS
	return &Ship{X: c.X, Y: c.Y, Owner: game.Pid(), Radius: c.Radius}
}

func Offset90(game Game, focus, against Entity, dist float64) []Point {
	angle := focus.Angle(against)

	options := []Point{focus.OffsetPolar(dist, angle-90.0), focus.OffsetPolar(dist, angle+90.0)}
	return options
}
