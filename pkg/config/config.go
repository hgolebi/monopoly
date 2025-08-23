package config

// game settings
const (
	MAX_ROUNDS       = 50
	START_PASS_MONEY = 200
	JAIL_POSITION    = 10
	JAIL_BAIL        = 50
	MAX_HOUSES       = 5
	MIN_PRICE        = 10
	MAX_OFFER_TRIES  = 3
	MAX_PLAYERS      = 4

	// game settings used for normalization of NEAT input/outputs
	LAST_FIELD_ID    = 39
	LAST_PROPERTY_ID = 27
	MAX_MONEY        = 15000
	MAX_JAIL_CARDS   = 20
	LAST_PLAYER_ID   = MAX_PLAYERS - 1
)

// Evaluator settings
const (
	PUNISHMENT_FIRST_THRESHOLD  = 10 // if player bankrupts before this round he will receive HIGHEST_PUNISHMENT
	PUNISHMENT_SECOND_THRESHOLD = 20 // if player bankrupts before this round he will receive SECOND_HIGHEST_PUNISHMENT
	HIGHEST_PUNISHMENT          = -60
	SECOND_HIGHEST_PUNISHMENT   = -30
	SECOND_PLACE_SCORE          = 20
	FIRST_PLACE_SCORE           = 100
	ROUND_LIMIT_WINNER_SCORE    = 5  // if player wins the game by reaching the round limit he will receive this score
	POINT_PER_PROPERTY          = 10 // points for each property owned by the player. Only for winner
	POINTS_PER_HOUSE            = 20 // points for each house on the property. Only for winner

	GAMES_PER_EPOCH = 4 // number of games every organism has to play during one epoch
	GROUP_SIZE      = 4 // number of players in each game
)

type GameSettings struct {
	MaxRounds      int
	StartPassMoney int
	JailPosition   int
	JailBail       int
	MaxHouses      int
	MinPrice       int
	MaxOfferTries  int
}

func NewGameSettings() GameSettings {
	return GameSettings{
		MaxRounds:      MAX_ROUNDS,
		StartPassMoney: START_PASS_MONEY,
		JailPosition:   JAIL_POSITION,
		JailBail:       JAIL_BAIL,
		MaxHouses:      MAX_HOUSES,
		MinPrice:       MIN_PRICE,
		MaxOfferTries:  MAX_OFFER_TRIES,
	}
}
