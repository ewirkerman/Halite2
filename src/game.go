package src

// Fohristiwhirl, November 2017.
// A sort of alternative starter kit.

import (
	"fmt"
	"io"
	"time"

	"github.com/fogleman/gg"
)

type Game struct {
	inited bool
	turn   int
	pid    int // Our own ID
	width  int
	height int

	initialPlayers int // Stored only once at startup. Never changes.
	currentPlayers int

	planetMap        map[int]*Planet      // Planet ID --> Planet
	shipMap          map[int]*Ship        // *Ship ID --> *Ship
	dockMap          map[int][]*Ship      // Planet ID --> *Ship slice
	playershipMap    map[int][]*Ship      // Player ID --> *Ship slice
	playerShipsetMap map[int]map[int]bool // Player ID --> *Ship slice
	allyPidStages    map[int]int
	weakest          int

	orders map[int]string

	logfile      *Logfile
	token_parser *TokenParser
	raw          string // The raw input line sent by halite.exe

	parse_time   time.Time
	longest_time time.Duration

	// These slices are kept as answers to common queries...

	all_ships_cache   []*Ship
	enemy_ships_cache []*Ship
	all_planets_cache []*Planet
	display           *gg.Context
	rushThisTurn      bool
	finishConditions  bool
}

type Command struct {
	OrderType OrderType
	Ship      *Ship
	target    Planet
	Speed     int
	Angle     int
}

func (c Command) String() string { return fmt.Sprintf("[ %v %v %v ]", c.Ship.Id, c.Speed, c.Angle) }

func (c Command) Result() Point {
	return c.Ship.OffsetPolar(float64(c.Speed), float64(c.Angle)).AsRadius(c.Ship.GetRadius())
}

func InitGame(debug bool, game *Game, pr io.Reader) {
	game.turn = -1
	game.playerShipsetMap = make(map[int]map[int]bool)
	game.token_parser = NewTokenParser(pr)
	game.pid = game.token_parser.Int()
	game.width = game.token_parser.Int()
	game.height = game.token_parser.Int()
	game.token_parser.ClearTokens() // This is just clearing the token_parser's "log".
	game.Parse()
	game.inited = true // Just means Parse() will increment the turn value before parsing.
	if debug {
		game.display = CreateDisplayContext(*game)
	}
}

func (g *Game) Turn() int            { return g.turn }
func (g *Game) Pid() int             { return g.pid }
func (g *Game) Width() int           { return g.width }
func (g *Game) Height() int          { return g.height }
func (g *Game) InitialPlayers() int  { return g.initialPlayers }
func (g *Game) CurrentPlayers() int  { return g.currentPlayers }
func (g *Game) ParseTime() time.Time { return g.parse_time }
