package src

import (
	"math"
	"sort"
)

func AmbushCluster(game Game, enemyFront []*Ship, myFront *map[int]bool) {
	game.Log("Starting Ambush Cluster")
	myShips := make([]*Ship, len(*myFront))
	i := 0
	for sid, _ := range *myFront {
		ship, _ := game.GetShip(sid)
		myShips[i] = ship
		i += 1
	}
	game.LogEach("Finding Mean of: ", myShips)
	myMean := game.MeanPoint(ShipsToEntities(myShips))
	// clump up
	// if clumped, nav as orb
	// all match the velocity of the orb
	// take a stab at which way they're going to go, left or right

	game.Log("Making cShip")
	c := make_circle(ShipsToEntities(myShips))
	c.Radius += SHIP_RADIUS
	cShip := &Ship{X: c.X, Y: c.Y, Owner: game.Pid(), Radius: c.Radius}

	game.Log("Made cShip")
	// determine target point
	target := Nearest(ShipsToEntities(enemyFront), cShip)
	if WithinDistanceAny(c, ShipsToEntities(enemyFront), MAX_SPEED*2.0+WEAPON_RADIUS+SHIP_RADIUS+cShip.Radius) {
		enemyMean := game.MeanPoint(ShipsToEntities(enemyFront))
		angleToTarget := cShip.Angle(target)
		angleToMean := cShip.Angle(enemyMean)
		angleOffset := 70.0
		angleOffset *= -1.0 * float64(AngleCW(angleToTarget, angleToMean))
		angleToTarget += angleOffset
		angleToTarget = math.Mod(angleToTarget, 360.0)
		target = cShip.OffsetPolar(7, angleToTarget)
	}

	game.Log("Found target")

	if !WithinDistanceAll(myMean, ShipsToEntities(myShips), .8) {
		Clump(game, myShips, target)
		return
	}

	// send them all to the target point by using all the same commands as the cShip
	cmd := NavigateWithoutShips(cShip, target, game)
	game.Log("Moving as cShip")
	for _, ship := range myShips {
		cmd.Ship = ship
		game.ThrustCommand(cmd)
	}

}

func Clump(game Game, ships []*Ship, target Entity) {
	sort.SliceStable(ships, func(i, j int) bool {
		return ships[i].Dist2(target) < ships[j].Dist2(target)
	})
	game.Log("Clumping on %v", ships[0])
	for _, ship := range ships[1:] {
		cmd := Navigate(ship, ship.ClosestPoint(ships[0], BUFFER_TOLERANCE), game, -1, -1, -1)
		game.ThrustCommand(cmd)
	}
}

func AngleCW(baseAngle, testAngle float64) int {
	testAngle = math.Mod(testAngle, 360)
	baseAngle = math.Mod(baseAngle, 360)

	if baseAngle == testAngle {
		return 0
	} else if baseAngle+180 > testAngle || math.Mod(baseAngle+180, 360) > testAngle {
		return 1
	} else {
		return -1
	}
}
