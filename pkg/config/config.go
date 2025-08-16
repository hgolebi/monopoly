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
