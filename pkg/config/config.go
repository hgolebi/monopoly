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

	// game settings used for normalization of NEAT input/outputs
	LAST_FIELD_ID    = 39
	LAST_PROPERTY_ID = 27
	MAX_MONEY        = 2000
	MAX_JAIL_CARDS   = 10
	LAST_PLAYER_ID   = MAX_PLAYERS - 1
)

// Evaluator settings
const (
	// PUNISHMENT_FIRST_THRESHOLD  = 0 // if player bankrupts before this round he will receive HIGHEST_PUNISHMENT
	// PUNISHMENT_SECOND_THRESHOLD = 0 // if player bankrupts before this round he will receive SECOND_HIGHEST_PUNISHMENT
	// HIGHEST_PUNISHMENT          = 0
	// SECOND_HIGHEST_PUNISHMENT   = 0
	ROUND_SCORE         = 1 // points for each round survived for non-winner players
	FIRST_PLACE_SCORE   = 500
	POINTS_PER_PROPERTY = 3  // points for each property owned by the player.
	POINTS_PER_HOUSE    = 10 // points for each house owned by the player.
	ALIVE_BONUS         = 50 // bonus points for not going bankrupt
	// SECOND_PLACE_SCORE          = 0
	// ROUND_LIMIT_WINNER_SCORE    = 0 // if player wins the game by reaching the round limit he will receive this score

	TOURNAMENT_TYPE       = "single_round" // "single_round" or "bracket"
	INCLUDE_HEURISTIC_BOT = true           // whether to include a heuristic bot in the games played during evaluation

	GAMES_PER_EPOCH = 1500 // number of games every organism has to play during one epoch
	GROUP_SIZE      = 4    // number of players in each game
	MAX_THREADS     = 100  // maximum number of threads used to evaluate organisms
	PRINT_EVERY     = 50   // saves logs and population to files every N epochs
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
	}
}
