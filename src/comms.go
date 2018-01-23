package src

// Fohristiwhirl, November 2017.
// A sort of alternative starter kit.

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------

type TokenParser struct {
	scanner    *bufio.Scanner
	count      int
	all_tokens []string // This is used for logging only. It is cleared each time it's asked-for.
}

func NewTokenParser(pr io.Reader) *TokenParser {
	ret := new(TokenParser)
	ret.scanner = bufio.NewScanner(pr)
	ret.scanner.Split(bufio.ScanWords)
	return ret
}

func (self *TokenParser) Int() int {
	bl := self.scanner.Scan()
	if bl == false {
		err := self.scanner.Err()
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		} else {
			panic(fmt.Sprintf("End of input."))
		}
	}
	self.all_tokens = append(self.all_tokens, self.scanner.Text())
	ret, err := strconv.Atoi(self.scanner.Text())
	if err != nil {
		panic(fmt.Sprintf("TokenReader.Int(): Atoi failed at token %d: \"%s\"", self.count, self.scanner.Text()))
	}
	self.count++
	return ret
}

func (self *TokenParser) DockedStatus() DockedStatus {
	return DockedStatus(self.Int())
}

func (self *TokenParser) Float() float64 {
	bl := self.scanner.Scan()
	if bl == false {
		err := self.scanner.Err()
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		} else {
			panic(fmt.Sprintf("End of input."))
		}
	}
	self.all_tokens = append(self.all_tokens, self.scanner.Text())
	ret, err := strconv.ParseFloat(self.scanner.Text(), 64)
	if err != nil {
		panic(fmt.Sprintf("TokenReader.Float(): ParseFloat failed at token %d: \"%s\"", self.count, self.scanner.Text()))
	}
	self.count++
	return ret
}

func (self *TokenParser) Bool() bool {
	val := self.Int()
	if val != 0 && val != 1 {
		panic(fmt.Sprintf("TokenReader.Bool(): Value wasn't 0 or 1 (was: \"%d\")", val))
	}
	return val == 1
}

func (self *TokenParser) Tokens(sep string) string {
	ret := strings.Join(self.all_tokens, sep)
	self.all_tokens = nil
	return ret
}

func (self *TokenParser) ClearTokens() {
	self.all_tokens = nil
}

// ---------------------------------------

func (g *Game) Parse() {

	g.orders = make(map[int]string) // Clear all orders.

	if g.inited {
		g.turn++
	}

	// Clear some info maps. We will recreate them during parsing.

	old_shipmap := g.shipMap // We need last turn's *Ship info for inferring birth.

	g.shipMap = make(map[int]*Ship)
	g.planetMap = make(map[int]*Planet)
	g.dockMap = make(map[int][]*Ship)
	g.playershipMap = make(map[int][]*Ship)

	// Player parsing.............................................................................

	player_count := g.token_parser.Int()

	g.parse_time = time.Now() // MUST happen AFTER the first token parse. <------------------------------------- important

	if g.initialPlayers == 0 {
		g.initialPlayers = player_count // Only save this at init stage.
	}

	players_with_ships := 0

	for p := 0; p < player_count; p++ {

		pid := g.token_parser.Int()

		ship_count := g.token_parser.Int()

		if ship_count > 0 {
			players_with_ships++
		}

		for s := 0; s < ship_count; s++ {

			var ship Ship

			sid := g.token_parser.Int()
			ship.Id = sid

			ship.Owner = pid
			ship.X = g.token_parser.Float()
			ship.Y = g.token_parser.Float()
			ship.HP = g.token_parser.Int()
			ship.Radius = SHIP_RADIUS
			g.token_parser.Float() // Skip deprecated "speedx"
			g.token_parser.Float() // Skip deprecated "speedy"
			ship.DockedStatus = g.token_parser.DockedStatus()
			ship.DockedPlanet = g.token_parser.Int()

			if ship.DockedStatus == UNDOCKED {
				ship.DockedPlanet = -1
			}

			ship.DockingProgress = g.token_parser.Int()
			g.token_parser.Int() // Skip deprecated "cooldown"

			oldShip, ok := old_shipmap[sid]

			if ok == false {
				ship.Birth = Max(0, g.turn) // Turn can be -1 in init stage.
				if g.playerShipsetMap[pid] == nil {
					g.playerShipsetMap[pid] = make(map[int]bool)
				}
				g.playerShipsetMap[pid][sid] = true
			} else {
				ship.LastLocation = oldShip.AsRadius(0)
				ship.LastSpeed = Round(oldShip.Dist(&ship))
				ship.LastAngle = Round(oldShip.Angle(&ship))
			}

			g.shipMap[sid] = &ship
			g.playershipMap[pid] = append(g.playershipMap[pid], &ship)
		}

		sort.Slice(g.playershipMap[pid], func(a, b int) bool {
			return g.playershipMap[pid][a].Id < g.playershipMap[pid][b].Id
		})
	}

	// Planet parsing.............................................................................

	planet_count := g.token_parser.Int()

	for p := 0; p < planet_count; p++ {

		var planet Planet

		plid := g.token_parser.Int()
		planet.Id = plid

		planet.X = g.token_parser.Float()
		planet.Y = g.token_parser.Float()
		planet.HP = g.token_parser.Int()
		planet.Radius = g.token_parser.Float()
		planet.DockingSpots = g.token_parser.Int()
		planet.CurrentProduction = g.token_parser.Int()
		g.token_parser.Int() // Skip deprecated "remaining production"
		planet.Owned = g.token_parser.Bool()
		planet.Owner = g.token_parser.Int()
		planet.usedShips = map[int]bool{}
		planet.SpawnPoint = planet.CalcSpawnPoint(*g)

		if planet.Owned == false {
			planet.Owner = -1
		}

		planet.DockedShips = g.token_parser.Int()

		// The dockMap is kept separately so that the Planet struct has no mutable fields.
		// i.e. the Planet struct itself does not get the following data:

		for s := 0; s < planet.DockedShips; s++ {

			// This relies on the fact that we've already been given info about the ships...

			sid := g.token_parser.Int()
			ship, ok := g.GetShip(sid)
			if ok == false {
				panic("Parser choked on GetShip(sid)")
			}
			g.dockMap[plid] = append(g.dockMap[plid], ship)
		}
		sort.Slice(g.dockMap[plid], func(a, b int) bool {
			return g.dockMap[plid][a].Id < g.dockMap[plid][b].Id
		})

		g.planetMap[plid] = &planet
	}

	// Look for allies on the first turn only
	if ALLY_BEE_DANCE {
		if g.turn == 1 {
			for i := 0; i < g.CurrentPlayers(); i++ {
				if i != g.pid && g.CheckBeeDance(i) {
					g.allyPidStages[i] = 1
				}
			}
		}
	}

	g.Log("Found ally players: %v", g.allyPidStages)

	// Check if we maintain the alliance
	breakAlliances := true
	for i := 0; i < g.InitialPlayers() && len(g.allyPidStages) > 0; i++ {
		if i == g.pid {
			continue
		}
		// check each living player
		if len(g.ShipsOwnedBy(i)) > 0 {
			g.Log("Player %v is alive: %v ships", i, len(g.ShipsOwnedBy(i)))
			g.Log("Allies: %v", g.allyPidStages)
			isAlly := false

			// check if they are an ally
			for pid, _ := range g.allyPidStages {
				if i == pid {
					isAlly = true
					break
				}
			}

			// if there are any living non-allies, keep the alliance
			if !isAlly {
				g.Log("Player %v is not an ally - continuing alliance", i)
				breakAlliances = false
				break
			} else {
				g.Log("Player %v is an ally", i)
			}
		}
	}

	if breakAlliances {
		g.Log("Breaking alliance with: %v", g.allyPidStages)
		g.allyPidStages = map[int]int{}
	}

	// For each ally, mark them as -2
	for pid, _ := range g.allyPidStages {
		g.MakeAlly(pid)
	}

	// Query responses (see info.go)... while these could be done interleaved with the above, they are separated for clarity.

	g.all_ships_cache = nil
	for _, ship := range g.shipMap {
		g.all_ships_cache = append(g.all_ships_cache, ship)
	}
	sort.Slice(g.all_ships_cache, func(a, b int) bool {
		return g.all_ships_cache[a].Id < g.all_ships_cache[b].Id
	})

	g.enemy_ships_cache = nil
	for _, ship := range g.shipMap {
		if ship.Owner != g.pid && ship.Owner >= 0 {
			g.enemy_ships_cache = append(g.enemy_ships_cache, ship)
		}
	}
	sort.Slice(g.enemy_ships_cache, func(a, b int) bool {
		return g.enemy_ships_cache[a].Id < g.enemy_ships_cache[b].Id
	})

	g.all_planets_cache = nil
	for _, planet := range g.planetMap {
		g.all_planets_cache = append(g.all_planets_cache, planet)
	}
	sort.Slice(g.all_planets_cache, func(a, b int) bool {
		return g.all_planets_cache[a].Id < g.all_planets_cache[b].Id
	})

	g.weakest = g.WeakestEnemy()

	// Some meta info...

	g.currentPlayers = players_with_ships
	g.raw = g.token_parser.Tokens(" ")
}

// ---------------------------------------
func (g *Game) MakeAlly(pid int) {
	for sid, ship := range g.shipMap {
		if stage, ok := g.allyPidStages[ship.Owner]; ok && stage > 0 {
			g.shipMap[sid].Owner = -2 //indicates an ally
			g.shipMap[sid].Radius = SHIP_RADIUS + WEAPON_RADIUS
		}
	}
	for pid, planet := range g.planetMap {
		if stage, ok := g.allyPidStages[planet.Owner]; ok && stage > 0 {
			g.planetMap[pid].Owner = -2
		}
	}
	if _, ok := g.playershipMap[-2]; !ok {
		g.playershipMap[-2] = make([]*Ship, 0)
	}
	g.playershipMap[-2] = append(g.playershipMap[-2], g.playershipMap[pid]...)

	if len(g.ShipsDockedOwnedBy(-2)) < 1 {
		for _, ship := range g.ShipsOwnedBy(-2) {
			ship.Radius = SHIP_RADIUS + WEAPON_RADIUS + MAX_SPEED
		}
	}
	delete(g.playershipMap, pid)
}

//func (g *Game) ConfirmBeeDance(pid int) bool {
//	for _, ship := range g.ShipsOwnedBy(pid) {
//		// confirmation is if the last ship
//		if ModInt(ship.LastAngle, 9) != ModInt(g.pid+ModInt(ship.GetId(), 3)+3*g.pid, 9) {
//			return false
//		}
//	}
//	return true
//}

func (g *Game) CheckBeeDance(pid int) bool {
	g.Log("Checking honey bee dance for player %v", pid)
	// for the (0, 1, 2, 3)rd player the (0, 1, 2, 0)th ship must move only 6
	if ship, ok := g.GetShip(pid*3 + int(math.Mod(float64(pid), 3.0))); !ok || ship.LastSpeed != 6 {
		g.Log("%v is not my bee", pid)
		return false
	}

	// all ship angles mod 9 must equal pid+eid mod 9
	for _, ship := range g.ShipsOwnedBy(pid) {
		if ModInt(ship.LastAngle, BEE_BANDWIDTH) != ModInt(pid+ship.GetId(), BEE_BANDWIDTH) {
			return false
		}
	}

	g.Log("%v is MY HONEY BEE!", pid)
	return true
}

func (g *Game) Unthrust(ship *Ship) {
	ship.NextSpeed = 0
	ship.NextAngle = 0
	delete(g.orders, ship.GetId())
}

func (g *Game) IsCurrentOrderThrust(ship *Ship) bool {
	return strings.Index(g.orders[ship.GetId()], "t") > -1
}

func (g *Game) Thrust(ship *Ship, speed, degrees int) {
	if ship.DockedStatus != UNDOCKED {
		panic(fmt.Sprintf("Issued an order to a docked ship %v!", ship))
	}
	if ship.Owner != g.pid {
		panic(fmt.Sprintf("Issued an order to an enemy ship %v!", ship))
	}
	for degrees < 0 {
		degrees += 360
	}
	degrees %= 360
	ship.NextSpeed = speed
	ship.NextAngle = degrees
	if STACKTRACE {
		trace := make([]byte, 1024)
		count := runtime.Stack(trace, true)
		g.Log("Stack of %d bytes: %s", count, trace)
	}
	tmp := fmt.Sprintf("t %d %d %d", ship.Id, speed, degrees)
	g.Log("Sending order: %v", tmp)
	if order, ok := g.orders[ship.Id]; ok {
		panic("Issued an order to " + ship.String() + " when it already had order " + order)
	}
	g.orders[ship.Id] = tmp
}

func (g *Game) ThrustCommand(command Command) {
	g.DrawLine(command.Ship, command.Ship.OffsetPolar(float64(command.Speed), float64(command.Angle)), RED, 1, ORDER_DISPLAY)
	g.Thrust(command.Ship, command.Speed, command.Angle)
}

func (g *Game) ThrustWithMessage(ship *Ship, speed, degrees int, message int) {
	for degrees < 0 {
		degrees += 360
	}
	degrees %= 360
	if message >= 0 && message <= 180 {
		degrees += (int(message) + 1) * 360
	}
	g.orders[ship.Id] = fmt.Sprintf("t %d %d %d", ship.Id, speed, degrees)
}

func (g *Game) Dock(ship *Ship, planet Planet) {
	g.orders[ship.Id] = fmt.Sprintf("d %d %d", ship.Id, planet.Id)
	trace := make([]byte, 1024)
	count := runtime.Stack(trace, true)
	g.Log("Stack of %d bytes: %s", count, trace)
	g.Log("Sending order: %v", g.orders[ship.Id])
}

func (g *Game) Undock(ship *Ship) {
	g.orders[ship.Id] = fmt.Sprintf("u %d", ship.Id)
}

func (g *Game) ClearOrder(ship *Ship) {
	delete(g.orders, ship.Id)
}

func (g *Game) CurrentOrder(ship *Ship) string {
	return g.orders[ship.Id]
}

func (g *Game) RawOrder(sid int, s string) {
	g.orders[sid] = s
}

func (g *Game) Send() {
	var commands []string
	for _, s := range g.orders {
		commands = append(commands, s)
	}
	out := strings.Join(commands, " ")
	fmt.Printf(out)
	fmt.Printf("\n")
	elapsed := time.Since(g.ParseTime())
	if elapsed > g.longest_time {
		g.longest_time = elapsed
	}
	g.Log("Total turn time: %v (longest: %v)", elapsed, g.longest_time)
}
