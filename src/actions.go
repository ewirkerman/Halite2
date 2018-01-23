package src

import "math"

// Custom actions for ships
func (e *Ship) AttackShipCommand(game Game, target *Ship, minDist float64) Command {
	if minDist <= 0 {
		minDist = WEAPON_RADIUS*.8 - SHIP_RADIUS
	}

	//game.Log("*Ship %v attacking target %v", e, target)
	end := target.Projection()
	dist := e.Dist(end)
	var dest Point
	if dist > WEAPON_RADIUS*1.1 && target.DockedStatus != UNDOCKED {
		planet, _ := game.GetPlanet(target.DockedPlanet)
		angle := planet.Angle(target)
		dest = target.OffsetPolar(WEAPON_RADIUS*.8, angle)
	} else {
		dest = e.ClosestPoint(end, minDist)
	}
	return Navigate(e, dest, game, -1, -1, -1)
}

func (e *Ship) AttackPlanet(game Game, target Planet) {
	targets := game.dockMap[target.Id]
	target_ents := game.NearestEntities(e, ShipsToEntities(targets), 1, -1)
	//	game.Log("%v is attacking planet %v (docked ship: %v)", e, target, target_ents[0])
	if len(target_ents) > 0 {
		e.AttackShip(game, target_ents[0].(*Ship), -1)
	} else {
		e.ApproachTarget(game, target)
	}
}

func (e *Ship) AttackShip(game Game, target *Ship, minDist float64) {
	command := e.AttackShipCommand(game, target, minDist)
	game.ThrustCommand(command)
}

func (e *Ship) MoveToDock(game Game, planet Planet) {
	command := Command{}
	if !WithinDistance(e, planet, planet.GetRadius()+DOCKING_RADIUS+e.GetRadius()) {
		//game.Log("Moving to dock at %v and dock point: %v", planet, closestPoint)
		command = Navigate(e, e.ClosestPoint(planet, 0), game, -1, -1, -1)
		game.ThrustCommand(command)
	} else {
		//game.Log("Close enough to dock point at %v", planet)
		command = Command{OrderType: DOCK, Ship: e, target: planet}
		if len(game.PlanetsOwnedBy(-2)) == 0 && len(game.EnemyPlanetsOf(game.pid)) == 0 && len(game.PlanetsOwnedBy(-1)) == 1 && len(game.EnemyShips()) > 0 && len(game.ShipsDockedOwnedBy(-2)) > 0 {
			// don't dock the last one
		} else {
			game.Dock(command.Ship, command.target)
		}
	}
}

func (e *Ship) DefendShipFromShip(game Game, protectee, attacker *Ship) {
	start := attacker
	end := protectee
	closestPoint, _ := GetIntersectionT(start, end, e)
	shipLineDist := e.Dist(closestPoint)
	game.LogOnce("How I defend can almost certainly be improved")
	if false && shipLineDist < e.GetRadius() {
		//game.Log("Ship %v is on the line to enemy %v", e, enemies[0])
		e.AttackShip(game, attacker, -1)
	} else {
		// try to stay out of attack range
		d := attacker.Dist(protectee) - (WEAPON_RADIUS + MAX_SPEED + SHIP_RADIUS*2)
		// minimum MAX_SPEED/2
		d = math.Max(d, float64(WEAPON_RADIUS))
		// max of 3 turns out
		d = math.Min(float64(MAX_SPEED)*3.0, d)
		midpoint := protectee.OffsetTowards(attacker, d)
		//game.Log("Heading off enemy: %v from dockers[0] %v", enemies[0], dockers[0])
		//game.Log("Angle to cutoff point %v is %v", midpoint, e.Angle(midpoint))
		command := Navigate(e, midpoint, game, -1, -1, -1)
		game.ThrustCommand(command)
	}
}

func (e *Ship) DefendPlanet(game Game, planet Planet, enemies []*Ship) {
	if len(enemies) < 1 {
		enemies = game.GetAttackers(planet, -1, -1)
	}

	if len(enemies) > 0 && len(game.ShipsDockedAt(planet)) > 0 {
		dockers := EntitiesToShips(game.NearestEntities(enemies[0], ShipsToEntities(game.ShipsDockedAt(planet)), 1, -1))
		e.DefendShipFromShip(game, dockers[0], enemies[0])
	} else if len(enemies) < 1 {
		dockers := EntitiesToShips(game.NearestEntities(e, ShipsToEntities(game.ShipsDockedAt(planet)), 1, -1))
		e.AttackShip(game, dockers[0], -1)
	} else if len(enemies) > 0 {
		e.AttackShip(game, enemies[0], -1)
	} else {
		command := Navigate(e, e.ClosestPoint(planet, -1), game, -1, -1, -1)
		game.ThrustCommand(command)
	}
}

func (ship *Ship) DestroyPlanet(game Game, planet Planet) {
	command := Navigate(ship, planet, game, -1, -1, -1)
	game.ThrustCommand(command)
}

func (e *Ship) ApproachTarget(game Game, entity Entity) {
	target := e.ClosestPoint(entity, -1)
	command := Navigate(e, target, game, -1, -1, -1)
	game.ThrustCommand(command)
}

func Bait(game Game, bait *Ship, lead Entity, enemy *Ship, radius float64) {
	if radius <= 0 {
		radius = MAX_SPEED + WEAPON_RADIUS + 2*SHIP_RADIUS
	}
	var cmd Command
	game.Log("Using %v to bait %v at distance %v (currDist=%v) towards %v", bait, enemy, radius, bait.Dist(enemy), lead)
	if !WithinDistance(bait, enemy, radius+MAX_SPEED) {
		game.Log("Dist %v is beyond limit of %v", bait.Dist(enemy), radius+MAX_SPEED)
		cmd = Navigate(bait, bait.ClosestPoint(enemy, radius), game, radius, -1, -1)
	} else {
		game.Log("Dist %v is within limit of %v", bait.Dist(enemy), radius+MAX_SPEED)
		choices := GetHitNotMiss(game, bait, enemy, game.EnemyShipsUndocked(), radius, radius, false)
		choice_ents := PointsToEntities(choices)
		game.Log("Choice_ents: %v", choice_ents)
		game.Log("lead: %v", lead)
		choice_ents = game.NearestEntities(lead, choice_ents, -1, -1)
		choices = EntitiesToPoints(choice_ents)
		game.Log("Choices closest to %v: %v", lead, choices)
		if len(choices) > 0 {
			cmd = Navigate(bait, choices[0], game, -1, -1, -1)
		} else {
			game.Log("Why was I unable to find choices??", bait.Dist(enemy), radius+MAX_SPEED)
			cmd = Navigate(bait, bait.ClosestPoint(enemy, radius), game, radius, -1, -1)
		}
	}
	game.ThrustCommand(cmd)
}

func Flank(game Game, sourcePoint, focus Entity, ship *Ship, dist float64, dir int, enemyRadius float64) bool {
	if enemyRadius < 0 {
		enemyRadius = 2*SHIP_RADIUS + WEAPON_RADIUS + MAX_SPEED
	}

	//	game.Log("Positioning %v to snipe %v!", ship, sourcePoint)
	angle := sourcePoint.Angle(focus)
	options := []Entity{}
	if dir == 0 {
		options = []Entity{focus.OffsetPolar(dist, angle+90.0), focus.OffsetPolar(dist, angle-90.0)}
	} else if dir > 0 {
		options = []Entity{focus.OffsetPolar(dist, angle+90.0)}
	} else {
		options = []Entity{focus.OffsetPolar(dist, angle-90.0)}
	}
	target := game.NearestEntities(ship, options, 1, -1)[0]
	command := Navigate(ship, target, game, enemyRadius, -1, -1)
	game.ThrustCommand(command)

	targetAngle := ship.Angle(target)
	actualAngle := float64(command.Angle)
	lower := math.Mod(targetAngle-89.9, 360)
	upper := math.Mod(targetAngle+89.9, 360)
	return InArc(game, Arc{lower: lower, upper: upper}, actualAngle)
}

func (e *Ship) FleeTarget(game Game, ships []*Ship) Point {
	if len(ships) < 1 {
		panic("Fleeing nothing?")
	}
	target := game.MeanPoint(ShipsToEntities(ships))

	//	game.LogEach("Fleeing from...", ships)
	target = e.OffsetTowards(target, -MAX_SPEED)
	return target

}
