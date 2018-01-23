package src

import "sort"

func AmbushSimple(game Game, enemyFront []*Ship, myFront *map[int]bool) {

	// find nearest enemy to our mean
	myShips := game.MyShipsUndocked()
	myMean := game.MeanPoint(ShipsToEntities(myShips))
	//	game.DrawEntity(myMean.AsRadius(1), ORANGE, 2, NAV_DISPLAY)
	enemyMean := game.MeanPoint(ShipsToEntities(enemyFront))
	game.DrawEntity(enemyMean.AsRadius(.2), ORANGE, 0, NAV_DISPLAY)

	if len(myShips) <= 0 {
		panic("I've already lost")
	} else if len(myShips) <= 1 {
		// hunt the enemy down if you have more health?
		// or maybe run away from him and dock?
		Circle(game, myShips)
	} else {
		//		targetDist := MAX_SPEED * 1.0
		//		if len(myShips) < len(enemyFront) {
		//		}

		c := make_circle(ShipsToEntities(myShips))
		flankDist := 3 * SHIP_RADIUS
		c.Radius = flankDist + SHIP_RADIUS
		game.DrawEntity(c, BLACK, 2, NAV_DISPLAY)
		game.Log("min circle: %v", c)
		cShip := &Ship{X: c.X, Y: c.Y, Owner: game.Pid(), Radius: c.Radius}

		targetDist := WEAPON_RADIUS + SHIP_RADIUS
		sort.Slice(enemyFront, func(i, j int) bool { return enemyFront[i].Dist(myMean) < enemyFront[j].Dist(myMean) })
		enemy := enemyFront[0]
		game.DrawEntity(enemy.AsRadius(SHIP_RADIUS), RED, 3, NAV_DISPLAY)
		target := enemy.OffsetTowards(myMean, targetDist)
		game.DrawEntity(target.AsRadius(SHIP_RADIUS), BLUE, 2, NAV_DISPLAY)
		// move lead to the spot nearest enemy toward their mean -WEAPON_RADIUS away
		sort.Slice(myShips, func(i, j int) bool { return myShips[i].Dist(cShip) < myShips[j].Dist(cShip) })

		// line these up so the lead doesn't get ahead of the others
		furthestWingman := myShips[len(myShips)-1]
		fWDist := furthestWingman.Dist(target)
		targetDist = 0
		if fWDist > MAX_SPEED {
			game.Log("FWD is higher than max speed, limiting the lead's speed")
			targetDist = fWDist - MAX_SPEED
		}

		lead := myShips[0]
		//		if lead.Dist(target) < MAX_SPEED*2.0+SHIP_RADIUS*2+WEAPON_RADIUS {
		target = target.OffsetTowards(lead, targetDist)
		//		}
		game.DrawEntity(target.AsRadius(1), RED, 2, NAV_DISPLAY)
		game.Log("Using lead %v to attack enemy %v", lead, enemy)

		cmd := NavigateWithoutShips(cShip, target, game)
		result := cShip.OffsetPolar(float64(cmd.Speed), float64(cmd.Angle))
		game.Log("Navigated orb %v with cmd %v to result %v", cShip, cmd, result)
		game.DrawEntity(result.AsRadius(cShip.GetRadius()), BLACK, 2, NAV_DISPLAY)

		leadCmd := Navigate(lead, result, game, SHIP_RADIUS, -1, -1)
		game.ThrustCommand(leadCmd)

		//find the two flanking positions
		focus := lead.OffsetPolar(float64(leadCmd.Speed), float64(leadCmd.Angle))
		angle := focus.Angle(enemy)
		options := []Entity{focus.OffsetPolar(flankDist, angle-90.0), focus.OffsetPolar(flankDist, angle+90.0)}
		myShips = myShips[1:]

		optionMap := map[int]int{}
		done := map[int]bool{}

		for i, option := range options {
			backProjection := option.OffsetPolar(float64(-lead.NextSpeed), float64(lead.NextAngle))
			game.DrawEntity(backProjection.AsRadius(1), BLUE, 2, NAV_DISPLAY)
			game.DrawEntity(option.AsRadius(SHIP_RADIUS), GREEN, 2, NAV_DISPLAY)
			for _, wingman := range myShips {
				if done[wingman.GetId()] {
					continue
				}
				dist := backProjection.Dist(wingman)
				if _, ok := optionMap[i]; !ok {
					optionMap[i] = wingman.GetId()
					continue
				} else {
					s, _ := game.GetShip(optionMap[i])
					if dist < backProjection.Dist(s) {
						optionMap[i] = wingman.GetId()
					}
				}
			}
			done[optionMap[i]] = true
		}

		for optI, wid := range optionMap {
			wingman, _ := game.GetShip(wid)
			option := options[optI].OffsetTowards(wingman, -1)
			game.ThrustCommand(Navigate(wingman, option, game, SHIP_RADIUS, -1, -1))
		}

		//		//find the wingman closest to an option
		//		var closestWingman *Ship
		//		var closestOption Entity
		//		d := 1000.0
		//		for _, wingman := range myShips[1:] {
		//			for _, option := range options {
		//				oDist := wingman.Dist(option)
		//				if oDist < d {
		//					d = oDist
		//					closestWingman = wingman
		//					closestOption = option
		//				}
		//			}
		//		}

		//		for i, option := range options {
		//			options[i] = options
		//		}

		//		//send closest to closest
		//		game.ThrustCommand(Navigate(closestWingman, closestOption, game, SHIP_RADIUS, -1, -1))
		//		//send other to other
		//		for _, wingman := range myShips[1:] {
		//			for _, option := range options {
		//				if wingman != closestWingman && option != closestOption {
		//					game.ThrustCommand(Navigate(wingman, option, game, SHIP_RADIUS, -1, -1))
		//				}
		//			}
		//		}
	}
}
