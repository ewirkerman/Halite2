package src

import (
	"math"
	"sort"

	ordered_map "github.com/cevaris/ordered_map"
)

func fightForSecondConditions(game Game) (int, bool) {

	if game.CurrentPlayers() <= 2 {
		return -1, false
	}

	pids := []int{0, 1, 2, 3}
	most := 0
	mostPid := -1

	for _, pid := range pids {
		game.Log("Ships owned by %v: %v", pid, len(game.ShipsOwnedBy(pid)))
		if stage, ok := game.allyPidStages[pid]; ok && stage > 0 {
			game.Log("Ally")
			continue
		}
		if len(game.ShipsOwnedBy(pid)) > most {
			most = len(game.ShipsOwnedBy(pid))
			mostPid = pid
		}
	}
	mine := len(game.ShipsOwnedBy(game.pid))

	if most >= 40 && (most > 2*mine) {
		game.Log("Mine: %v vs most: %v - Flee", mine, most)
		return pids[mostPid], true
	}

	game.Log("Mine: %v vs most: %v - Fight", mine, most)
	return -1, false
}

func SetThingsStraight(game Game) {
	// If there is only one planet I don't own, blockade it
	// murder players in the order of lowest ship count to highest
}

func fightForSecond(game Game, avoidPlayer int, ships []*Ship) {

	// sort the currentPlayer ints by created shipCount
	pids := game.CurrentStandings()
	for i, pid := range pids {
		if game.Pid() == pid {
			// I'm in first, set things straight
			if i == 0 && false {
				SetThingsStraight(game)
				// I'm in second, turtle
			} else if i == 1 {
				Turtle(game, ships)

				// I'm in third or worse, murder
			} else {
				Circle(game, ships)
				//				murder(game, pids[0], ships)
			}
		}
	}
}

func murder(game Game, avoidPlayer int, ships []*Ship) {
	murderPids := []int{}
	for i := 0; i < game.InitialPlayers(); i++ {
		if game.pid == i || avoidPlayer == i {
			continue
		}

		// we'll only murder players without planets
		if len(game.PlanetsOwnedBy(i)) > 0 {
			continue
		}
		murderPids = append(murderPids, i)
	}

	shipTargets := []*Ship{}

	for _, pid := range murderPids {
		shipTargets = append(shipTargets, game.ShipsOwnedBy(pid)...)
	}

	shipEntTargets := ShipsToEntities(shipTargets)
	if len(ships) > 40 {
		for _, corner := range game.Corners() {
			sort.SliceStable(ships, func(i, j int) bool {
				return corner.Dist2(ships[i]) < corner.Dist2(ships[j])
			})
			cShips := ships[:10]
			ships = ships[10:]
			for _, s := range cShips {
				w := 5.0
				if s.GetX() <= w || s.GetX() >= float64(game.Width())-w || s.GetY() <= w || s.GetY() >= float64(game.Height())-w {
					s.AttackShip(game, Nearest(shipEntTargets, s).(*Ship), 3)
				} else {
					s.ApproachTarget(game, corner)
				}
			}
		}
	}
	for _, s := range ships {
		s.AttackShip(game, Nearest(shipEntTargets, s).(*Ship), 3)
	}

}

func blade(game Game, avoidPlayer int, ships []*Ship) {
	murderPids := []int{}
	for i := 0; i < game.InitialPlayers(); i++ {
		if game.pid == i || avoidPlayer == i {
			continue
		}

		// we'll only murder players without planets
		if len(game.PlanetsOwnedBy(i)) > 0 {
			continue
		}
		murderPids = append(murderPids, i)
	}

	//	// undock all ships
	//	for _, ship := range game.MyShipsDocked() {
	//		game.Undock(ship)
	//	}

	// set up the lists of attacks and targets
	//	enemyShips := EntitiesToShips(Filter(ShipsToEntities(game.EnemyShips()), func(e Entity) bool {
	//		for _, i := range murderPids {
	//			if e.(*Ship).Owner == i {
	//				return true
	//			}
	//		}
	//		return false
	//	}))

	// planet murder - disabled effectively by thte planet check during murder pids generation
	planetTargets := []*Planet{}

	for _, pid := range murderPids {
		planetTargets = append(planetTargets, game.PlanetsOwnedBy(pid)...)
	}
	if len(planetTargets) > 0 {
		myMean := game.MeanPoint(ShipsToEntities(ships))
		planet := EntitiesToPlanets(game.NearestEntities(myMean, PlanetsToEntities(planetTargets), 1, -1))[0]
		for _, ship := range ships {
			ship.AttackPlanet(game, *planet)
		}
		return
	}

	// scoop formation
	// calculate scoop size
	// if in formation:
	// if hit corner, turn corner
	// move scoop
	var base *Ship
	edgeDist := float64(game.Width())
	for _, ship := range ships {
		xEdge := math.Min(ship.X, float64(game.Width())-ship.X)
		yEdge := math.Min(ship.Y, float64(game.Height())-ship.Y)
		d := math.Min(xEdge, yEdge)
		if d < edgeDist {
			base = ship
			edgeDist = d
		}
	}

	// use the ship closests to the edge to generate the root
	if base != nil {

		//		panic("Extract, cap the length of the blade and have the other blades be a mirror of this one")
		xEdge := math.Min(base.X, float64(game.Width())-base.X)
		yEdge := math.Min(base.Y, float64(game.Height())-base.Y)
		var root Point
		var angle float64

		// root is the edge point nearest to the base ship
		// calc the scoop angle while we're here
		if xEdge < yEdge {
			root.Y = base.Y
			if base.X > float64(game.Width())/2 {
				root.X = float64(game.Width() - 1)
				angle = 210.0
			} else {
				root.X = 1
				angle = 30.0
			}
		} else {
			root.X = base.X
			if base.Y > float64(game.Height())/2 {
				root.Y = float64(game.Height() - 1)
				angle = 300.0
			} else {
				root.Y = 1
				angle = 120.0
			}
		}
		// game.Log("Root was: %v", root)
		if targetMap, ok := InPlaceForScoop(game, root, ships, angle); !ok {
			MoveToTargets(game, root, targetMap, -1.0)
		} else {
			offsetAngle := -1.0
			if WithinDistance(base, root, WEAPON_RADIUS+SHIP_RADIUS) {
				loopBuffer := float64(MAX_SPEED / 2)
				if root.Y < loopBuffer && !(root.X < loopBuffer) {
					offsetAngle = 180.0
				} else if root.X > float64(game.Width())-loopBuffer {
					offsetAngle = 270.0
				} else if root.Y > float64(game.Height())-loopBuffer {
					offsetAngle = 0.0
				} else if root.X < loopBuffer {
					offsetAngle = 90.0
				}
			}
			// game.Log("Root is: %v", root)
			targetMap, _ := InPlaceForScoop(game, root, ships, angle)
			MoveToTargets(game, root, targetMap, offsetAngle)
		}
	}
}

func MoveToTargets(game Game, root Point, targetMap ordered_map.OrderedMap, offsetAngle float64) {
	iter := targetMap.IterFunc()
	offsetDist := MAX_SPEED * 2.0
	for kv, ok := iter(); ok; kv, ok = iter() {
		ship := kv.Key.(*Ship)
		slot := kv.Value.(Point)

		if offsetAngle >= 0 {
			slot = slot.OffsetPolar(offsetDist, offsetAngle)
		}
		if !WithinDistance(ship, slot, 1) {
			game.ThrustCommand(Navigate(ship, slot, game, -1, -1, -1))
		}
	}
}

func InPlaceForScoop(game Game, root Point, ships []*Ship, angle float64) (ordered_map.OrderedMap, bool) {

	// Form the line
	targetMap := ordered_map.NewOrderedMap()
	return *targetMap, true
	outOfPositionShips := 0
	for i := 0; len(ships) > 0; i++ {
		slot := root.OffsetPolar(float64(i)*2.2*SHIP_RADIUS, angle)
		game.DrawEntity(slot.AsRadius(1.0), ORANGE, 1.0, NAV_DISPLAY)
		ships = EntitiesToShips(game.NearestEntities(slot, ShipsToEntities(ships), -1, -1))
		ship := ships[0]
		targetMap.Set(ship, slot)

		localPlanets := EntitiesToPlanets(game.NearestEntities(slot, PlanetsToEntities(game.AllPlanets()), 1, MAX_PLANET_RADIUS))
		if !WithinDistance(ship, slot, 1) {
			if !(len(localPlanets) > 0) || !WithinDistance(slot, localPlanets[0], localPlanets[0].GetRadius()) {
				// game.Log("%v NOT within distance of slot %v", ship, slot)
				outOfPositionShips += 1
			} else {
				// game.Log("%v within distance of slot %v", ship, slot)

			}
		}
		ships = ships[1:]
	}

	return *targetMap, outOfPositionShips < targetMap.Len()*8/10
}

func Turtle(game Game, ships []*Ship) {
	// undock all ships
	for _, ship := range game.MyShipsDocked() {
		game.Undock(ship)
	}

	myMean := game.MeanPoint(ShipsToEntities(game.MyShips()))

	corner := ClosestCorner(game, myMean)

	for _, ship := range ships {
		game.ThrustCommand(Navigate(ship, corner, game, SHIP_RADIUS*2, -1, -1))
	}

	//	for i, corner := range corners[:1] {
	//		ships = EntitiesToShips(game.NearestEntities(corner, ShipsToEntities(ships), -1, -1))
	//		part := int(1.0 / float64(len(corners)-i) * float64(len(ships)))
	//		cShips := ships[:part]
	//		ships = ships[part:]
	//		for _, ship := range cShips {
	//			game.ThrustCommand(Navigate(ship, corner, game, 3*MAX_SPEED, -1, -1))
	//		}
	//	}
}

func Circle(game Game, ships []*Ship) {
	center := Point{X: float64(game.Width() / 2.0), Y: float64(game.Height() / 2.0)}
	for _, ship := range game.MyShipsDocked() {
		game.Undock(ship)
	}

	for _, ship := range ships {
		point := ship.OffsetTowards(center, -MAX_SPEED)
		loopBuffer := float64(MAX_SPEED / 2)
		if ship.Y < loopBuffer && !(ship.X < loopBuffer) {
			point = ship.OffsetPolar(MAX_SPEED, 180)
		} else if ship.X > float64(game.Width())-loopBuffer {
			point = ship.OffsetPolar(MAX_SPEED, 270)
		} else if ship.Y > float64(game.Height())-loopBuffer {
			point = ship.OffsetPolar(MAX_SPEED, 0)
		} else if ship.X < loopBuffer {
			point = ship.OffsetPolar(MAX_SPEED, 90)
		}
		game.ThrustCommand(Navigate(ship, point, game, MAX_SPEED+WEAPON_RADIUS+2*SHIP_RADIUS, -1, -1))
	}
}
func (game *Game) Corners() []Point {
	return []Point{
		Point{X: 0, Y: 0},
		Point{X: float64(game.Width()), Y: 0},
		Point{X: 0, Y: float64(game.Height())},
		Point{X: float64(game.Width()), Y: float64(game.Height())}}
}

func ClosestCorner(game Game, p Point) Point {
	corners := game.Corners()

	return game.NearestEntities(p, PointsToEntities(corners), -1, -1)[0].(Point)
}
