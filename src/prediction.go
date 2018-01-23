package src

import (
	"sort"
)

func SimulateEnemyMoves(game Game, includeKite bool) ([]*Ship, map[int]bool) {
	var enemyFront []*Ship
	myFront := make(map[int]bool)

	enemies := game.EnemyShipsUndocked()

	for _, enemy := range enemies {
		enemy.DistanceTo = 1000
		for _, friend := range game.MyShipsUndocked() {
			d := enemy.Dist(friend)
			if enemy.DistanceTo > d {
				enemy.DistanceTo = d
			}
		}
	}

	sort.SliceStable(enemies, func(i, j int) bool { return enemies[i].DistanceTo < enemies[j].DistanceTo })

	for _, enemy := range enemies {
		if SimulateMove(game, enemy, includeKite) {
			enemyFront = append(enemyFront, enemy)
			for _, friend := range enemy.Combatants {
				myFront[friend.GetId()] = true
			}
			game.DrawLine(enemy, enemy.Projection(), GREEN, 2, PREDICTION_DISPLAY)
			game.DrawEntity(enemy.Projection(), GREEN, 0, PREDICTION_DISPLAY)
			game.DrawEntity(enemy.Projection().AsRadius(WEAPON_RADIUS+SHIP_RADIUS*2), LIGHT_GREEN, 2, PREDICTION_DISPLAY)
		}
	}
	return enemyFront, myFront
}

func SimulateMove(game Game, enemy *Ship, includeKite bool) bool {
	friend_ents := ShipsToEntities(game.MyShipsUndocked())
	friend_ents = Filter(friend_ents, func(e Entity) bool { return includeKite || !kite_conditions(game, e.(*Ship)) })
	friend_ents = game.NearestEntities(enemy, friend_ents, -1, MAX_SPEED*2+SHIP_RADIUS*2+WEAPON_RADIUS+BUFFER_TOLERANCE)
	//	panic("Not sure in the replay why ship 0 isn't being picked up, excludeKite flag")

	// game.Log("Targets near enemy: %v/%v...", len(friend_ents), len(game.MyShipsUndocked()))
	if len(friend_ents) > 0 {
		targets := game.NearestEntities(enemy, friend_ents, -1, -1)
		// game.Log("Targets near enemy %v...", enemy)
		//		for _, friend_ent := range targets {
		//			// game.Log("Friend %v (d=%v)", friend_ent, enemy.Dist(friend_ent))
		//		}
		target := targets[0]
		command := enemy.AttackShipCommand(game, target.(*Ship), .1)
		SimulateOrder(command)
		// game.Log("Projection %v for %v", enemy.Projection(), enemy)
		friends := EntitiesToShips(game.NearestEntities(enemy.Projection(), friend_ents, -1, DAMAGABLE_RANGE))
		for _, friend := range friends {
			friend.Combatants = append(friend.Combatants, enemy)
			enemy.Combatants = append(enemy.Combatants, friend)
		}
		game.Log("Combatants of %v: %v", enemy, enemy.Combatants)
		return true
	}
	return false
}

func SimulateOrder(command Command) {
	command.Ship.NextSpeed = command.Speed
	command.Ship.NextAngle = command.Angle
}
