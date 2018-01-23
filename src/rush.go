package src

import (
	"math"
	"sort"
)

func RushConditions(game Game) bool {
	if !RUSH || game.InitialPlayers() > 2 {
		game.rushThisTurn = false
		return false
	}

	nearestEnemyPlayer := GetNearestEnemyPlayerId(game, game.MyShips())
	nearestPlayerShips := game.ShipsOwnedBy(nearestEnemyPlayer)

	myMean := game.MeanPoint(ShipsToEntities(game.MyShipsUndocked()))
	enemyMean := game.MeanPoint(ShipsToEntities(nearestPlayerShips))
	if len(game.MyPlanets()) < 1 && WithinDistance(myMean, enemyMean, 11*MAX_SPEED) {
		game.rushThisTurn = true
	} else {
		game.rushThisTurn = false
	}

	game.Log("Rush scenario")
	return game.rushThisTurn
}

func Rush(game Game) bool {
	game.Log("Rush scenario")
	nearestEnemyPlayer := GetNearestEnemyPlayerId(game, game.MyShips())
	nearestPlayerShips := game.ShipsOwnedBy(nearestEnemyPlayer)

	enemyFront := game.EnemyShipsUndocked()
	myFront := map[int]bool{}
	for _, ship := range game.MyShipsUndocked() {
		myFront[ship.GetId()] = true
	}

	myMean := game.MeanPoint(ShipsToEntities(game.MyShipsUndocked()))
	enemyMean := game.MeanPoint(ShipsToEntities(nearestPlayerShips))
	if len(game.ShipsUndockedOwnedBy(nearestEnemyPlayer)) >= len(game.MyShipsUndocked()) && WithinDistance(myMean, enemyMean, 5*MAX_SPEED) {
		AmbushCluster(game, enemyFront, &myFront)
		return true
	}

	if !WithinDistanceAny(myMean, ShipsToEntities(nearestPlayerShips), 8*MAX_SPEED) {
		return false
	}

	//	game.Log("myFront %v vs enemy %v", len(myFront), len(enemyFront))
	if len(myFront) > len(game.ShipsUndockedOwnedBy(nearestEnemyPlayer)) {
		pool := game.ShipsDockedOwnedBy(nearestEnemyPlayer)
		if len(pool) <= 0 {
			return false
		}
		nearest := Nearest(ShipsToEntities(pool), myMean)
		for sid, _ := range myFront {
			ship, _ := game.GetShip(sid)
			ship.ApproachTarget(game, nearest)
		}
		return true
	}

	bigEnough := []*Planet{}
	for _, neut := range game.PlanetsOwnedBy(-1) {
		if neut.DockingSpots >= 3 {
			bigEnough = append(bigEnough, neut)
		}
	}
	if len(bigEnough) == 0 {
		bigEnough = game.PlanetsOwnedBy(-1)
	}

	sort.SliceStable(bigEnough, func(i, j int) bool {
		return bigEnough[i].Dist(myMean) < bigEnough[j].Dist(myMean)
	})

	if len(bigEnough) > 0 {
		nearest := bigEnough[0]

		game.Log("Approaching target: %v", nearest)
		for _, ship := range game.MyShipsUndocked() {
			ship.ApproachTarget(game, nearest)
		}
		return true
	}
	return false
}

func TryUndocking(game Game, planet *Planet) {
	nearestEnemyPlayer := GetNearestEnemyPlayerId(game, game.MyShips())
	nearestPlayerShips := game.ShipsOwnedBy(nearestEnemyPlayer)
	enemyTurnsTo := TurnsTo(game, nearestPlayerShips, planet, DOCKING_RADIUS)

	if enemyTurnsTo <= DOCKING_TURNS+1 {
		for _, ship := range game.dockMap[planet.GetId()] {
			game.Undock(ship)
		}
	}

}

func GetNearestEnemyPlayerId(game Game, myShips []*Ship) int {
	var minShip *Ship
	minDist := 1000.0

	allEnemies := Filter(ShipsToEntities(game.AllShips()), func(e Entity) bool {
		//		game.Log("%v with owner %v vs friend %v", e, e.(*Ship).Owner, friend.Owner)
		return e.(*Ship).Owner != myShips[0].Owner && e.(*Ship).Owner != -2
	})
	for _, friend := range myShips {
		enemies := game.NearestEntities(friend, allEnemies, 1, minDist)
		if len(enemies) > 0 {
			enemy := enemies[0].(*Ship)
			d := friend.Dist(enemy)
			if minDist > d {
				minShip = enemy
				minDist = d
			}
		}
	}
	return minShip.Owner
}

func TooFarToRush(game Game, defenderShips, attackerShips []*Ship) bool {
	modePlanet := GetFirstPlanet(game, defenderShips)
	if modePlanet == nil {
		return false
	}
	// game.Log("Checking rush distance to %v", modePlanet)

	//	turnsToFirst := TurnsTo(game, defenderShips, modePlanet, DOCKING_RADIUS)
	// game.Log("Turns to arrival: %v", turnsToFirst)
	kiteFactor := 0
	if KITE {
		kiteFactor -= 1
	}

	attackerTurnsTo := TurnsTo(game, attackerShips, modePlanet, DOCKING_RADIUS)

	defenderTurnsToNextShip := TurnsToNextShip(game, defenderShips, *modePlanet, KITE)

	// game.Log("Turns to next ship: %v", defenderTurnsToNextShip)
	// game.Log("Turns to attacker arrival: %v", attackerTurnsTo)
	// game.Log("Max ships at arrival: %v", MaxShipsAtTurn(game, defenderShips, *modePlanet, KITE, attackerTurnsTo))
	if modePlanet.DockingSpots >= 3 && attackerTurnsTo > defenderTurnsToNextShip {
		// game.Log("PREDICTION: will survive")
	} else {
		// game.Log("PREDICTION: will be overrun")
	}
	return attackerTurnsTo > defenderTurnsToNextShip
}

func GetFirstPlanet(game Game, ships []*Ship) *Planet {
	planetChoices := map[int]int{}

	// If there are alreaedy some dock, we should make sure our count reflects that
	pid := ships[0].Owner
	for _, p := range game.AllPlanets() {
		if p.Owner == pid {
			planetChoices[p.GetId()] += p.DockedShips
		}
	}

	// find the one that has the most ships wanting to go to it, including those there already
	mode := -1
	for _, s := range ships {
		if s.Objectives == nil {
			GenerateObjectives(s, game)
		}
		p := -1

		// Not allowed to use someone else's planet as your first
		for _, obj := range *s.Objectives {
			pl := (obj.(Planet))
			if !pl.Owned || pl.Owner == s.Owner {
				// game.Log("Found a planet %v tha tis neutral or owned by s.Owner %v", pl, s.Owner)
				p = pl.GetId()
				break
			}
		}
		c := planetChoices[p]
		planetChoices[p] = c + 1
		if c+1 > planetChoices[mode] {
			mode = p
		}
	}
	ret, _ := game.GetPlanet(mode)
	// game.Log("first choice planet of ships %v: %v", ships, ret)
	return ret
}

func TurnsToNextShip(game Game, ships []*Ship, planet Planet, excludeKite bool) int {
	turns := 0
	waitTurns := map[int]int{}
	currentProduction := planet.CurrentProduction
	for _, s := range ships {
		if excludeKite && kite_conditions(game, s) {
			// do nothing, we're skipping the kite
		} else if s.IsUndocked() {
			waitTurns[s.GetId()] = TurnsTo(game, []*Ship{s}, planet, DOCKING_RADIUS) + DOCKING_TURNS // the one extra is the docking order
		} else if s.DockedStatus == DOCKING {
			//			game.Log("Progress is %v", s.DockingProgress)
			waitTurns[s.GetId()] = s.DockingProgress - 1
		} else if s.DockedStatus == DOCKED {
			waitTurns[s.GetId()] = -1
		}
	}
	if len(waitTurns) > 0 {
		for currentProduction < PRODUCTION_PER_SHIP {
			//			game.Log("currentProduction: %v", currentProduction)
			for k, v := range waitTurns {
				if v > 0 {
					waitTurns[k] -= 1
				} else {
					currentProduction += PRODUCTIVITY
				}
			}
			turns += 1
		}
	} else {
		turns = 10000
	}
	return turns
}

func MaxShipsAtTurn(game Game, ships []*Ship, planet Planet, excludeKite bool, turnsOut int) int {
	turns := 0
	waitTurns := map[int]int{}
	currentProduction := planet.CurrentProduction
	shipCount := 0
	//	sequence := 10000

	for i, s := range ships {
		if i > planet.DockingSpots {
			break
		}
		if excludeKite && kite_conditions(game, s) {
			// do nothing, we're skipping the kite
		} else if s.IsUndocked() {
			waitTurns[s.GetId()] = TurnsTo(game, []*Ship{s}, planet, DOCKING_RADIUS) + DOCKING_TURNS // the one extra is the docking order
		} else if s.DockedStatus == DOCKING {
			//			game.Log("Progress is %v", s.DockingProgress)
			waitTurns[s.GetId()] = s.DockingProgress - 1
		} else if s.DockedStatus == DOCKED {
			waitTurns[s.GetId()] = 0
		}
	}
	if len(waitTurns) > 0 {
		for turns < turnsOut {
			//			game.Log("currentProduction: %v", currentProduction)
			for k, v := range waitTurns {
				if v > 0 {
					waitTurns[k] -= 1
				} else {
					currentProduction += PRODUCTIVITY
				}
			}
			if currentProduction >= PRODUCTION_PER_SHIP {
				shipCount += 1
				currentProduction -= PRODUCTION_PER_SHIP
			}
			if turnsOut-turns == DOCKING_TURNS+1 {
				shipCount += len(waitTurns)
				return shipCount
			}
			turns += 1

			//          // If we can get some benefit out of investing this ship, do so
			//			if turnsOut-turns >= 2*(DOCKING_TURNS+1) {
			//				waitTurns[sequence] = DOCKING_TURNS
			//				sequence += 1

		}
	}
	return shipCount
}

// The number of turns that it will take for the furthest in a set of ships to
// arrive at an entity. This is an minimum, but it may be higher because of
// path considerations.
func TurnsTo(game Game, ships []*Ship, target Entity, radius float64) int {
	dists := make([]float64, len(ships))
	for i, s := range ships {
		dists[i] = s.Dist(target)
	}
	sort.Float64s(dists)

	maxDist := dists[len(ships)-1]
	if ships[0].Owner == game.pid && KITE && len(ships) >= 2 {
		maxDist = dists[len(ships)-2]
	}

	maxDist = math.Max(maxDist-radius-target.GetRadius(), 0)

	return int(math.Ceil(maxDist / MAX_SPEED))
}
