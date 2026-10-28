package config

// game settings
const (
	MAX_ROUNDS       = 50
	START_PASS_MONEY = 200
	JAIL_POSITION    = 10
	JAIL_BAIL        = 50
	MAX_HOUSES       = 5
	MIN_PRICE        = 10
	MAX_OFFER_TRIES  = 1
	MAX_PLAYERS      = 4
	MAX_STD_ACTIONS  = 5
	MAX_TRADE_ROUNDS = 10
	PRICE_STEP       = 10 // granularity of prices searched by the bot and of its bids

	// game settings used for normalization of NEAT input/outputs
	LAST_FIELD_ID    = 39
	LAST_PROPERTY_ID = 27
	MAX_SET_SIZE     = 4
	MAX_MONEY        = 2000
	MAX_JAIL_CARDS   = 10
	LAST_PLAYER_ID   = MAX_PLAYERS - 1
)

// Evaluator settings
const (
	FIRST_PLACE_SCORE = 400 // points for winning a game (last player not bankrupt)
	DRAW_WINNER_SCORE = 200 // points for the player with the highest net worth when the game ends by the round limit

	GAMES_PER_EPOCH = 1000 // number of games every organism has to play during one epoch
	GROUP_SIZE      = 4    // number of players in each game
	MAX_THREADS     = 100  // maximum number of threads used to evaluate organisms
	PRINT_EVERY     = 20   // saves logs and population to files every N epochs

	// compat_threshold from the NEAT options is the starting value; after every epoch it moves towards TARGET_SPECIES
	// species: up by COMPAT_THRESHOLD_STEP_UP, down by the smaller COMPAT_THRESHOLD_STEP_DOWN, because new species
	// appear a few epochs after the threshold drops and a fast drop creates too many at once.
	// (goNEAT keeps the threshold fixed, and then the number of species grows with the genome distances.)
	// A higher threshold only stops new species: goNEAT never merges species, extra ones disappear when they die out
	// (dropoff_age), so the threshold is capped at COMPAT_THRESHOLD_MAX instead of growing while they do.
	TARGET_SPECIES             = 10
	COMPAT_THRESHOLD_STEP_UP   = 0.3
	COMPAT_THRESHOLD_STEP_DOWN = 0.1
	COMPAT_THRESHOLD_MIN       = 0.3
	COMPAT_THRESHOLD_MAX       = 15.0
)

type GameSettings struct {
	MaxRounds            int
	StartPassMoney       int
	JailPosition         int
	JailBail             int
	MaxHouses            int
	MinPrice             int
	MaxOfferTries        int
	MaxStdActionsPerTurn int
	MaxTradeRounds       int
}

func NewGameSettings() GameSettings {
	return GameSettings{
		MaxRounds:            MAX_ROUNDS,
		StartPassMoney:       START_PASS_MONEY,
		JailPosition:         JAIL_POSITION,
		JailBail:             JAIL_BAIL,
		MaxHouses:            MAX_HOUSES,
		MinPrice:             MIN_PRICE,
		MaxOfferTries:        MAX_OFFER_TRIES,
		MaxStdActionsPerTurn: MAX_STD_ACTIONS,
		MaxTradeRounds:       MAX_TRADE_ROUNDS,
	}
}
