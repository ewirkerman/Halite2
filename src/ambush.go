package src

import (
	"math"
)

func Ambush(game Game, enemyFront []*Ship, myFront *map[int]bool) {
	enemy := FindTarget(game, myFront, enemyFront)

	SimulateEnemyMoves(game, true)

	// Checking for trying to charge an enemy outright
	for _, enemy := range enemyFront {
		chargeCount := 0
		for _, friend := range enemy.Combatants {
			if len(friend.Combatants) == 1 {
				chargeCount += 1
			}
		}
		if chargeCount >= MIN_ATTACK_SHIPS {
			//			for _, friend := range enemy.Combatants {
			//				friend.AttackShip(game, enemy, SHIP_RADIUS+1)
			//			}
			ChargeEnemy(game, enemy, myFront)
			return
		}
	}

	myShips := EntitiesToShips(game.NearestEntities(enemy, ShipsToEntities(game.MyShipsUndocked()), -1, -1))

	// game.Log("len(myShips) = %v vs len(MyUndocked())", len(myShips), len(game.MyShipsUndocked()))

	if len(myShips) <= 1 {
		ship := myShips[0]
		target := ClosestPlanetElseShip(game, []int{enemy.Owner}, enemy, ship)
		if target == nil {
			target = Point{X: float64(game.Width() / 2), Y: float64(game.Height() / 2)}
		}
		Bait(game, ship, target, enemy, WEAPON_RADIUS+SHIP_RADIUS*2+MAX_SPEED)
		return
	}

	// game.Log("Trying to snipe unprojections")
	for _, e := range enemyFront {
		if MoveToSnipe(game, e.Unprojection(game), myShips, SHIP_RADIUS*2+WEAPON_RADIUS) {
			return
		}
	}

	// nearest is bait (stays out of enemy range, but still closest (ideally)
	bait := myShips[0]
	delete(*myFront, bait.Id)
	lead := myShips[1]
	//	flank := myShips[2]

	for i := 1.0; i <= MAX_SPEED; i++ {
		game.DrawEntity(lead.AsRadius(i), LIGHT_GRAY, 1.0, NAV_DISPLAY)
		//		game.DrawEntity(flank.AsRadius(i), LIGHT_GRAY, 1.0, NAV_DISPLAY)
	}

	//	flank := enemy.Combatants[2]
	//command := Navigate(bait, enemy.OffsetTowards(lead, 1 ), game, MAX_SPEED + 2 * SHIP_RADIUS + WEAPON_RADIUS,-1,-1)
	Bait(game, bait, lead, enemy, MAX_SPEED+2*SHIP_RADIUS+WEAPON_RADIUS)
	//command := Navigate(bait, bait.ClosestPoint(enemy, MAX_SPEED + 2 * SHIP_RADIUS + WEAPON_RADIUS).OffsetTowards(lead, 1 ), game, -1,-1,-1)
	//game.ThrustCommand(command)

	// game.Log("Trying to snipe projections")
	for _, e := range enemyFront {
		if MoveToSnipe(game, e.Projection(), myShips[1:], SHIP_RADIUS*2+WEAPON_RADIUS) {
			return
		}
	}

	// game.Log("Taking up positions")
	Flank(game, enemy.Projection(), bait.Projection(), lead, 6, 0, -1)
	// game.Log("Positioning remaining: %v", myShips[2:])
	for _, ship := range myShips[2:] {
		Flank(game, enemy.Unprojection(game), lead.Projection(), ship, 1, 0, -1)
	}
}

func MoveToSnipe(game Game, enemyPosition *Ship, ships []*Ship, radius float64) bool {
	if radius <= 0 {
		radius = WEAPON_RADIUS + MAX_SPEED + 2*SHIP_RADIUS
	}
	shipMoveChoices := GenerateMoveChoices(game, enemyPosition, SetOfShips(ships), radius, radius)
	if len(shipMoveChoices) < MIN_ATTACK_SHIPS {
		// game.Log("Fewer that MIN_ATTACK_SHIPS %v ships on enemy %v", MIN_ATTACK_SHIPS, enemyPosition)
		return false
	}
	var bestCombo *[]Command
	bestCombo = BestMoveCombo(game, shipMoveChoices, bestCombo, enemyPosition, 1000)
	if bestCombo != nil && len(*bestCombo) >= len(ships) {
		for _, command := range *bestCombo {
			game.ThrustCommand(command)
		}
		return true
	}

	// game.Log("Moved none to snipe!")
	return false
}

func AmbushCombatants(game Game, enemyFront []*Ship) {
	for _, enemy := range enemyFront {
		for _, ship := range game.MyShipsUndocked() {
			s := EntitiesToShips(game.NearestEntities(ship, ShipsToEntities(game.EnemyShipsUndocked()), -1, -1))
			ship.Combatants = s
		}
		enemy.Combatants = EntitiesToShips(game.NearestEntities(enemy, ShipsToEntities(game.MyShipsUndocked()), -1, -1))
		// game.Log("enemy%v.Combatants = %v", enemy, enemy.Combatants)
		// game.Log("enemy%v.Combatants[1:] = %v", enemy, enemy.Combatants[1:])
		// game.Log("enemy%v.Combatants[2:] = %v", enemy, enemy.Combatants[2:])
		// game.Log("enemy%v.Combatants[3:] = %v", enemy, enemy.Combatants[3:])
	}
}

func FindTarget(game Game, myFront *map[int]bool, enemyFront []*Ship) *Ship {
	var closest *Ship
	closestDist := 1000.0
	for _, enemy := range enemyFront {
		//		sum := SumDist(game, enemy, ShipsToEntities(game.MyShipsUndocked()))
		near := game.NearestEntities(enemy, ShipsToEntities(game.MyShipsUndocked()), 1, -1)
		var sum float64
		if len(near) < 1 {
			near = game.NearestEntities(enemy, ShipsToEntities(game.MyShips()), 1, -1)
		}
		sum = near[0].Dist(enemy)

		if closestDist > sum {
			closest = enemy
			closestDist = sum
		}
	}
	return closest
}

func SumDist(game Game, source Entity, targets []Entity) float64 {
	sum := 0.0
	for _, target := range targets {
		sum += source.Dist(target)
	}
	// game.Log("Sum of distances to %v from %v is %v", source, targets, sum)
	return sum
}

func BestMoveCombo(game Game, shipMoveChoices map[int][]Point, BEST *[]Command, enemy *Ship, bestMutualDist float64) *[]Command {
	combos := GeneratePointCombinations(game, shipMoveChoices)
	// for each possible combination of choices
	for _, combo := range combos {
		if BEST != nil && len(combo) <= len(*BEST) {
			//			game.Log("Combo isn't longer")
			continue
		}

		if ValidateMoves(game, combo) {
			// game.Log("Valid moveset:")
			//			for _, option := range combo {
			//				ship, _ := game.GetShip(option.ShipIDRef)
			//				// game.Log("Ship %v to %v (d=%v)", ship, option, option.Dist(enemy))
			//			}

			// keep track of the longest valid combo
			if BEST == nil || len(combo) >= len(*BEST) {
				mutualDist := 0.0
				for i, c1 := range combo {
					for _, c2 := range combo[i:] {
						mutualDist += c1.Dist(c2)
					}
				}

				if mutualDist >= bestMutualDist {
					continue
				}
				if BEST != nil && len(combo) > len(*BEST) {
					bestMutualDist = mutualDist
				}

				commandCombo := []Command{}
				for _, option := range combo {
					ship, _ := game.GetShip(option.ShipIDRef)
					command := Command{Ship: ship, Speed: option.Speed, Angle: option.Degrees}
					commandCombo = append(commandCombo, command)
				}

				BEST = &commandCombo
				// game.Log("Found a mutual dist as low as %v with len %v!", mutualDist, len(commandCombo))
			}

			// if it's as long as it can be, just stop, we'll take it
			//if len(*BEST) == len(*myFront) {
			//	break All
			//}
		}
	}
	return BEST
}

func ExplodeSetRecursive(game Game, set []int, points map[int][]Point) [][]Point {
	if len(set) < 1 {
		return [][]Point{[]Point{}}
	} else {
		// game.Log("Exploding recursive set: %v", set)
		var sets [][]Point
		for _, point := range points[set[0]] {
			// game.Log("To point %v adding...", point)
			extensions := ExplodeSetRecursive(game, set[1:], points)
			// game.Log("... each of %v", extensions)
			for _, extension := range extensions {
				newSet := append(extension, point)
				sets = append(sets, newSet)
			}
		}
		// game.Log("Returning recursive [][]Point sets: %v", sets)
		return sets
	}
}

func GeneratePointCombinations(game Game, points map[int][]Point) [][]Point {
	// game.Log("Set combinations")
	//	for shipid, set := range points {
	//		// game.Log("Set for Ship %v: %v", shipid, set)
	//	}

	var pointKeys []int
	for key, _ := range points {
		pointKeys = append(pointKeys, key)
	}

	var sets [][]int
	for i := 1.0; i < math.Pow(2.0, float64(len(points))); i++ {
		var set []int
		for j := 0; j < len(points); j++ {
			if hasBit(int(i), uint(j)) {
				set = append(set, pointKeys[j])
			}
		}
		sets = append(sets, set)
	}

	// game.Log("Set combinations")
	//	for _, set := range sets {
	//		// game.Log("Set: %v", set)
	//	}

	var allSets [][]Point
	for _, set := range sets {
		// This if is limiting it to full combinations, not partial matches, which should be singles anyway
		if len(set) == len(points) {
			allSets = append(allSets, ExplodeSetRecursive(game, set, points)...)
		}
	}

	// game.Log("AllSet combinations")
	//	for _, set := range allSets {
	//		// game.Log("Set: %v", set)
	//	}
	return allSets
}
