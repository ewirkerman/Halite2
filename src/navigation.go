package src

import (
	"fmt"
	"math"
	"sort"
)

func NavigateGeo(ship *Ship, target Entity, game Game, enemyRange float64, speed_direct, angle_direct int) Command {
	//	panic("Rewrite this to get rid of comets?")

	if speed_direct > 0 {
		//		game.Log("Navigating %v at speed %v and angle %v", ship, speed_direct, angle_direct)
		//game.Log("These should really be put into a collision TEST method, rather than a navigate method")
		game.DrawLine(ship, ship.OffsetPolar(float64(speed_direct), float64(angle_direct)), BLACK, 1, NAV_DISPLAY)
	} else {
		game.Log("Navigating %v to target %v", ship, target)
		game.DrawLine(ship, target, BLACK, 1, NAV_DISPLAY)
	}
	//game.Log("Elapsed time since turn start: %v", time.Since(game.ParseTime()))

	distanceToTarget := ship.Dist(target) + POINT_MATCH_TOLERANCE

	speed := MAX_SPEED
	if game.turn < 1 && game.CurrentPlayers() > 2 && ALLY_BEE_DANCE {
		if ship.GetId() == game.pid*3+ModInt(game.pid, 3) {
			speed = 6
		}
	}
	//game.Log("speed %v vs. dist to target %v", speed, distanceToTarget)
	maybeKiteAttackers := game.GetAttackers(ship, -1, WEAPON_RADIUS+MAX_SPEED)
	//game.Log("If this navigator is a kite, then his attackers are %v", maybeKiteAttackers)

	if !ship.IsKite || len(maybeKiteAttackers) < 1 {
		speedFloat := math.Floor(float64(speed) + POINT_MATCH_TOLERANCE)
		if distanceToTarget < speedFloat {
			speed = Min(Round(distanceToTarget), MAX_SPEED)
			game.Log("Reduced speed to %v based on a dist to target of %v", speed, distanceToTarget)
		}
		if speed_direct > 0 {
			speed = speed_direct
		} else {
			speed = int(speed)
		}
	} else {
		//game.Log("Threatened Kite Moves at full speed")
	}

	//game.Log("Navigating with speed: %v", speed)
	var targetAngle float64
	if angle_direct > -1 {
		targetAngle = float64(angle_direct)
	} else {
		targetAngle = float64(int(ship.Angle(target)))
	}

	obstacles := PlanetsToEntities(game.AllPlanets())
	maxPlanetSize := 0.0
	for _, planet := range obstacles {
		maxPlanetSize = MaxFloat(maxPlanetSize, planet.GetRadius())
	}
	obstacles = append(obstacles, ShipsToEntities(game.AllShips())...)
	obstacles = game.NearestEntities(ship, obstacles, -1, float64(MAX_SPEED)+SHIP_RADIUS+maxPlanetSize+BUFFER_TOLERANCE)

	//	game.LogEach("Obstacles for "+ship.String()+":", obstacles)
	arcs := GenArcs(game, ship, target, obstacles, speed, enemyRange)
	//	game.LogEach("Arcs for "+ship.String()+":", arcs)
	//	 game.Log("Found the following arcs for %v: %v", ship, arcs)
	min, max, ok := GenClosestAngles(game, arcs, targetAngle, true)

	//	game.Log("Found the following closest angles for %v and target %v: %v, %v", ship, targetAngle, min, max)
	angle := GetBestAngle(min, max, targetAngle)
	//	game.Log("Found the BEST angle for %v: %v", ship, angle)

	command := Command{Ship: ship, OrderType: THRUST, Speed: speed, Angle: angle}
	if !ok {
		command.Speed = 0
	}
	game.Log("Produced command: %v", command.String())
	return command
}

func NavigateWithoutShips(ship *Ship, target Entity, game Game) Command {
	game.DrawLine(ship, target, BLACK, 1, NAV_DISPLAY)
	//game.Log("Elapsed time since turn start: %v", time.Since(game.ParseTime()))

	distanceToTarget := ship.Dist(target) + POINT_MATCH_TOLERANCE

	speed := MAX_SPEED

	if distanceToTarget < float64(speed) {
		speed = Min(Round(distanceToTarget), MAX_SPEED)
		//game.Log("Reduced speed to %v based on a dist to target of %v", speed, distanceToTarget)
	}
	targetAngle := float64(int(ship.Angle(target)))

	obstacles := PlanetsToEntities(game.AllPlanets())
	maxPlanetSize := 0.0
	for _, planet := range obstacles {
		maxPlanetSize = MaxFloat(maxPlanetSize, planet.GetRadius())
	}
	obstacles = game.NearestEntities(ship, obstacles, -1, float64(MAX_SPEED)+SHIP_RADIUS+maxPlanetSize+BUFFER_TOLERANCE)

	//	game.LogEach("Obstacles for "+ship.String()+":", obstacles)
	arcs := GenArcs(game, ship, target, obstacles, speed, SHIP_RADIUS)
	//	game.LogEach("Arcs for "+ship.String()+":", arcs)
	// game.Log("Found the following arcs for %v: %v", ship, arcs)
	min, max, ok := GenClosestAngles(game, arcs, targetAngle, true)

	// game.Log("Found the following closest angles for %v and target %v: %v, %v", ship, targetAngle, min, max)
	angle := GetBestAngle(min, max, targetAngle)
	// game.Log("Found the BEST angle for %v: %v", ship, angle)

	cmd := Command{Ship: ship, OrderType: THRUST, Speed: speed, Angle: angle}
	if !ok {
		cmd.Speed = 0
	}
	//game.Log("Produced command: %v", command.String())
	return cmd
}

func NavigateAsOrb(game Game, myShips []*Ship, target, orientAgainst Entity, orbSize float64, cmd *Command) {

	game.DrawLine(cmd.Ship, target, LIGHT_GRAY, 2, NAV_DISPLAY)
	game.DrawEntity(target.AsRadius(SHIP_RADIUS), RED, 4, NAV_DISPLAY)
	game.Log("Navigated orb with cmd %v to result %v", *cmd, target)
	//	game.DrawEntity(result.AsRadius(cShip.GetRadius()), BLACK, 2, NAV_DISPLAY)

	sort.Slice(myShips, func(i, j int) bool { return myShips[i].Dist(target) < myShips[j].Dist(target) })

	spacing := SHIP_RADIUS * 2.5

leadLoop:
	for !WithinDistance(target, myShips[0], .99999) {
		nextSpots := append(Offset90(game, target, orientAgainst, spacing), target.AsRadius(0))
		for _, ship := range myShips {
			game.Unthrust(ship)
		}

		for _, wingman := range myShips {
			//			nextSpots := PointsToEntities(Offset90(game, nextSpot, orientAgainst, spacing))
			game.Log("nextSpots: %v", nextSpots)
			backSpots := make([]Point, len(nextSpots))
			for i, spot := range nextSpots {
				backSpots[i] = spot.OffsetPolar(float64(-(*cmd).Speed), float64((*cmd).Angle))
			}
			game.Log("backSpots: %v", backSpots)
			backSpot := Nearest(PointsToEntities(backSpots), wingman)
			i := EntIndex(PointsToEntities(backSpots), backSpot)
			game.Log("Selected backspot: %v", i)
			nextSpot := nextSpots[i]
			game.Log("Selected nextSpot: %v", nextSpot)
			nextSpots[i], nextSpots[len(nextSpots)-1] = nextSpots[len(nextSpots)-1], nextSpots[i]
			nextSpots = nextSpots[:len(nextSpots)-1]

			if !WithinDistance(nextSpot, wingman, MAX_SPEED+.25) {
				game.Log("Old target: %v", target)
				target = target.OffsetTowards(myShips[0], 1)
				cmd.Speed -= 1
				game.Log("Rolled target: %v", target)
				game.Log("Unrolling because %v to %v: %v", nextSpot, wingman, nextSpot.Dist(wingman))
				continue leadLoop
			}
			game.DrawLine(backSpot, nextSpot, LIGHT_GREEN, 2, NAV_DISPLAY)
			game.DrawEntity(backSpot.AsRadius(SHIP_RADIUS), LIGHT_GREEN, 2, NAV_DISPLAY)
			game.DrawEntity(nextSpot.AsRadius(SHIP_RADIUS), LIGHT_GREEN, 0, NAV_DISPLAY)
			cmd := Navigate(wingman, nextSpot, game, SHIP_RADIUS, -1, -1)

			//			diffAngle := nextSpot.Angle(cmd.Result())
			//			diffSpeed := nextSpot.Dist(cmd.Result())

			//			for i, spot := range nextSpots {
			//				nextSpots[i] = spot.OffsetPolar(diffSpeed, diffAngle)
			//				game.Log("Adjusted next spot %v to %v", spot, nextSpots[i])
			//			}

			game.ThrustCommand(cmd)
		}
		break
	}

}

func IntersectsSegmentCircle(game Game, start, end, circle Entity, radius float64) bool {
	closestPoint, _ := GetIntersectionT(start, end, circle)
	//	game.Log("Distance between segment and circle is %v (vs radius %v)", circle.Dist(closestPoint), radius)
	return circle.Dist(closestPoint) <= radius
}

func ValidateCommands(game Game, commands []Command) bool {
	points := make([]Point, len(commands))
	for i, command := range commands {
		points[i] = command.Result().AsRadius(command.Ship.GetRadius())
		points[i].ShipIDRef = command.Ship.GetId()
	}
	return ValidateMoves(game, points)
}

func ValidateMoves(game Game, points []Point) bool {
	// Check if they collide with any planets or ships that are moving already
	for _, move := range points {
		ship, _ := game.GetShip(move.ShipIDRef)
		obstacles := game.NearestEntities(ship, PlanetsToEntities(game.AllPlanets()), -1, -1)
		for _, ship := range game.AllShips() {
			inMoveSet := false
			for _, points := range points {
				if ship.Id == points.ShipIDRef {
					inMoveSet = true
					break
				}
			}
			if !inMoveSet {
				obstacles = append(obstacles, ship)
			}
		}
		for _, first := range obstacles {
			firstStart := first
			secondStart := ship
			radius := 1000.0
			var secondEnd Point
			if _, ok := first.(*Planet); ok {
				radius = first.GetRadius() + SHIP_RADIUS
				secondEnd = move
			} else if shipObs, ok := first.(*Ship); ok {
				radius = 2 * SHIP_RADIUS
				secondEnd = move.OffsetPolar(float64(-shipObs.NextSpeed), float64(shipObs.NextAngle))
			} else {
				panic("Unexpected Type Error!")
			}
			//			game.Log("firstStart: %v", firstStart)
			//			game.Log("secondStart: %v", secondStart)
			//			game.Log("secondEnd: %v", secondEnd)
			if IntersectsSegmentCircle(game, secondStart, secondEnd, firstStart, radius) {
				// game.Log("Ship %v to move %v collides with decided obstacle %v (r=%v)", ship, move, first, radius)
				return false
			}
		}
	}

	// Need to make sure these don't collidge with anything that ISN"T part of this command set also.
	// game.Log("Checking consistency of moves %v", points)
	for i, first := range points {
		// game.Log("Checking move %v (%v) against remaining options %v", i, first, points[i+1:])
		for _, second := range points[i+1:] {
			// game.Log("Checking move consistency of moving to %v and to %v at the same time", first, second)
			firstStart := first.OffsetPolar(float64(-first.Speed), float64(first.Degrees))
			secondStart := second.OffsetPolar(float64(-second.Speed), float64(second.Degrees))
			secondEnd := second.OffsetPolar(float64(-first.Speed), float64(first.Degrees))
			// game.Log("firstStart: %v", firstStart)
			// game.Log("secondStart: %v", secondStart)
			// game.Log("secondEnd: %v", secondEnd)
			if IntersectsSegmentCircle(game, secondStart, secondEnd, firstStart, 2*SHIP_RADIUS) {
				// game.Log("Ship %v to move %v collides with firstStart %v", secondStart, secondEnd, firstStart)
				return false
			}
		}
	}

	// game.Log("Moveset is consistent: %v", points)
	return true
}

func WillCollideNeighbors(game Game, ship *Ship) bool {
	neighbors := game.NearestEntities(ship, ShipsToEntities(game.MyShips()), -1, 2*MAX_SPEED+2*SHIP_RADIUS+BUFFER_TOLERANCE)
	for _, n := range neighbors {
		if WillCollide(game, ship, n) {
			return true
		}
	}
	return false
}

func WillCollideSlice(game Game, ship *Ship, neighbors []Entity) bool {
	for _, n := range neighbors {
		if WillCollide(game, ship, n) {
			return true
		}
	}
	return false
}

func WillCollide(game Game, ship *Ship, obstacle Entity) bool {
	//	game.Log("Collision - Test ship: %v", ship)
	start := obstacle
	end := obstacle
	if s, ok := obstacle.(*Ship); ok {
		end = s.Projection()
	}

	//	game.Log("Collision - obstacle start: %v", start)
	//	game.Log("Collision - obstacle end: %v", end)

	end = end.OffsetPolar(float64(-ship.NextSpeed), float64(ship.NextAngle))

	//	game.Log("Collision - obstacle effective start: %v", start)
	//	game.Log("Collision - obstacle effective end: %v", end)

	collision := IntersectsSegmentCircle(game, start, end, ship, ship.Radius+obstacle.GetRadius()+BUFFER_TOLERANCE)
	//	if s, ok := obstacle.(*Ship); ok && collision {
	//		//		game.Log("%v [%v, %v] will collide with %v [%v, %v]", ship, ship.NextSpeed, ship.NextAngle, s, s.NextSpeed, s.NextAngle)
	//	} else {
	//		//		game.Log("No collision")
	//	}
	return collision
}

type Arc struct {
	lower, upper float64
	obstacle     Entity
}

func (a Arc) String() string {
	return fmt.Sprintf("{L: %v, U: %v, O: %v}", a.lower, a.upper, a.obstacle)
}

func (a Arc) MidAngle() float64 {
	halfWidth := a.Width() / 2.0
	return math.Mod(a.lower+halfWidth, 360.0)
}

func (a Arc) Width() float64 {
	if a.lower > a.upper {
		// it goes over the 0-360 gap
		return ((360 - a.lower) + a.upper)
	} else {
		return (a.upper - a.lower)
	}
}

func GenClosestAngles(game Game, arcs []Arc, targetAngle float64, includeEnemies bool) (int, int, bool) {
	// game.Log("Target Angle: %v", targetAngle)
	// game.Log("includeEnemies: %v", includeEnemies)

	min := int(math.Floor(targetAngle))
	max := int(math.Ceil(targetAngle))
	minMovedBy := 0
	maxMovedBy := 0

	blockers := true
	originalArcs := make([]Arc, len(arcs))
	for i, arc := range arcs {
		originalArcs[i] = arc
	}

	for len(arcs) > 0 && blockers {
		// game.Log("Checking min %v and max %v against arcs %v", min, max, arcs)
		blockers = false

		var blocker Arc

		for _, arc := range arcs {
			if obsShip, ok := arc.obstacle.(*Ship); arc.obstacle == nil || ok && obsShip.Owner != game.pid && !includeEnemies {
				continue
			}
			game.Log("Checking min %v and max %v against arc %v", min, max, arc)
			if InArc(game, arc, float64(min)) {
				game.Log("Min %v was blocked by arc %v", min, arc)
				prevMin := min
				min = int(math.Floor(arc.lower))
				prevMin = int(math.Abs(float64(min - prevMin)))
				if prevMin > 180 {
					prevMin = 360 - prevMin
				}
				minMovedBy += prevMin
				blockers = true
			}
			if InArc(game, arc, float64(max)) {
				game.Log("Max %v was blocked by arc %v", max, arc)
				prevMax := max
				max = int(math.Ceil(arc.upper))
				prevMax = int(math.Abs(float64(max - prevMax)))
				if prevMax > 180 {
					prevMax = 360 - prevMax
				}
				maxMovedBy += prevMax
				blockers = true
			}
			if blockers {
				// game.Log("Blocked by arc %v", arc)
				//game.Log("min: %v", min)
				//game.Log("max: %v", max)
				blocker = arc
				break
			}
		}
		if !blockers {
			// game.Log("Final min, max: %v, %v", min, max)
		} else {
			tmp := []Arc{}
			for _, a := range arcs {
				if a != blocker {
					tmp = append(tmp, a)
				}
			}
			arcs = tmp
		}

	}

	// game.Log("maxMovedBy: %v", maxMovedBy)
	// game.Log("minMovedBy: %v", minMovedBy)
	if maxMovedBy+minMovedBy >= 360 {
		if includeEnemies {
			return GenClosestAngles(game, originalArcs, targetAngle, !includeEnemies)
		} else {
			return min, max, false
		}
	}
	return min, max, true
}
func InArc(game Game, arc Arc, testAngle float64) bool {
	t := float64(testAngle)
	ret := false
	if arc.lower > arc.upper {
		game.Log("Checking %v < %v or %v > %v = %v", t, arc.upper, t, arc.lower, t < arc.upper || t > arc.lower)
		ret = t <= arc.upper || t >= arc.lower
	} else {
		game.Log("Checking %v <= %v <= %v = %v", arc.lower, t, arc.upper, arc.lower <= t && t <= arc.upper)
		ret = arc.lower <= t && t <= arc.upper
	}
	game.Log("Result: %v", ret)
	return ret
}

func GetBestAngle(min, max int, targetAngle float64) int {
	angle := max
	if min > max {
		if targetAngle > float64(min) {
			if math.Abs(targetAngle-float64(min)) < math.Abs(targetAngle-360.0-float64(max)) {
				angle = min
			}
		} else if targetAngle < float64(max) {
			if math.Abs(targetAngle+360.0-float64(min)) < math.Abs(targetAngle-float64(max)) {
				angle = min
			}
		}
	} else if math.Abs(targetAngle-float64(min)) < math.Abs(targetAngle-float64(max)) {
		angle = min
	}

	return angle
}

func GenArcs(game Game, ship *Ship, target Entity, obstacles []Entity, speed int, enemyRange float64) []Arc {
	var arcs []Arc
	game.Log("Generating arcs for obstacles %v", obstacles)
	for _, obstacle := range obstacles {

		e_r := -1.0
		if obs, ok := obstacle.(*Ship); ok && obs.Owner != game.pid && obstacle != target {
			e_r = enemyRange
		}
		// game.Log("Using enemy_range as %v", e_r)
		// game.Log("Generating arcs for %v and obstacle %v", ship, obstacle)
		a := GetObstacleArcs(game, ship, obstacle, speed, e_r)
		game.Log("Generated arcs for obstacle %v: %v", obstacle, a)
		arcs = append(arcs, a...)
	}

	arcs = append(arcs, GenMapEdgeArcs(game, ship, speed)...)

	if game.turn < 1 && game.CurrentPlayers() > 2 && ALLY_BEE_DANCE {
		arcs = append(arcs, GenHoneyBeeArcs(game, ship)...)
	}
	//	game.LogEach("All arcs:", arcs)
	return arcs
}

func GetHoneyBeeMod(game Game, ship *Ship) int {
	return ModInt(game.pid+ship.GetId(), BEE_BANDWIDTH)
}

func GenHoneyBeeArcs(game Game, ship *Ship) []Arc {
	arcs := []Arc{}

	mod := GetHoneyBeeMod(game, ship)

	for i := float64(mod); i < 360.0+BEE_BANDWIDTH; i += BEE_BANDWIDTH {
		c := i - (BEE_BANDWIDTH / 2.0)
		w := ((BEE_BANDWIDTH - 1.0) / 2.0)
		arc := Arc{c - w, c + w, ship}
		arcs = append(arcs, arc)
	}
	return arcs
}

func GenMapEdgeArcs(game Game, ship *Ship, speed int) []Arc {
	arcs := []Arc{}

	buffer := MAX_SPEED * 4
	NW := Point{X: SHIP_RADIUS, Y: SHIP_RADIUS}
	NE := Point{X: float64(game.Width()) - SHIP_RADIUS, Y: SHIP_RADIUS}
	SE := Point{X: float64(game.Width()) - SHIP_RADIUS, Y: float64(game.Height()) - SHIP_RADIUS}
	SW := Point{X: SHIP_RADIUS, Y: float64(game.Height()) - SHIP_RADIUS}

	lineString := []Point{NW, NE, SE, SW, NW}

	for i := 0; i < 4; i++ {
		f := lineString[i]
		s := lineString[i+1]

		f = f.OffsetTowards(s, float64(-buffer))
		s = s.OffsetTowards(f, float64(-buffer))

		intersections := IntersectSegmentCircleCoords(f, s, ship.AsRadius(float64(speed)), false)

		if len(intersections) > 1 {
			arc := CreateArc(ship, &Planet{Owner: ship.Owner}, Chord{lower: intersections[0], upper: intersections[1]})
			arcs = append(arcs, arc)
		}
	}

	return arcs
}

type Chord struct {
	lower, upper Point
}

func GetObstacleArcs(game Game, ship *Ship, obstacle Entity, speed int, effectiveRadius float64) []Arc {
	var arcs []Arc
	if effectiveRadius < 0 {
		effectiveRadius = obstacle.GetRadius() + ship.GetRadius() + BUFFER_TOLERANCE
	}
	//game.Log("effectiveRadius of obstacle %v: %v", obstacle, effectiveRadius)

	shipObstacleDistance := ship.Dist(obstacle)

	var chords []Chord
	if shipObstacleDistance < effectiveRadius {
		c := ChordFromInterior(ship, obstacle)
		// game.Log("Chords from interior: %v", c)
		chords = append(chords, c)
	} else {
		c := ChordsFromComet(game, ship, obstacle, speed, effectiveRadius)
		game.Log("Chords from comet: %v", c)
		chords = append(chords, c...)
	}

	// game.Log("Chords for obstacle %v: %v", obstacle, chords)
	for _, c := range chords {
		arcs = append(arcs, CreateArc(ship, obstacle, c))
	}

	// game.Log("Arcs for obstacle %v: %v", obstacle, arcs)
	return arcs
}

func ChordsFromComet(game Game, ship Entity, obstacle Entity, speed int, effectiveRadius float64) []Chord {
	shipObstacleDistance := ship.Dist(obstacle)
	shipObstacleAngle := ship.Angle(obstacle)

	vector := ship.AsRadius(0)
	obstacleEnd := obstacle.AsRadius(0)

	if obs, ok := obstacle.(*Ship); ok {
		vector = ship.OffsetPolar(float64(obs.NextSpeed), float64(obs.NextAngle))
		//		game.Log("Obstacle is a ship, so projecting vector %v: %v", ship, vector)
		obstacleEnd = obs.Projection().AsRadius(0)
		//		game.Log("Obstacle %v is a ship, so projecting it: %v", obstacle, obstacleEnd)
	}
	speedCircle := ship.AsRadius(float64(speed))
	headCircle := obstacleEnd.AsRadius(float64(effectiveRadius))

	//	game.Log("speedCircle: %v", speedCircle.GetRadius())
	//	game.Log("headCircle: %v", headCircle.GetRadius())
	//	game.Log("effectiveRadius: %v", effectiveRadius)
	// the head circle forms the comets with the two line segments
	theta := RadToDeg(math.Asin(effectiveRadius / shipObstacleDistance))
	//	game.Log("Theta: %v", theta)
	//	game.Log("shipObstacleDistance: %v", shipObstacleDistance)
	//	game.Log("shipObstacleAngle: %v", shipObstacleAngle)
	tanDist := shipObstacleDistance * math.Cos(DegToRad(theta))
	//	game.Log("tanDist: %v", tanDist)
	tanUpper := vector.OffsetPolar(tanDist, shipObstacleAngle+theta)
	tanLower := vector.OffsetPolar(tanDist, shipObstacleAngle-theta)
	infUpper := vector.OffsetPolar(1000, shipObstacleAngle+theta)
	infLower := vector.OffsetPolar(1000, shipObstacleAngle-theta)

	game.DrawArc(headCircle, headCircle.Radius, headCircle.Angle(tanLower), headCircle.Angle(tanUpper), GREEN, .5, COMET_DISPLAY)
	game.DrawLineString([]Entity{tanLower, infLower, infUpper, tanUpper}, GREEN, .5, COMET_DISPLAY)

	// these are the intersections with each separate part of the comet
	//	game.Log("tanUpper: %v, infUpper: %v, speedCircle: %v", tanUpper, infUpper, speedCircle)
	sideUpper := IntersectSegmentCircleCoords(tanUpper, infUpper, speedCircle, true)
	game.Log("Upper points: %v", sideUpper)
	//	game.Log("tanLower: %v, infLower: %v, speedCircle: %v", tanLower, infLower, speedCircle)
	sideLower := IntersectSegmentCircleCoords(tanLower, infLower, speedCircle, true)
	game.Log("Lower points: %v", sideLower)
	//	game.Log("speedCircle: %v, headCircle: %v", tanUpper, infUpper, speedCircle)
	head := IntersectCircleCircleCoords(game, speedCircle, headCircle)

	head = EntitiesToPoints(Filter(PointsToEntities(head), func(e Entity) bool { return e.Dist(vector) < tanDist }))
	game.Log("Head points: %v", head)

	numUpper := len(sideUpper)
	numLower := len(sideLower)
	numHead := len(head)
	sum := numUpper + numLower + numHead

	var chords []Chord
	var points []Point
	//game.Log("Sum of comet points: %v", sum)

	game.Log("LHU - L:%v, H:%v, U:%v intersections with the comet of %s", numLower, numHead, numUpper, obstacle)
	if sum < 2 {

	} else if sum == 2 {
		points = append(points, sideLower...)
		points = append(points, head...)
		points = append(points, sideUpper...)
		chords = append(chords, Chord{points[0], points[1]})
	} else if sum == 3 {
		if len(head) > 0 && ship.Dist(head[0])+headCircle.GetRadius()-speedCircle.GetRadius() < POINT_MATCH_TOLERANCE {
			points = append(points, sideLower...)
			points = append(points, sideUpper...)
			chords = append(chords, Chord{points[0], points[1]})
		} else if numUpper == 2 {
			chords = append(chords, Chord{sideUpper[0], sideUpper[1]})
		} else if numLower == 2 {
			chords = append(chords, Chord{sideLower[0], sideLower[1]})
		}
	} else if sum == 4 {
		chords = append(chords, Chord{sideLower[0], sideUpper[0]})
		sideLower = sideLower[1:]
		sideUpper = sideUpper[1:]
		points = append(points, sideLower...)
		points = append(points, head...)
		points = append(points, sideUpper...)
		chords = append(chords, Chord{points[0], points[1]})
	} else {
		panic("I shouldn't be able to get 5 intersections")
	}
	return chords
}
func IntersectCircleCircleCoords(game Game, circle, target Point) []Point {
	//game.Log("circle = %v, target = %v", circle, target)
	d := circle.Dist(target)
	sumR := circle.GetRadius() + target.GetRadius()
	diffR := math.Abs(circle.GetRadius() - target.GetRadius())

	var intersections []Point

	//game.Log("distance from circle to head: %v", d)
	//game.Log("min distance for overlap: %v", sumR)
	if math.Pow(d-sumR, 2) < math.Pow(POINT_MATCH_TOLERANCE, 2) {
		//game.Log("returning tangent point")
		intersections = []Point{target.OffsetTowards(circle, target.GetRadius())}
	} else if math.Pow(d-diffR, 2) < math.Pow(POINT_MATCH_TOLERANCE, 2) {
		//game.Log("returning tangent point")
		intersections = []Point{target.OffsetTowards(circle, target.GetRadius())}
	} else if d > sumR || d < diffR {
		//game.Log("Outside the range for overlap")
		intersections = []Point{}
	} else {
		//game.Log("IntersectCircleCircleCoords - d: %v", d)
		//game.Log("IntersectCircleCircleCoords - circle.r: %v", circle.GetRadius())
		//game.Log("IntersectCircleCircleCoords - target.r: %v", target.GetRadius())
		//  theta = math.degrees(math.acos(   (circle.radius      **           2 + d**2 - target.radius         **        2 )  / (2*circle.radius     *d)))
		//	thetaRad := math.Acos((math.Pow(circle.GetRadius(), 2) + d*d - math.Pow(target.GetRadius(), 2)) /  (2*circle.GetRadius()*d))
		//game.Log("IntersectCircleCircleCoords - ThetaRad: %v", thetaRad)
		theta := RadToDeg(math.Acos((math.Pow(circle.GetRadius(), 2) + d*d - math.Pow(target.GetRadius(), 2)) / (2 * circle.GetRadius() * d)))
		//game.Log("IntersectCircleCircleCoords - Theta: %v", theta)
		base := circle.Angle(target)
		intersections = []Point{circle.OffsetPolar(circle.GetRadius(), base-theta), circle.OffsetPolar(circle.GetRadius(), base+theta)}
	}
	return intersections
}

func IntersectSegmentCircleCoords(start Point, end Point, circle Point, mustBeOnSegment bool) []Point {
	dx := end.GetX() - start.GetX()
	dy := end.GetY() - start.GetY()

	// Degenerate line segment
	if dx == 0 && dy == 0 {
		if start.Dist(circle) == circle.GetRadius() {
			return []Point{start}
		} else {
			return []Point{}
		}
	}

	// Degenerate circle
	if circle.GetRadius() == 0 {
		if IntersectionOnSegment(start, end, circle, 0, 0, true) != nil {
			return []Point{circle}
		} else {
			return []Point{}
		}
	}

	closestPoint, _ := GetIntersectionT(start, end, circle)
	closestDistance := closestPoint.Dist(circle)

	// No intersection
	if closestDistance > circle.GetRadius() {
		return []Point{}
	}

	var base, theta float64
	if closestPoint.IsSameLoc(start) || closestPoint.IsSameLoc(end) {
		otherEnd := start
		if closestPoint.IsSameLoc(start) {
			otherEnd = end
		}

		AB0 := closestPoint.Angle(circle)
		DB0 := closestPoint.Angle(otherEnd)
		ABD := 360 - math.Abs(AB0-DB0) // watch out for 0-360 wrap around
		ADB := RadToDeg(math.Asin(closestDistance * math.Sin(DegToRad(ABD)) / circle.GetRadius()))
		DAB := 180 - ABD - ADB
		base = circle.Angle(closestPoint)
		theta = DAB
	} else {
		base = circle.Angle(closestPoint)
		theta = RadToDeg(math.Acos(closestDistance / circle.GetRadius()))
	}

	// find which of the angles from the circle are actually on the line segment
	var intersections []Point
	for _, angle := range []float64{-theta, theta} {
		intersection := IntersectionOnSegment(start, end, circle, base, angle, mustBeOnSegment)
		if intersection != nil {
			inList := false
			for _, e := range intersections {
				if e.IsSameLoc(*intersection) {
					inList = true
					break
				}
			}
			if !inList {
				intersections = append(intersections, *intersection)
			}
		}
	}
	return intersections
}
func GetIntersectionT(start, end, circle Entity) (Point, float64) {
	dx := end.GetX() - start.GetX()
	dy := end.GetY() - start.GetY()

	a := dx*dx + dy*dy
	b := -2 * (start.GetX()*start.GetX() - start.GetX()*end.GetX() - start.GetX()*circle.GetX() + end.GetX()*circle.GetX() +
		start.GetY()*start.GetY() - start.GetY()*end.GetY() - start.GetY()*circle.GetY() + end.GetY()*circle.GetY())
	t := -b / (2 * a)
	if t > 1 {
		t = 1
	} else if t < 0 {
		t = 0
	}
	return Point{X: start.GetX() + dx*t, Y: start.GetY() + dy*t}, t

}

func IntersectionOnSegment(start, end, circle Point, base, theta float64, mustBeOnSegment bool) *Point {
	intersection := circle.OffsetPolar(circle.Radius, base+theta)
	if !mustBeOnSegment {
		return &intersection
	}
	AP := start.Dist(intersection)
	PB := end.Dist(intersection)
	AB := start.Dist(end)
	if AP+PB-AB < POINT_MATCH_TOLERANCE {
		return &intersection
	}
	return nil
}

func ChordFromInterior(ship *Ship, obstacle Entity) Chord {
	base := ship.Angle(obstacle)
	theta := 120.0
	lower := ship.OffsetPolar(1.0, base-theta) // dist doesn't matter, CreateArc cares only about the Angle
	upper := ship.OffsetPolar(1.0, base+theta)
	return Chord{lower, upper}
}
func CreateArc(ship, obstacle Entity, chord Chord) Arc {
	lower := ship.Angle(chord.lower)
	upper := ship.Angle(chord.upper)
	if lower > 180 && upper == 0 {
		upper = 360
	}
	arc := Arc{lower, upper, obstacle}
	return arc
}
