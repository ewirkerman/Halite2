package src

import (
	"fmt"
	"math"
	"sort"

	om "github.com/cevaris/ordered_map"
)

func UseShipsForDefense(objective *Planet, assignableShips *om.OrderedMap, objectiveShips *map[int]*[]int, game Game, heavyCombat bool) {
	enemies := game.GetAttackers(objective, -1, -1)
	//	broadEs := game.EnemyShipsUndocked()
	//	broadEs = EntitiesToShips(Filter(ShipsToEntities(broadEs), func(e Entity) bool {
	//		if EntIndex(ShipsToEntities(enemies), e) >= 0 {
	//			return false
	//		}
	//		nearestPlanet := Nearest(PlanetsToEntities(append(game.EnemyPlanetsOf(e.(*Ship).Owner), objective)), e).(*Planet)
	//		return nearestPlanet.GetId() == objective.GetId()
	//	}))
	//	enemies = append(enemies, broadEs...) // enemies for whom this planet is nearest among this + enemy colonized (for them)
	combaters := game.NearestEntities(objective, ShipsToEntities(game.MyShipsUndocked()), -1, MAX_SPEED*3)
	combaters = Filter(combaters, func(e Entity) bool {
		return e.(*Ship).NextSpeed > 0 && len(e.(*Ship).Combatants) > 0
	})
	assignedShipsId := (*objectiveShips)[objective.GetId()]
	game.Log("Using ships for defense at %v: %v", objective, assignedShipsId)
	assignedShips := game.GetShips(*assignedShipsId)
	game.Log("Using ships for defense at %v: %v", objective, assignedShips)
	assignedShips = EntitiesToShips(game.NearestEntities(objective, ShipsToEntities(assignedShips), -1, -1))

	//	enemies = EntitiesToShips(Filter(ShipsToEntities(enemies), func(e Entity) bool { return len(e.(*Ship).Combatants) < 2 }))
	// game.Log("Defending against unchecked enemies: %v", enemies)as
	numberRequired := len(enemies) + objective.OpenSpots()

	// start by assuming everything assigned will be returned
	extra_ents := Filter(ShipsToEntities(assignedShips), func(e Entity) bool { _, ok := objective.usedShips[e.GetId()]; return !ok })
	extras := EntitiesToShips(extra_ents)

	//	game.Log("Objective %v needs %v ships, has used %v and has %v left to pull from", objective, numberRequired, len(objective.usedShips), len(extras))

	// Defend the planet from attackers
	for i := 0; i < len(enemies)-len(combaters) && len(extras) > 0; i++ {
		//		game.Log("Objective %v defendly needs %v ships, has used %v and has %v left to pull from", objective, len(enemies), len(objective.usedShips), len(extras))
		enemy := enemies[i]

		if len(enemy.Combatants) >= 2 {
			// loop over assigned ships to see if one of the assigned is a combatant
			// if none are combatants, then skip defending against this
			defendWithCombatant := false

			// game.Log("enemy.Combatants: %v", enemy.Combatants)
			// game.Log("assignableShip: %v", assignedShipsId)
			for _, ship := range enemy.Combatants {
				// game.Log("assignableShip: %v", ship)
				for _, sid := range *assignedShipsId {
					if sid == ship.GetId() {
						// game.Log("assignableShip in combatants: %v", ship)
						defendWithCombatant = true
						break
					}
				}
			}

			if !defendWithCombatant {
				//				game.Log("Skipping enemy because the assignees are not his combatants")
				//				numberRequired -= 1
				continue
			}
		}

		if len(game.dockMap[objective.GetId()]) > 0 {
			// find the closest docker to the enemy
			closestDocker := game.NearestEntities(enemy, ShipsToEntities(game.dockMap[objective.GetId()]), 1, -1)[0].(*Ship)
			// find the defender closest to that line (with a tendency towards to docker)
			sort.SliceStable(extras, func(i, j int) bool {
				return closestDocker.Dist(extras[i])*2+extras[i].Dist(enemy) < closestDocker.Dist(extras[j])*2+extras[j].Dist(enemy)
			})
			// the first in that list is the defender we want to use
			ship := extras[0]
			extras = extras[1:]
			if !objective.usedShips[ship.Id] {
				ship.DefendShipFromShip(game, closestDocker, enemy)
				objective.usedShips[ship.Id] = true
			}
		} else {
			ship := extras[0]
			extras = extras[1:]
			if !objective.usedShips[ship.Id] {
				ship.DefendPlanet(game, *objective, enemies)
				objective.usedShips[ship.Id] = true
			}
		}
	}

	// Attempt to fill open spots on the planet
	for len(objective.usedShips) < numberRequired && len(extras) > 0 && objective.OpenSpots() > 0 {
		//		game.Log("Objective %v needs %v ships, has used %v and has %v left to pull from", objective, numberRequired, len(objective.usedShips), len(extras))
		ship := extras[0]
		if heavyCombat && len(objective.usedShips)+objective.DockedShips >= 1 && !WithinDistance(ship, objective, DOCKING_RADIUS+objective.GetRadius()) {
			//			game.Log("Done with %v because of heavy combat, I'm already sending one/have 1 docked and I'm not close enough", objective)
			objective.ObjectiveMet = true
			break
			// check if the planet will fill itself up before you get there (with a buffer for sensibility)
		} else if numberRequired <= 1 && TurnsTo(game, []*Ship{ship}, *objective, DOCKING_RADIUS)+3 > TurnsToNextShip(game, []*Ship{}, *objective, KITE) {
			//			game.Log("%v will be full before %v ship could get there, or shortly after - objective met", objective, ship)
			objective.ObjectiveMet = true
			break
		}
		extras = extras[1:]

		if !objective.usedShips[ship.Id] {
			//			game.Log("Using %v to dock at %v", ship, objective)
			ship.MoveToDock(game, *objective)
			objective.usedShips[ship.Id] = true
		}
	}

	if len(objective.usedShips) >= numberRequired {
		objective.ObjectiveMet = true
	}

	game.Log("Returning extra ships to the pool: %v", extras)
	for _, extra := range extras {
		UnassignShip(objectiveShips, game, extra)
		assignableShips.Set(extra.GetId(), extra)
	}

	if len(objective.usedShips) == numberRequired {
		objective.ObjectiveMet = true
	} else {
		// game.Log("Unmet objective of planet %v with ships %v", objective, objective.usedShips)
	}

}

func UseShipsForOffense(objective *Planet, objectiveShips *map[int]*[]int, game Game) {
	planetDamage := 0
	assignedShipsId := (*objectiveShips)[objective.GetId()]
	assignedShips := make([]*Ship, len(*assignedShipsId))
	for i, shipId := range *assignedShipsId {
		s, _ := game.GetShip(shipId)
		assignedShips[i] = s
	}
	assignedShips = EntitiesToShips(game.NearestEntities(objective, ShipsToEntities(assignedShips), -1, -1))
	for _, ship := range assignedShips {
		if WithinDistance(ship, objective, float64(MAX_SPEED*2)) {
			planetDamage += ship.HP
		}
	}
	//	localFriendlies := game.NearestEntities(objective, ShipsToEntities(game.MyShipsUndocked()), -1, DOCKING_RADIUS+objective.GetRadius()+SHIP_RADIUS)
	//	localEnemies := game.NearestEntities(objective, ShipsToEntities(game.EnemyShips()), -1, DOCKING_RADIUS+objective.GetRadius()+SHIP_RADIUS)
	if PLANET_DESTRUCTION && planetDamage > objective.HP /*&& len(localEnemies) > len(localFriendlies) */ {
		for _, ship := range assignedShips {
			if objective.usedShips[ship.Id] {
				ship.DestroyPlanet(game, *objective)
				objective.usedShips[ship.Id] = true
			}
		}
	} else {
		for _, ship := range assignedShips {
			if !objective.usedShips[ship.Id] {
				ship.AttackPlanet(game, *objective)
				objective.usedShips[ship.Id] = true
			}
		}
	}

}

type Triple struct {
	s *Ship
	p Planet
	v float64
}

func (t Triple) String() string {
	return fmt.Sprintf("[s: %v, p: %v, v: %v]", t.s.Id, t.p.Id, t.v)
}

func EnsureExpansion(game Game, assignableShips *om.OrderedMap, shipObjectiveMap *map[int]*[]int) {
	triples := []Triple{}

	tripleShips := game.MyShipsUndocked()

	for _, planet := range game.PlanetsOwnedBy(game.pid) {
		//check if a theoretical ship at the spawn point would go to an expansion
		game.Log("Fake spawning test ship to see if it would go to an expansion")
		spawnPoint := planet.SpawnPoint
		testShip := Ship{Id: -1, X: spawnPoint.GetX(), Y: spawnPoint.GetY(), Birth: game.turn + 1, Owner: game.pid}
		GenerateObjectives(&testShip, game)
		tripleShips = append(tripleShips, &testShip)
	}

	for _, ship := range tripleShips {
		if len(ship.Combatants) < 1 {
			//			game.Log("Checking for the triple of ship %v", ship)
			objs := *ship.Objectives
			for len(objs) > 0 && objs[0].(Planet).ObjectiveMet {
				objs = objs[1:]
			}
			for _, objective := range objs {
				//				game.Log("Checking remaining objective: %v", objective)
				if p, _ := objective.(Planet); p.IsSafeExpandable(game) {
					triples = append(triples, Triple{s: ship, p: p, v: p.Value})
					break
				} else {
					//					game.Log("%v isn't a triple because it's not safely expandable", objective)
				}
			}
		} else {
			//			game.Log("No triple because it's in combat")
		}
	}

	if len(triples) < 1 {
		return
	}

	sort.Slice(triples, func(i, j int) bool {
		return triples[i].v < triples[j].v || (math.Abs(triples[i].v-triples[j].v) < 1.0 && triples[i].s.Birth < triples[j].s.Birth)
	})
	//	game.LogEach("Triples", triples)

	safeExpanding := false
	for _, p := range game.AllPlanets() {
		if p.IsSafeExpandable(game) {
			mapItself := *shipObjectiveMap
			shipListRef := mapItself[p.GetId()]
			if shipListRef != nil {
				listItself := *shipListRef
				if len(listItself) > 0 {
					safeExpanding = true
					//					game.Log("Found ships heading to %v", p)
					break
				}
			}
		}
	}

	if !safeExpanding {
		//		game.Log("Ah! I'm not safeExpanding to anything!")
		if len(triples) > 0 {
			triple := triples[0]
			if !game.IsCurrentOrderThrust(triple.s) {
				return
			}
			if triple.s.Birth > game.turn {
				return // this means we are waiting for a real ship to instantiate the fake one we are expecting
			}

			if WillCollideNeighbors(game, triple.s) {
				return
			}

			for len(*triple.s.Objectives) > 0 && (*triple.s.Objectives)[0].GetId() != triple.p.GetId() {
				objs := (*triple.s.Objectives)[1:]
				triple.s.Objectives = &objs
			}
			game.Unthrust(triple.s)
			UnassignShip(shipObjectiveMap, game, triple.s)
			AssignToBestObjective(shipObjectiveMap, game, triple.s)
			ProcessObjective(game, &triple.p, assignableShips, shipObjectiveMap, false)

			//			for _, triple := range triples {
			//				game.Log("Found the first triple that was a safe expandable: %v", triple)
			//				if game.IsCurrentOrderThrust(triple.s) {
			//					continue
			//				}
			//				for len(*triple.s.Objectives) > 0 {
			//					//						game.Log("Popped until %v", (*triple.s.Objectives)[0])
			//					game.Unthrust(triple.s)
			//					UnassignShip(shipObjectiveMap, game, triple.s)
			//					AssignToBestObjective(shipObjectiveMap, game, triple.s)
			//					ProcessObjective(game, &triple.p, assignableShips, shipObjectiveMap, false)
			//					return
			//				}
			//				objs := (*triple.s.Objectives)[1:]
			//				triple.s.Objectives = &objs

			//			}
		} else {
			//			game.Log("No triples!")
		}
	}
}

func AssignShipsToObjectives(game Game, assignableShips *om.OrderedMap, shipObjectiveMap *map[int]*[]int) {
	iter := assignableShips.IterFunc()
	// game.Log("Generating all objectives for my ships!")
	myShips := make([]*Ship, assignableShips.Len())
	i := 0

	//Get a list of my ships by best unfull planet score
	var leadObjectives *[]Entity
	for kv, ok := iter(); ok; kv, ok = iter() {
		ship := kv.Value.(*Ship)
		myShips[i] = ship
		i += 1
		assignableShips.Delete(kv.Key)
		if ship.Objectives == nil {
			if game.rushThisTurn {
				panic("Please follow the leader if we are being rushed, so I don't get an underful planet")
				if leadObjectives == nil {
					GenerateObjectives(ship, game)
					leadObjectives = ship.Objectives
				} else {
					objs := make([]Entity, len(*leadObjectives))
					copy(objs, *leadObjectives)
					ship.Objectives = &objs
				}
			} else {
				GenerateObjectives(ship, game)
			}
		}
	}

	iter = assignableShips.IterFunc()
	// game.Log("Assigning all my ships to objectives!")
	for _, ship := range myShips {
		// game.Log("Trying to assign if not kite: %v", ship)
		if !kite_conditions(game, ship) || !kite_actions(game, ship) {
			//			game.Log("Assigned Ship %v - remaining unassigned: %v", ship, *assignableShips)
			AssignToBestObjective(shipObjectiveMap, game, ship)
		}
	}
}

func UnassignShip(shipObjectiveMap *map[int]*[]int, game Game, ship *Ship) {
	for pid, ships := range *shipObjectiveMap {
		for i, sid := range *ships {
			if ship.Id == sid {

				game.Log("Unassigning ship %v from objective %v", ship, pid)
				(*ships)[i] = (*ships)[len(*ships)-1]
				shorter := (*(*shipObjectiveMap)[pid])[:len(*ships)-1]
				(*shipObjectiveMap)[pid] = &shorter
				return
			}
		}
	}
}

func AssignToBestObjective(shipObjectiveMap *map[int]*[]int, game Game, ship *Ship) {
	shipId := ship.GetId()
	game.Log("%v has objectives %v", ship, ship.Objectives)
	for len(*ship.Objectives) > 0 {
		p := (*ship.Objectives)[0].(Planet)
		planetId := p.GetId()
		realPlanet, _ := game.GetPlanet(planetId)
		newObjectives := (*ship.Objectives)[1:]
		ship.Objectives = &newObjectives
		if realPlanet.ObjectiveMet {
			game.Log("Skipping %v because it's objectives are already met", p)
			continue
		}

		//game.Log("This should probably just be assign unless met or safely full")
		// Retrieve or create and retrieve the list of ships for the planet
		ships_ptr, _ := (*shipObjectiveMap)[planetId]
		var ships []int
		if ships_ptr == nil {
			ships = make([]int, 0)
		} else {
			ships = *ships_ptr
		}

		if !p.Owned { // Neutral, so it's an objective
			ships = append(ships, shipId)
		} else if p.Owner == ship.Owner { // Mine
			if game.finishConditions {
				continue
			}
			enemies := game.GetAttackers(p, -1, -1)
			if len(enemies) > 0 { // Mine under attack
				ships = append(ships, shipId)
			} else if !p.IsFull() { // Mine needs to be filled
				ships = append(ships, shipId)
			} else { // Mine safely full
				continue
			}
		} else { // Enemy, we'll attack it in some way
			ships = append(ships, shipId)
		}
		(*shipObjectiveMap)[planetId] = &ships
		game.Log("Assigned %v to planet %v", ship, p)
		break
	}

}

func InHeavyCombat(game Game) bool {
	fighterCount := len(Filter(ShipsToEntities(game.MyShipsUndocked()), func(e Entity) bool { return len(e.(*Ship).Combatants) > 0 }))

	// game.LogOnce("Heavy Combat is limited to OVER a certain ship count")
	return len(game.MyShipsUndocked()) > 12 && float64(fighterCount)/float64(len(game.MyShipsUndocked())) > HEAVY_COMBAT_PROPORTION
}

func UseAssignedShips(game Game, assignableShips *om.OrderedMap, shipObjectiveMap *map[int]*[]int) {
	game.Log("Ship objectives: %v", shipObjectiveMap)
	keys := make([]int, len(*shipObjectiveMap))
	i := 0
	for objId, _ := range *shipObjectiveMap {
		keys[i] = objId
	}

	heavyCombat := InHeavyCombat(game)

	sort.SliceStable(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for objId, _ := range *shipObjectiveMap {
		objective, _ := game.GetPlanet(objId)
		ships := (*shipObjectiveMap)[objId]
		if objective.ObjectiveMet || len(*ships) < 1 {
			//			game.Log("ObjectiveMet: %v, len(ships): %v", objective.ObjectiveMet, len(*ships))
			continue
		}
		//		game.LogEach(fmt.Sprintf("Ships assigned to unmet objective %v:", objective), *ships)
		ProcessObjective(game, objective, assignableShips, shipObjectiveMap, heavyCombat)
	}
	//	game.Log("After using ships, the following are still looking for assignments: %v", assignableShips)
}

func ProcessObjective(game Game, objective *Planet, assignableShips *om.OrderedMap, shipObjectiveMap *map[int]*[]int, heavyCombat bool) {

	if objective.Owned {
		if objective.Owner == game.pid { // mine
			UseShipsForDefense(objective, assignableShips, shipObjectiveMap, game, heavyCombat)
		} else { // enemy
			UseShipsForOffense(objective, shipObjectiveMap, game)
		}
	} else {
		UseShipsForDefense(objective, assignableShips, shipObjectiveMap, game, heavyCombat)
	}
}

func GenerateObjectives(ship *Ship, game Game) []Entity {
	// game.Log("Generating objectives for ship %v:", ship)
	//	game.Log("All plaents: %v", game.AllPlanets())
	objectives := make([]*Planet, 0)

	//	game.Log("Planet_ents: %v", PlanetsToEntities(objectives))
	neutrals := EntitiesToPlanets(Filter(PlanetsToEntities(game.AllPlanets()), func(e Entity) bool { return !e.(*Planet).Owned }))

	for _, planet := range game.AllPlanets() {
		//game.Log("Doing planet %v: %v", planet)
		shipDist := ship.Dist(planet)
		size := float64(MAX_SPEED * planet.DockingSpots)
		centerDist := 0.0
		if game.CurrentPlayers() > 2 && len(game.ShipsOwnedBy(-2)) <= 0 {
			centerDist = planet.Dist(Point{X: float64(game.Width()) / 2, Y: float64(game.Height()) / 2})
		}
		ownerFactor := 0.0
		if planet.Owner == -2 {
			//			game.Log("Skipping planet %v because it's owned by an ally", planet)
			continue // ally planets never make it into my objective list
		} else if planet.Owned && planet.Owner != game.Pid() {
			//			if planet.Owner != game.weakest {
			//				continue // only attack the planets of the weakest player (intentionally)
			//			}
			//no additional bonus for being an enemy planet
		} else {
			ownerFactor = 1.0 * MAX_SPEED * 3

			// every neutral gets this subtracted from their neutral score because a larger neutral score means a planet is better
			maxDist := 7.0 * MAX_SPEED
			ownerFactor += maxDist

			// if there is another neutral within maxDist range, we'll damage the neutral score by the distance
			// else, we'll damage it by the full maxDist penalty
			nearestNeutrals := game.NearestEntities(planet, PlanetsToEntities(neutrals), 1, maxDist)
			if len(nearestNeutrals) > 0 {
				ownerFactor -= planet.Dist(nearestNeutrals[0])
			} else {
				ownerFactor -= maxDist
			}
			//				game.Log("neutral valuation of %v for %v (higher is more attractive): %v", planet, ship, neutral)
			if planet.Owner == game.Pid() && WithinDistance(ship, planet, DOCKING_RADIUS+planet.GetRadius()) && planet.OpenSpots() > 0 {
				game.Log("%v could dock right now at %v!", ship, planet)
				ownerFactor *= 2.0
			}
		}

		planet.Value = shipDist*2 - size*MAX_SPEED/2 - centerDist - ownerFactor
		//		game.Log("value of planet %v (Owner %v): %v", planet, planet.Owner, planet.Value)
		objectives = append(objectives, planet)
	}

	//game.Log("Pre-sort: %v", objectives)

	sort.Sort(ByValue(objectives))
	objectiveEnts := make([]Entity, len(objectives))
	for i, obj := range objectives {
		objectiveEnts[i] = *obj
	}

	//	objectiveEnts := PlanetsToEntities(objectives)
	ship.Objectives = &objectiveEnts
	//	game.Log("Generated objectives for ship %v: %v", ship, ship.Objectives)
	return objectiveEnts
}

func CheckForPlanetRetreat(game Game, pid int) {
	planet, _ := game.GetPlanet(pid)
	if planet.Owned && planet.Owner == game.pid && !planet.ObjectiveMet {
		TryUndocking(game, planet)
	}
}

type ByValue []*Planet

func (s ByValue) Len() int           { return len(s) }
func (s ByValue) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByValue) Less(i, j int) bool { return s[i].Value < s[j].Value }
