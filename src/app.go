package src

import (
	"flag"
	"fmt"
	"io"
	//	"io/ioutil"
	//	"net/http"
	"os"
	//	"os"
	"runtime/debug"
	//	"runtime/pprof"
	//	"strconv"
	"time"

	"runtime/pprof"

	ordered_map "./github.com/cevaris/ordered_map"
)

const (
	NAME    = "UltimateBot"
	VERSION = "1"
)

var thisGame *Game

func Run() {
	debug := flag.Bool("debug", false, "Save files of the map at every turn")
	//	web := flag.Bool("web", false, "Save files of the map at every turn")
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()
	if *cpuprofile != "" {
		f, _ := os.Create(*cpuprofile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	game := new(Game)
	var pr io.Reader
	pr = os.Stdin
	InitGame(*debug, game, pr)
	game.StartLog(fmt.Sprintf("log%d.txt", game.Pid()))
	game.LogWithoutTurn("--------------------------------------------------------------------------------")
	game.LogWithoutTurn("%s %s starting up at %s", NAME, VERSION, time.Now().Format("2006-01-02T15:04:05Z"))

	fmt.Printf("%s %s\n", NAME, VERSION)
	thisGame = game
	// game.Log("Trying Web Listener")

	//	if *web {
	//		tr, pw := io.Pipe()
	//		// game.Log("Starting Web Listener")
	//		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//			thisGame.Log("Forwarding HTTP request to STDin")
	//			body, err := ioutil.ReadAll(r.Body)
	//			if err == nil {
	//				thisGame.Log("Error: %v", err)
	//			}
	//			fmt.Fprintf(w, string(body))
	//			thisGame.Log("Returning: %v", string(body))
	//			fmt.Fprintf(pw, string(body))
	//			thisGame.Log("To StdIn: %v", string(body))

	//		})
	//		go http.ListenAndServe(":8080", nil)
	//		pr = tr
	//	} else {
	//	}
	//	// game.Log("Building game")
	//	InitGame(*debug, game, pr)

	for {
		thisGame.Parse()
		fantasticUnbeatableAI(thisGame)
		thisGame.Send()
		thisGame.ShowMap()
	}

}

func fantasticUnbeatableAI(game *Game) {
	defer func() {
		if x := recover(); x != nil {
			game.Log("%v", x)
			if x != "ABORT" {
				panic(x)
				debug.PrintStack()
			}
		}
	}()

	//	game.LogEach("Showing all ships:", game.AllShips())

	assignableShips := ordered_map.NewOrderedMap()
	for _, ship := range game.MyShipsUndocked() {
		assignableShips.Set(ship.GetId(), ship)
	}

	if RushConditions(*game) && Rush(*game) {
		return
	}
	if leader, ok := fightForSecondConditions(*game); ok {
		game.Log("Fight for second!")
		leader = 0
		iter := assignableShips.IterFunc()
		ships := make([]*Ship, assignableShips.Len())
		i := 0

		for kv, ok := iter(); ok; kv, ok = iter() {
			ship := kv.Value.(*Ship)
			ships[i] = ship
			i++
		}
		// game.Log("Murder ships: %v", ships)
		fightForSecond(*game, leader, ships)
		return
	}

	enemyFront, myFront := SimulateEnemyMoves(*game, false)
	//game.Log("Enemies: %v",enemyFront)
	//game.Log("Mine: %v",myFront)
	if !FinishConditions(game) || game.rushThisTurn {
		shipsUsedByCombat := ResolveCombatSimple(*game, enemyFront, &myFront)
		// game.Log("Front ships used by combat: %v", shipsUsedByCombat)

		// take the ships used in combat out of those assignable
		for shipId, _ := range shipsUsedByCombat {
			// game.Log("Removed Ship %v from the set of assignable ships", shipId)
			assignableShips.Delete(shipId)
		}
	}

	shipObjectiveMap := map[int]*[]int{}

	if (len(game.PlanetsOwnedBy(game.pid))+len(game.PlanetsOwnedBy(-2)) == len(game.AllPlanets())) ||
		(len(game.PlanetsOwnedBy(-2)) == 0 && len(game.ShipsOwnedBy(-2)) > 0 && len(game.PlanetsOwnedBy(game.pid))+1 == len(game.AllPlanets())) {
		iter := assignableShips.IterFunc()
		myShips := make([]*Ship, assignableShips.Len())
		i := 0

		for kv, ok := iter(); ok; kv, ok = iter() {
			ship := kv.Value.(*Ship)
			myShips[i] = ship
			i++
		}

		game.Log("Using for murder: %v", myShips)
		murder(*game, -2, myShips)
	} else {
		for assignableShips.Len() > 0 {
			//			game.Log("Assignable ships (%v): %v:", assignableShips.Len(), assignableShips)
			// game.Log("Ship objectives: %v:", shipObjectiveMap)
			AssignShipsToObjectives(*game, assignableShips, &shipObjectiveMap)
			UseAssignedShips(*game, assignableShips, &shipObjectiveMap)
		}
		if len(game.MyPlanets()) > 0 {
			EnsureExpansion(*game, assignableShips, &shipObjectiveMap)
		}

	}

	for _, ship := range game.MyShipsUndocked() {
		if WillCollideNeighbors(*game, ship) {
			game.Log(fmt.Sprintf("%v is on a collision course!", ship))
		}
	}
}

func FinishConditions(game *Game) bool {
	game.Log("Checking finish conditions...")
	game.finishConditions = true
	for pid := 0; pid < game.InitialPlayers(); pid++ {
		if pid == game.pid {
			continue
		}
		game.Log("3x Them %v vs. 2x Mine %v", len(game.ShipsOwnedBy(pid))*3, len(game.MyShips())*2)
		if len(game.ShipsOwnedBy(pid))*3 >= len(game.MyShips())*2 {
			game.finishConditions = false
		}
	}
	if game.finishConditions {
		game.Log("We're skipping combat so we can finish it faster")
	}
	return game.finishConditions
}
