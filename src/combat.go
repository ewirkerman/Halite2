package src

import (
	"math"
	"sort"
)

type MoveChoice struct {
	speed, angle int
	Result       Point
	SortValue    float64
}

func ResolveCombat(game Game, enemyFront []*Ship, myFront *map[int]bool) map[int]bool {

	// Capture a list of ships we could use for combat
	// whatever is missing from this at the end is what we DID use for combat
	usedShips := map[int]bool{}
	for int, _ := range *myFront {
		usedShips[int] = true
	}

	sort.Sort(sort.Reverse(byId(enemyFront)))
	sort.Sort(sort.Reverse(byLenCombatants(enemyFront)))

	retreatableShips := map[int]bool{}
	releasedShips := map[int]bool{}

	// game.Log("Enemy front is...")
	//	for _, enemy := range enemyFront {
	//		// game.Log("%v: %v", enemy, len(enemy.Combatants))
	//	}
	for _, enemy := range enemyFront {
		inDamageRadius := func(f Entity) bool { return f.Dist(enemy) < WEAPON_RADIUS+2*SHIP_RADIUS }
		damageRangeCombatants := Filter(ShipsToEntities(enemy.Combatants), inDamageRadius)
		damage := WEAPON_DAMAGE * len(damageRangeCombatants)
		// game.LogOnce("I should make these guys automatically flee, because they won't have any damage")

		// only ships this didn't have enough to attack are going to be retreatable
		// ships that we don't need (for example, capped chasing) we use in any way
		for _, ship := range enemy.Combatants {
			if len(ship.Combatants) > 1 {
				retreatableShips[ship.GetId()] = true
			}
		}

		//		for _, f := range damageRangeCombatants {
		//			// game.Log("Insta-damage %v (%v vs %v): %v", enemy, enemy.HP, damage, f)
		//		}
		if enemy.HP < damage {
			// game.Log("Skipping enemy %v because it's already dead", enemy)
			continue
		}
		// game.LogOnce("I need to make sure that these ships that are in range of an enemy already aren't counted for other damage")
		// game.LogOnce("Also, the damage this expects them to deal up there is incorrect, it assumes full damage to each rather than split weapon damage")

		shipMoveChoices := GenerateMoveChoices(game, enemy.Projection(), *myFront, -1, -1)

		if len(shipMoveChoices) < MIN_ATTACK_SHIPS {
			//game.Log("Not enough attackers for %v", enemy)
			continue
		}

		// game.Log("Attackers on enemy %v: %v", enemy, enemy.Combatants)
		if ChargeConditions(game, enemy) {
			chargeRelease := ChargeEnemy(game, enemy, myFront)
			//			game.Log("Released from enemy %v: %v", enemy, chargeRelease)
			for ship, _ := range chargeRelease {
				//				game.Log("Releasing: %v", sid)
				releasedShips[ship] = true
			}
		} else {
			AcceptMovesOnEnemy(game, enemy, shipMoveChoices, myFront)
		}
	}

	// game.Log("Attempting to retreat: %v", retreatableShips)
	// game.Log("Release, not retreat: %v", releasedShips)
	RetreatShips(game, retreatableShips, releasedShips, myFront)

	// game.Log("Taking ships left in %v and removing them from usedShips %v", myFront, usedShips)
	// Those ships still in myFront are the ones that we did NOT use
	// so I remove them from the set of used ships
	for int, _ := range usedShips {
		if val, _ := (*myFront)[int]; val {
			//game.Log("Removing %v from used ships %s", int, usedShips)
			delete(usedShips, int)
		}
	}

	//game.Log("Ships remaining are what I used for combat: %v", usedShips)
	return usedShips
}

func GenerateMoveChoices(game Game, enemy *Ship, myFront map[int]bool, hitRange, missRange float64) map[int][]Point {
	// game.Log("Generating moves possible against %v...", enemy)

	if missRange <= 0 {
		missRange = WEAPON_RADIUS + 2*SHIP_RADIUS
	}
	if hitRange <= 0 {
		hitRange = WEAPON_RADIUS + 2*SHIP_RADIUS
	}

	eligible := map[int][]Point{}
	for _, friend := range game.MyShipsUndocked() {

		if !myFront[friend.GetId()] {
			continue
		}
		// game.Log("Generating choices against %v with %v", enemy, friend)

		//game.Log("friend.Combatants: %v", friend.Combatants)
		choices := GetHitNotMiss(game, friend, enemy.Projection(), game.EnemyShipsUndocked(), hitRange, missRange, true)
		if len(choices) > 0 {
			eligible[friend.GetId()] = choices
		}
	}

	return eligible
}
func AcceptMovesOnEnemy(game Game, enemy *Ship, shipMoveChoices map[int][]Point, myFront *map[int]bool) {
	//game.LogOnce("AcceptMovesOnEnemy should probably be distance to the enemy projection")

	// This is just to get the keys out of the map
	// Getting all the ships with choices
	ships := make([]*Ship, len(shipMoveChoices))
	i := 0

	//panic("Sort the ship move choices before using them")
	for shipId, _ := range shipMoveChoices {
		ship, _ := game.GetShip(shipId)
		ships[i] = ship
		i += 1
	}

	//Sorting them so that those nearest to the enemy go first
	ship_ents := ShipsToEntities(ships)
	ship_ents = game.NearestEntities(enemy, ship_ents, -1, -1)

	usedShips := 0

	distMap := map[Point]float64{}
	sortFunc := func(p1, p2 *Point) bool {
		//game.Log("distMap[*p1] (%v) < distMap[*p2] (%v)", distMap[*p1], distMap[*p2])
		return distMap[*p1] < distMap[*p2]
	}

	// game.Log("Accepting attack moves on enemy %v with ships %v", enemy, ships)
	// iterating over my attackers, going nearest first
	for _, ship_ent := range ship_ents {
		ship := ship_ent.(*Ship)
		shipId := ship.GetId()
		// game.Log("Accepting attack move on enemy %v with ship %v", enemy, ship)

		choices := make([]Point, len(shipMoveChoices[shipId]))
		for i, p := range shipMoveChoices[shipId] {
			choices[i] = p
		}
		//game.Log("Pre-sort:")
		//	for _, choice := range choices {
		//		//game.Log("Choice: %v, %v, %v, %v", choice.Speed, choice.Degrees, choice, choice.Distance)
		//	}

		for _, choice := range choices {
			var minNeighbor *Entity
			minNeighborDist := 1000.0
			for _, neighbor := range ship_ents {
				if neighbor.GetId() != ship_ent.GetId() {
					dist := choice.Dist(neighbor.(*Ship).Projection())
					// game.LogOnce("This is recording 35 as a possible neighbor when it's the moving ship")
					// game.Log("Possible neighbor: %v", neighbor)
					if dist < minNeighborDist || minNeighbor == nil {
						minNeighbor = &neighbor
						minNeighborDist = dist
					}
				}
			}
			distMap[choice] = minNeighborDist
			//game.Log("Choice: %v, minNeighborDist: %v, distMap[choice]: %v", choice, minNeighborDist, distMap[choice])
		}

		PointsBy(sortFunc).Sort(choices)

		// game.Log("Post-sort:")
		//		for _, choice := range choices {
		//			// game.Log("Choice: %v, %v, %v, %v", choice.Speed, choice.Degrees, choice, distMap[choice])
		//		}

		// Now that we have the sorted list of choice_ents, find a match
		for _, choice := range choices {
			command := Navigate(ship, choice, game, -1, choice.Speed, choice.Degrees)
			//game.Log("Checking %v if navgeo %v, %v == choiceEnt %v,%v", ship_ent, command.Speed, command.Angle, choice.Speed, choice.Degrees)
			if command.Speed == choice.Speed && command.Angle == choice.Degrees {
				game.ThrustCommand(command)
				delete(*myFront, shipId)
				usedShips += 1
				break
			}
		}
		// game.LogOnce("Sometimes not all of the ships get to move before I send one of them")
		// game.LogOnce("Look ahead doesn't work because one could get blocked by another")
		// game.LogOnce("Rollback is another option")
	}
}
func ChargeEnemy(game Game, enemy *Ship, myFront *map[int]bool) map[int]bool {
	enemyAlone := true
	for _, ship := range enemy.Combatants {
		if len(ship.Combatants) > 1 {
			enemyAlone = false
		}
	}

	// game.Log("Charge at %v with %v", enemy, enemy.Combatants)
	// game.Log("Enemy alone: %v", enemyAlone)

	numUsedShips := 0
	chargeRelease := map[int]bool{}
	for _, ship := range enemy.Combatants {
		if !(*myFront)[ship.GetId()] {
			// game.Log("Skipping %v because it's not in myFront, meaning it's already been ordered?")
			continue
		}
		// game.LogOnce("Removed max_chasers to allow parity test")
		if enemyAlone && numUsedShips >= MAX_CHASERS {
			// game.LogOnce("I want to release chasers but right now anyone not used goes into strategic retreat mode")
			chargeRelease[ship.GetId()] = true
		} else {
			e, _ := game.GetShip(enemy.GetId())
			point := ship.ClosestPoint(e, SHIP_RADIUS+1)
			// game.Log("Using ship %v to charge at %v (target = %v)", ship, enemy, point)
			command := Navigate(ship, point, game, -1, -1, -1)
			game.ThrustCommand(command)
			delete(*myFront, ship.GetId())
			numUsedShips += 1
		}
	}

	// Remove the released ships from enemyCombatants so that we know to defend against it still
	// We don't defend against checked enemy attackers right now
	enemy.Combatants = EntitiesToShips(Filter(ShipsToEntities(enemy.Combatants), func(e Entity) bool { return chargeRelease[e.GetId()] }))

	// game.Log("Ships on chargeRelease: %v", chargeRelease)
	return chargeRelease
}
func RetreatShips(game Game, retreatableShips, chargeRelease map[int]bool, myFront *map[int]bool) {
	protectees := make(map[int]Entity)

	myOriginalFront := []*Ship{}

	// game.LogOnce("Switching the below range will get me back to the retreatable behavior")
	for shipId, _ := range retreatableShips {
		// game.Log("Ships retreatable: %v", retreatableShips)
		// game.Log("Ships released from chase: %v", chargeRelease)
		// game.Log("Ships unordered: %v", myFront)
		retreatCond := (*myFront)[shipId] && !chargeRelease[shipId]
		if retreatCond {
			ship, _ := game.GetShip(shipId)
			protectees[shipId] = FindProtectee(game, ship)
			ship.DistanceTo = ship.Dist(protectees[shipId])
			myOriginalFront = append(myOriginalFront, ship)
		}
	}
	ship_ents := ShipsToEntities(myOriginalFront)
	sort.Sort(byDist(ship_ents))
	myOriginalFront = EntitiesToShips(ship_ents)

	for _, ship := range myOriginalFront {
		StrategicRetreat(game, ship, protectees[ship.GetId()], myFront)
	}
}
func FindProtectee(game Game, ship *Ship) Entity {
	dockedShips := make([]Entity, len(game.MyShips())-len(game.MyShipsUndocked()))

	i := 0
	for _, s := range game.MyShips() {
		if !s.IsUndocked() {
			dockedShips[i] = s
			i++
		}
	}

	protectees := game.NearestEntities(ship, dockedShips, 1, -1)

	if len(protectees) != 1 {
		protectees = game.NearestEntities(ship, PlanetsToEntities(game.AllPlanets()), 1, -1)
	}
	return protectees[0]
}

func StrategicRetreat(game Game, ship *Ship, target Entity, myFront *map[int]bool) {
	enemy := game.NearestEntities(ship, ShipsToEntities(ship.Combatants), 1, -1)[0].(*Ship)

	// game.Log("Retreating %v from %v", ship, enemy)
	enemyRange := WEAPON_RADIUS + 2*SHIP_RADIUS
	choices := GetHitNotMiss(game, ship, enemy, ship.Combatants, enemyRange, enemyRange, false)
	choice_ents := PointsToEntities(choices)
	choice_ents = game.NearestEntities(target, choice_ents, -1, -1)
	choices = EntitiesToPoints(choice_ents)

	fullRetreat := false
	// game.LogOnce("Full retreat option for outnumbered ships is below")

	//for _, e := range ship.Combatants {
	////game.Log("Enemy Id %v: %v", e.GetId(), e)
	////game.Log("Enemy %v of ship %v has combatants %v", e, ship, e.Combatants)
	//	if len(e.Combatants) < len(ship.Combatants) {
	//		fullRetreat = true
	//	//game.Log("Ship %v outnumbered: fullRetreat", ship)
	//	}
	//}

	//game.Log("Choices: %v", choices)

	if len(choices) > 0 && !fullRetreat {
		for len(choices) > 0 {
			choice := choices[0]
			choices = choices[1:]
			command := Navigate(ship, choice, game, -1, choice.Speed, choice.Degrees)
			if command.Angle == choice.Degrees {
				game.ThrustCommand(command)
				delete(*myFront, ship.GetId())
				return
			}
		}
	}

	// Move towards your protectee if you're distant, else just fall back to non-combat behaviors
	if !WithinDistance(ship, target, MAX_SPEED) {
		ship.ApproachTarget(game, target)
		delete(*myFront, ship.GetId())
	}
}

func GetHitNotMiss(game Game, mover, hit *Ship, misses []*Ship, hitRange, missRange float64, inside bool) []Point {
	hitUnproj := hit.Unprojection(game)

	missProjections := []Ship{}
	for _, miss := range misses {
		//		game.Log("Checking if miss %v is the same as the hit %v", miss.GetId(), hitUnproj.GetId())
		if miss.GetId() != hitUnproj.GetId() {
			missProjections = append(missProjections, *miss.Projection())
		}
	}
	//	game.Log("Misses: %v", missProjections)

	// game.LogOnce("This shouldn't, but may affect combat as well")
	//hit = hit.Projection()

	//	game.Log("hit: %v and hitMissRange: %v", hit, hitRange)
	distMoverHit := mover.Dist(hit)
	intersections := []Point{}
	if distMoverHit > hitRange+MAX_SPEED {
		// do nothing, we'll return an empty set of intersections
	} else {
		for i := 0; i <= MAX_SPEED; i++ {
			//			game.Log("Checking speed %v!", i)
			speedCircle := mover.AsRadius(float64(i))
			xs := IntersectCircleCircleCoords(game, speedCircle, hit.AsRadius(hitRange))
			for _, x := range xs {

				iAngleExact := mover.Angle(x)
				inAndOut := []int{int(math.Floor(iAngleExact)), int(math.Ceil(iAngleExact))}
				for _, angle := range inAndOut {
					//					game.Log("Checking angle at %v!", angle)
					result := mover.OffsetPolar(float64(i), float64(angle))
					result.ShipIDRef = mover.GetId()
					result.Speed, result.Degrees = i, angle

					if (result.Dist(hit) <= hitRange) != inside {
						game.DrawEntity(result.AsRadius(.1), RED, 1, NAV_DISPLAY)
						continue
					}
					game.DrawEntity(result.AsRadius(.1), GREEN, 0, NAV_DISPLAY)
					hitsMiss := false
					for _, miss := range missProjections {
						if result.Dist(&miss) <= missRange {
							//							game.Log("%v hits a miss %v (d=%v vs. %v)!", x, miss.Unprojection(game), result.Dist(&miss), missRange)
							game.DrawEntity(result.AsRadius(.1), RED, 1, NAV_DISPLAY)
							hitsMiss = true
							break
						}
					}
					if hitsMiss {
						continue
					}

					//					game.Log("Found an intersection at %v (d=%v)!", result, result.Dist(hit))
					game.DrawEntity(result.AsRadius(.1), GREEN, 0, NAV_DISPLAY)
					intersections = append(intersections, result)
				}
			}
		}
	}

	//	game.Log("Hit not miss intersections: %v", intersections)
	return intersections
}

func ChargeConditions(game Game, enemy *Ship) bool {
	outNumber := true
	for _, friend := range enemy.Combatants {
		//game.Log("Friend %v wants to charge %v: my enemies %v vs. target enemies %v", friend, enemy, len(friend.Combatants), len(enemy.Combatants))
		if len(friend.Combatants) >= len(enemy.Combatants) {
			//game.Log("Outnumbered: can't charge :(")
			outNumber = false
			break
		}
	}
	//game.Log("Charge decision: %v", outNumber)
	return outNumber
}

type byLenCombatants []*Ship

func (a byLenCombatants) Len() int      { return len(a) }
func (a byLenCombatants) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byLenCombatants) Less(i, j int) bool {
	if len(a[i].Combatants) == len(a[j].Combatants) {
		return a[i].GetId() > a[j].GetId()
	} else {
		return len(a[i].Combatants) < len(a[j].Combatants)
	}
}

type byId []*Ship

func (a byId) Len() int           { return len(a) }
func (a byId) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byId) Less(i, j int) bool { return a[i].GetId() < a[j].GetId() }
