package src

import (
	"math"
	"strconv"

	"github.com/fogleman/gg"
)

const (
	DISPLAY_WIDTH   = 2000
	DISPLAY_ENABLED = true
)

type Color struct {
	R, G, B float64
}

var RED = Color{1, 0, 0}
var LIGHT_RED = Color{1, .85, .85}
var BLACK = Color{0, 0, 0}
var BLUE = Color{0, 0, 1}
var LIGHT_BLUE = Color{.85, .85, 1}
var GREEN = Color{0, 1, 0}
var LIGHT_GREEN = Color{.85, 1, .85}
var YELLOW = Color{1, 1, 0}
var ORANGE = Color{1, .5, 0}
var WHITE = Color{1, 1, 1}
var GRAY = Color{.7, .7, .7}
var LIGHT_GRAY = Color{.9, .9, .9}
var PURPLE = Color{1, 0, 1}
var LIGHT_PURPLE = Color{1, .85, 1}

var PlayerColors = map[int]Color{
	-1: GRAY,
	0:  BLUE,
	1:  GREEN,
	2:  RED,
	3:  ORANGE,
}

type SystemDisplay int

const (
	NAV_DISPLAY SystemDisplay = iota
	ORDER_DISPLAY
	WEAPON_DISPLAY
	COMET_DISPLAY
	KITE_DISPLAY
	MAP_DISPLAY
	PREDICTION_DISPLAY
	COMBAT_DISPLAY
)

var SystemsEnabled = map[SystemDisplay]bool{
	NAV_DISPLAY:        true,
	COMET_DISPLAY:      true,
	KITE_DISPLAY:       false,
	MAP_DISPLAY:        true,
	ORDER_DISPLAY:      true,
	WEAPON_DISPLAY:     false,
	PREDICTION_DISPLAY: false,
	COMBAT_DISPLAY:     false,
}

func (game Game) IsDisplayingSystem(sys SystemDisplay) bool {
	return game.display != nil && DISPLAY_ENABLED && SystemsEnabled[sys]
}

func (game Game) IsDisplaying() bool {
	return game.display != nil && DISPLAY_ENABLED
}

func (game Game) ShowMap() {
	if !game.IsDisplaying() {
		game.LogOnce("Display is nil, not saving images")
		return
	}
	game.LogOnce("Display exists, saving images")

	d := game.display

	for _, planet := range game.AllPlanets() {
		//game.Log("Drawing circle for planet %v", planet)
		//		safetyColor := RED
		//		if safe, _ := IsPlanetSafe(game, planet, game.pid, game.Width()/MAX_SPEED); safe {
		//			safetyColor = GREEN
		//		}
		//		game.DrawEntity(planet, safetyColor, 3, MAP_DISPLAY)
		game.DrawEntity(planet, PlayerColors[planet.Owner], 2, MAP_DISPLAY)
	}

	for _, ship := range game.AllShips() {
		//game.Log("Drawing circle for ship %v", ship)
		if !ship.IsUndocked() {
			planet, _ := game.GetPlanet(ship.DockedPlanet)
			game.DrawLine(ship, planet, BLACK, 2, MAP_DISPLAY)
		} else {
			c := PlayerColors[ship.Owner]
			if c.R != 1.0 {
				c.R = .9
			}
			if c.G != 1.0 {
				c.G = .9
			}
			if c.B != 1.0 {
				c.B = .9
			}
			game.DrawEntity(ship.AsRadius(WEAPON_RADIUS+SHIP_RADIUS), c, 1, WEAPON_DISPLAY)
		}
	}

	for _, ship := range game.AllShips() {
		h := float64(ship.HP)/512 + 128
		game.DrawEntity(ship, Color{h, h, h}, 0, MAP_DISPLAY)
		game.DrawEntity(ship, PlayerColors[ship.Owner], 2, MAP_DISPLAY)
		game.DrawString(strconv.Itoa(ship.Id), ship.GetX(), ship.GetY(), BLACK, 1, MAP_DISPLAY)
	}

	path := "thoughts/" + strconv.Itoa(game.pid) + "-" + strconv.Itoa(game.turn) + ".png"
	game.Log("Saving image to path: %v", path)
	d.SavePNG(path)
	d.SetRGB(1, 1, 1)
	d.Clear()
}

func (game Game) DrawEntity(e Entity, color Color, w float64, sys SystemDisplay) {
	if game.IsDisplayingSystem(sys) {
		d := game.display

		ratio := Ratio(game)
		d.DrawCircle(e.GetX()*ratio, e.GetY()*ratio, e.GetRadius()*ratio)
		game.SetContextDisplay(color, w)
	}
}

func (game Game) SetContextDisplay(color Color, w float64) {
	d := game.display
	d.SetRGBA(color.R, color.G, color.B, 1)
	if w > 0 {
		d.SetLineWidth(w)
		d.Stroke()
	} else {
		d.Fill()
	}
}

func (game Game) DrawLineString(points []Entity, color Color, w float64, sys SystemDisplay) {
	if game.IsDisplayingSystem(sys) {
		for i, _ := range points {
			if i+1 < len(points) {
				game.DrawLine(points[i], points[i+1], color, w, sys)
			}
		}
	}
}

func (game Game) DrawPolygon(points []Entity, color Color, w float64, sys SystemDisplay) {
	if game.IsDisplayingSystem(sys) {
		for i, _ := range points {
			if i+1 == len(points) {
				game.DrawLine(points[i], points[0], color, w, sys)
			} else {
				game.DrawLine(points[i], points[i+1], color, w, sys)
			}
		}
	}
}

func (game Game) DrawLine(start, end Entity, color Color, w float64, sys SystemDisplay) {
	if game.IsDisplayingSystem(sys) {
		d := game.display
		ratio := Ratio(game)
		d.DrawLine(start.GetX()*ratio, start.GetY()*ratio, end.GetX()*ratio, end.GetY()*ratio)
		game.SetContextDisplay(color, w)
	}
}

func (game Game) DrawString(s string, x, y float64, color Color, w float64, sys SystemDisplay) {
	if game.IsDisplayingSystem(sys) {
		d := game.display
		ratio := Ratio(game)
		d.DrawStringAnchored(s, x*ratio, y*ratio, .4, .4)
		game.SetContextDisplay(color, w)
	}
}

func (game Game) DrawArc(source Entity, radius, lowAngle, highAngle float64, color Color, w float64, sys SystemDisplay) {
	if game.IsDisplayingSystem(sys) {
		d := game.display
		ratio := Ratio(game)

		if math.Abs(lowAngle-highAngle) > 180 {
			if highAngle > lowAngle {
				highAngle -= 360
			} else {
				lowAngle -= 360
			}
		}
		d.DrawArc(source.GetX()*ratio, source.GetY()*ratio, radius*ratio, gg.Radians(lowAngle), gg.Radians(highAngle))

		game.SetContextDisplay(color, w)
	}
}

func Ratio(game Game) float64 {
	return float64(DISPLAY_WIDTH) / float64(game.width)
}

func CreateDisplayContext(game Game) *gg.Context {
	w := DISPLAY_WIDTH
	h := int(float64(game.height) * Ratio(game))
	ctx := gg.NewContext(w, h)
	ctx.SetRGB(1, 1, 1)
	ctx.Clear()
	return ctx
}
