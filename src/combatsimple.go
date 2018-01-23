package src

import "sort"

type CombatUsage int

const (
	USED CombatUsage = iota
	UNUSED
	RELEASED
)

func ResolveCombatSimple(game Game, enemyFront []*Ship, myFront *map[int]bool) map[int]bool {

	usedShips := map[int]CombatUsage{}
	for int, _ := range *myFront {
		usedShips[int] = UNUSED
	}

	safes := EntitiesToShips(Filter(ShipsToEntities(game.MyShips()), func(e Entity) bool {
		safe := true
		s := e.(*Ship)
		//		game.Log("%v has combatants: %v", s, s.Combatants)
		for _, enemy := range s.Combatants {
			if len(s.Combatants) >= len(enemy.Combatants) {
				//				game.Log("%v isn't safe because %v has only %v combatants", s, enemy, len(enemy.Combatants))
				safe = false
				break
			}
		}
		return safe
	}))
	//	game.LogEach("List of safe ships:", safes)

	unsafes := map[int][]*Ship{}

	for _, enemy := range enemyFront {
		enemyAlone := true
		for _, ship := range enemy.Combatants {
			if len(ship.Combatants) > 1 {
				enemyAlone = false
			}
		}

		if enemyAlone {
			//			game.Log("Enemy %v is alone, so not chasing", enemy)
			continue
		}

		if len(enemy.Combatants) < 1 {
			//			game.Log("Enemy %v has no combatants, skipping", enemy)
			continue
		}

		//		game.Log("Resolving combat with enemy %v", enemy)
		// check if outnumbered
		safe := true
		for _, friend := range enemy.Combatants {
			unmoved := EntitiesToShips(Filter(ShipsToEntities(enemy.Combatants), func(e Entity) bool {
				status, _ := usedShips[e.GetId()]
				return status == UNUSED
			}))
			if len(friend.Combatants) >= len(unmoved) {
				safe = false
			}
		}

		// if any of our attackers are outnumbered, go to a safe or flee
		if !safe {
			//			game.Log("Not safe to attack %v", enemy)
			if _, ok := unsafes[enemy.GetId()]; !ok {
				unsafes[enemy.GetId()] = []*Ship{}
			}
			unsafes[enemy.GetId()] = append(unsafes[enemy.GetId()], enemy.Combatants...)

			continue
		}

		// leads are those ships closest to the enemies where I outnumber them
		lead := enemy.Combatants[0]
		if status, _ := usedShips[lead.GetId()]; status != UNUSED {
			game.Log("Skipping %v because it's already been used", lead)
			continue
		}
		//		panic("If leadCombatants and wingman combtants == 1 then don't chase this guy")
		//		game.Log("Using lead %v to attack enemy %v", lead, enemy)

		// Make it so the lead doesn't charge ahead as blindly (waits for wingmen)
		furthestWingman := enemy.Combatants[len(enemy.Combatants)-1]
		if len(enemy.Combatants) > 8 {
			furthestWingman = enemy.Combatants[8]
		}
		fWDist := furthestWingman.Dist(enemy)
		targetDist := WEAPON_RADIUS - float64(len(enemy.Combatants))
		if fWDist-targetDist > MAX_SPEED {
			targetDist = fWDist - MAX_SPEED
		}

		//		if targetDist < 3.0 {
		//			targetDist = 3.0
		//		}

		target := enemy.OffsetTowards(lead, targetDist)
		game.ThrustCommand(Navigate(lead, target, game, -1, -1, -1))
		usedShips[lead.GetId()] = USED

		// Move wingmen in, following the lead
		for _, wingman := range enemy.Combatants[1:] {
			if status, _ := usedShips[wingman.GetId()]; status != UNUSED {
				game.Log("Skipping %v because it's already been used", wingman)
				continue
			}
			game.Log("Using wingman %v to attack enemy %v with lead %v", wingman, enemy, lead)
			usedShips[wingman.GetId()] = USED
			wentTowards := Flank(game, enemy, lead.Projection(), wingman, 4*SHIP_RADIUS, 0, SHIP_RADIUS)
			if !wentTowards {
				//				game.Log("Releasing wingman because he didn't go towards his destination")
				game.LogOnce("I can't release ships right now because people are expecting them to move")
				usedShips[wingman.GetId()] = USED
			}
		}

	}

	safeProjections := make([]*Ship, len(safes))
	for i, safe := range safes {
		safeProjections[i] = safe.Projection()
	}
	safeMap := map[int]*Ship{}
	allUnsafes := []*Ship{}

	//	game.Log("Assigning unsafes to safes")
	for _, enemy := range enemyFront {
		for _, friend := range unsafes[enemy.GetId()] {

			// ship has already moved, skip it
			if status, _ := usedShips[friend.GetId()]; status != UNUSED {
				//				game.Log("Skipping %v because it's already been used", friend)
				continue
			}
			safeProjections = EntitiesToShips(game.NearestEntities(friend, ShipsToEntities(safeProjections), -1, -1))

			if len(safes) > 0 {
				safe := safeProjections[0]
				safeMap[friend.GetId()] = safe
				allUnsafes = append(allUnsafes, friend)
			} else {
				target := friend.FleeTarget(game, friend.Combatants)
				game.ThrustCommand(Navigate(friend, target, game, SHIP_RADIUS, -1, -1))
			}
			//			game.Log("Using %v to retreat in simple combat", friend)
			usedShips[friend.GetId()] = USED

		}
	}

	// sorting them so those closest to their safes move first
	sort.Slice(allUnsafes, func(i, j int) bool {
		return allUnsafes[i].Dist(safeMap[allUnsafes[i].GetId()]) < allUnsafes[j].Dist(safeMap[allUnsafes[j].GetId()])
	})

	// moving the unsafes to their safes
	radius := MAX_SPEED + WEAPON_RADIUS + 2*SHIP_RADIUS
	for _, friend := range allUnsafes {
		safe := safeMap[friend.GetId()]
		//		game.Log("Retreating %v to safe %v", friend, safe)
		enemy := friend.Combatants[0]
		choices := GetHitNotMiss(game, friend, enemy, friend.Combatants[1:], radius, radius, true)
		//		game.LogEach("Found choices: ", choices)
		if len(choices) > 0 && len(enemy.Combatants) > 1 {
			nearest := Nearest(PointsToEntities(choices), safe)
			//			game.Log("Nearest to safe: %v", nearest)
			game.ThrustCommand(Navigate(friend, nearest, game, SHIP_RADIUS, -1, -1))
		} else {
			choices = GetHitNotMiss(game, friend, enemy, friend.Combatants[1:], radius, radius, false)
			if len(choices) > 0 {
				nearest := Nearest(PointsToEntities(choices), safe)
				//				game.Log("Nearest to safe: %v", nearest)
				game.ThrustCommand(Navigate(friend, nearest, game, SHIP_RADIUS, -1, -1))
			} else {
				mean := game.MeanPoint(ShipsToEntities(friend.Combatants))
				Flank(game, mean, safe, friend, 1, 0, SHIP_RADIUS*2)
			}
		}
	}

	boolUsedShips := map[int]bool{}

	for sid, status := range usedShips {
		if status != USED {
			delete(usedShips, sid)
			s, _ := game.GetShip(sid)
			game.Unthrust(s)
		} else {
			boolUsedShips[sid] = true
		}
	}

	game.Log("Ships used by SimpleCombat: %v", usedShips)
	return boolUsedShips
}
