package src

func kite_conditions(game Game, ship *Ship) bool {
	ship.IsKite = false
	if !KITE {
		return false
	}
	if ship.Id != 3*ship.Owner {
		return false
	}
	return true
}

func kite_actions(game Game, ship *Ship) bool {
	//game.Log("Trying to kite")
	ship.IsKite = true
	enemies := ShipsToEntities(game.EnemyShips())
	enemies = game.NearestEntities(ship, enemies, -1, -1)
	nearestEnemy, _ := enemies[0].(*Ship)

	planets := game.AllPlanets()
	planet_ents := PlanetsToEntities(planets)
	planet_ents = Filter(planet_ents, func(entity Entity) bool {
		return entity.(*Planet).Owner == nearestEnemy.Owner
	})
	//points = EntitiesToPlanets(game.NearestEntities(*Ship, planet_ents, 1, -1))
	var bestPlanetOfEnemy Planet

	for len(*ship.Objectives) > 0 {
		o := (*ship.Objectives)[0]
		newObjectives := (*ship.Objectives)[1:]
		ship.Objectives = &newObjectives
		if o.(*Planet).Owner == nearestEnemy.Owner {
			bestPlanetOfEnemy = *(o.(*Planet))
			break
		}
	}

	if len(planet_ents) > 0 {
		kite_prep(ship, game, bestPlanetOfEnemy)
		ship.AttackPlanet(game, bestPlanetOfEnemy)
		kite_unprep(game)
	} else {

		var target Entity
		if game.CurrentPlayers() > 2 {
			// try to lead it towards another player's planet/ships
			target = ClosestPlanetElseShip(game, []int{game.pid, nearestEnemy.Owner}, nearestEnemy, ship)
			target = target.OffsetTowards(ship, -30)
		} else {
			// bait them towards the target, aka, stall hopefully
			target = ClosestPlanetElseShip(game, []int{nearestEnemy.Owner}, nearestEnemy, ship)
			// center := Point{X: float64(game.Width() / 2.0), Y: float64(game.Height() / 2.0)}
			// target = center.OffsetTowards(target, -target.Dist(center))
		}
		Bait(game, ship, target, nearestEnemy, WEAPON_RADIUS+SHIP_RADIUS*2+MAX_SPEED)
	}

	return true
}
func kite_prep(ship *Ship, game Game, planet Planet) {
	target_ents := ShipsToEntities(game.ShipsDockedAt(planet))
	targets := game.NearestEntities(ship, target_ents, 1, -1)
	if len(targets) > 0 {
		for _, enemy := range game.EnemyShipsUndocked() {
			enemy.SetRadius(WEAPON_RADIUS + SHIP_RADIUS + MAX_SPEED)
		}
	}
}
func kite_unprep(game Game) {
	for _, enemy := range game.EnemyShipsUndocked() {
		enemy.SetRadius(SHIP_RADIUS)
	}
}

func ClosestPlanetElseShip(game Game, avoidPids []int, fromShip, excludeShip *Ship) Entity {
	// nearest planet that isn't ours or theirs OR if not, then nearest enemy that isn't ours or theirs
	var target Entity
	var planets []*Planet
	var ships []*Ship
	for i := 0; i < 4; i++ {
		inAvoidPids := false
		for _, pid := range avoidPids {
			if i == pid {
				inAvoidPids = true
			}
		}
		if !inAvoidPids {
			planets = append(planets, game.PlanetsOwnedBy(i)...)
			ships = append(ships, game.ShipsOwnedBy(i)...)
		}
	}

	// game.Log("Planets owned by !%v: %v", avoidPids, planets)
	// game.Log("Ships owned by !%v: %v", avoidPids, ships)
	if len(planets) > 0 {
		target = game.NearestEntities(fromShip, PlanetsToEntities(planets), 1, -1)[0]
	} else {
		targets := game.NearestEntities(fromShip, ShipsToEntities(ships), 2, -1)
		for _, s := range targets {
			if s != excludeShip {
				target = s
				break
			}
		}
	}
	return target
}
