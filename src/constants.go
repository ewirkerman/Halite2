package src

// Fohristiwhirl, November 2017.
// A sort of alternative starter kit.

const (
	// Game constants or derivatives thereof
	SHIP_RADIUS         = 0.5
	DOCKING_RADIUS      = 4.0
	WEAPON_RADIUS       = 5.0
	WEAPON_DAMAGE       = 64
	DOCKING_TURNS       = 5
	MAX_SPEED           = 7
	DAMAGABLE_RANGE     = WEAPON_RADIUS + 2*SHIP_RADIUS + MAX_SPEED
	MAX_PLANET_RADIUS   = 16
	PRODUCTIVITY        = 6
	PRODUCTION_PER_SHIP = 72
	SPAWN_RADIUS        = 2

	// bot functionality constants or flags
	KITE                    = false
	PLANET_DESTRUCTION      = false
	RUSH                    = true
	ALLY_BEE_DANCE          = true
	STACKTRACE              = false
	BEE_BANDWIDTH           = 9
	BUFFER_TOLERANCE        = .1
	POINT_MATCH_TOLERANCE   = .0000001
	MIN_ATTACK_SHIPS        = 2
	MAX_CHASERS             = 0
	HEAVY_COMBAT_PROPORTION = .40
	MURDER_PROPORTION       = .6
	SAFE_HORIZON            = 20
)

type DockedStatus int

const (
	UNDOCKED DockedStatus = iota
	DOCKING
	DOCKED
	UNDOCKING
)

type EntityType int

const (
	UNSET EntityType = iota
	SHIP
	PLANET
	POINT
	NOTHING
)

type OrderType int

const (
	NO_ORDER OrderType = iota
	THRUST
	DOCK
	UNDOCK
)
