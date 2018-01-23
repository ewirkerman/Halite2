package src

import (
	"container/heap"
	"math"
	"time"
)

func Navigate(ship *Ship, target Entity, game Game, enemyRange float64, speed_direct, angle_direct int) Command {
	elapsed := time.Since(game.ParseTime())
	if elapsed.Seconds() > 1.75 {
		panic("ABORT")
	}

	if speed_direct > 0 {
		//		game.Log("Navigating %v at speed %v and angle %v", ship, speed_direct, angle_direct)
		//game.Log("These should really be put into a collision TEST method, rather than a navigate method")
		game.DrawLine(ship, ship.OffsetPolar(float64(speed_direct), float64(angle_direct)), BLACK, 1, NAV_DISPLAY)
	} else {
		game.Log("Navigating %v to target %v", ship, target)
		game.DrawLine(ship, target, BLACK, 1, NAV_DISPLAY)
	}
	//game.Log("Elapsed time since turn start: %v", time.Since(game.ParseTime()))

	distanceToTarget := ship.Dist(target) + POINT_MATCH_TOLERANCE

	speed := MAX_SPEED
	if game.turn < 1 && game.CurrentPlayers() > 2 && ALLY_BEE_DANCE {
		if ship.GetId() == game.pid*3+ModInt(game.pid, 3) {
			speed = 6
		}
	}
	//game.Log("speed %v vs. dist to target %v", speed, distanceToTarget)
	maybeKiteAttackers := game.GetAttackers(ship, -1, WEAPON_RADIUS+MAX_SPEED)
	//game.Log("If this navigator is a kite, then his attackers are %v", maybeKiteAttackers)

	if !ship.IsKite || len(maybeKiteAttackers) < 1 {
		speedFloat := math.Floor(float64(speed) + POINT_MATCH_TOLERANCE)
		if distanceToTarget < speedFloat {
			speed = Min(Round(distanceToTarget), MAX_SPEED)
			game.Log("Reduced speed to %v based on a dist to target of %v", speed, distanceToTarget)
		}
		if speed_direct > 0 {
			speed = speed_direct
		} else {
			speed = int(speed)
		}
	} else {
		//game.Log("Threatened Kite Moves at full speed")
	}

	//game.Log("Navigating with speed: %v", speed)
	var targetAngle float64
	if angle_direct > -1 {
		targetAngle = float64(angle_direct)
	} else {
		targetAngle = float64(int(ship.Angle(target)))
	}

	obstacles := PlanetsToEntities(game.AllPlanets())
	maxPlanetSize := 0.0
	for _, planet := range obstacles {
		maxPlanetSize = MaxFloat(maxPlanetSize, planet.GetRadius())
	}
	obstacles = append(obstacles, ShipsToEntities(game.AllShips())...)
	ally := ShipsToEntities(game.ShipsDockedOwnedBy(-2))
	if len(ally) == 0 {
		ally = ShipsToEntities(game.ShipsUndockedOwnedBy(-2))
	}
	obstacles = append(obstacles, ally...)
	obstacles = game.NearestEntities(ship, obstacles, -1, float64(MAX_SPEED)+SHIP_RADIUS+maxPlanetSize+BUFFER_TOLERANCE)

	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	first := int(targetAngle)
	width := 91
	pq := make(PriorityQueue, 0) //91 to each side, plus the most target-ish angle per layer + 1 for speed 0

	for speed := 0; speed < MAX_SPEED; speed++ {
		for angle := first - width; angle <= first+width; angle += 1 + len(game.MyShipsUndocked())/100 {
			//			game.Log("Generating %v/%v (%v): %v, %v", i, len(pq), j, speed+1, angle)
			a := ModInt(angle, 360)
			if game.turn < 1 && game.CurrentPlayers() > 2 && ALLY_BEE_DANCE {
				mod := GetHoneyBeeMod(game, ship)
				if ModInt(a, BEE_BANDWIDTH) != mod {
					continue
				}
				if ship.GetId() == game.pid*3+ModInt(game.pid, 3) && speed+1 != 6 {
					continue
				}
			}
			pq = append(pq, &Item{
				value:    Command{Ship: ship, Speed: speed + 1, Angle: a},
				priority: ship.OffsetPolar(float64(speed+1), float64(a)).Dist2(target),
				index:    len(pq),
			})
			//			game.Log("Generated %v: %v", pq[len(pq)-1].priority, pq[len(pq)-1].value)
		}
	}
	command := Command{Ship: ship, Speed: 0, Angle: int(targetAngle)}
	pq = append(pq, &Item{value: command, priority: ship.Dist2(target), index: len(pq) - 1})

	//	game.LogEach("Options:", pq)
	heap.Init(&pq)

	// Take the items out; they arrive in decreasing priority order.

	game.Log("Angle to target: %v", int(targetAngle))
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		speed := item.value.Speed
		workingAngle := item.value.Angle
		//		game.Log("Testing %v: %v", item.priority, item.value)
		point := ship.OffsetPolarInt(speed, workingAngle)
		if point.GetX() <= 2*SHIP_RADIUS || point.GetX() >= float64(game.Width())-2*SHIP_RADIUS {
			continue
		}
		if point.GetY() <= 2*SHIP_RADIUS || point.GetY() >= float64(game.Height())-2*SHIP_RADIUS {
			continue
		}
		ok := TestShipAngle(game, ship, speed, workingAngle, obstacles)
		if ok {
			command = item.value
			break
		}
	}

	game.Log("Produced command: %v", command.String())
	return command
}

func TestShipAngle(game Game, ship *Ship, speed, angle int, obstacles []Entity) bool {
	ship.NextSpeed = speed
	ship.NextAngle = angle
	collide := WillCollideSlice(game, ship, obstacles)
	ship.NextSpeed = 0
	ship.NextAngle = 0
	return !collide
}

// An Item is something we manage in a priority queue.
type Item struct {
	value    Command // The value of the item; arbitrary.
	priority float64 // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value Command, priority float64) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}
