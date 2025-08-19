package neatnetwork

import (
	"math"
	cfg "monopoly/pkg/config"
	"monopoly/pkg/monopoly"
)

var propertyInputs = map[int]map[string]int{
	0: { // brown1
		"OWNER":        0,
		"IS_MORTGAGED": 1,
		"HOUSES":       2,
	},
	1: { // brown2
		"OWNER":        3,
		"IS_MORTGAGED": 4,
		"HOUSES":       5,
	},
	2: { // railway1
		"OWNER":        6,
		"IS_MORTGAGED": 7,
	},
	3: { // lightblue1
		"OWNER":        8,
		"IS_MORTGAGED": 9,
		"HOUSES":       10,
	},
	4: { // lightblue2
		"OWNER":        11,
		"IS_MORTGAGED": 12,
		"HOUSES":       13,
	},
	5: { // lightblue3
		"OWNER":        14,
		"IS_MORTGAGED": 15,
		"HOUSES":       16,
	},
	6: { // railway2
		"OWNER":        17,
		"IS_MORTGAGED": 18,
	},
	7: { // pink1
		"OWNER":        19,
		"IS_MORTGAGED": 20,
		"HOUSES":       21,
	},
	8: { // pink2
		"OWNER":        22,
		"IS_MORTGAGED": 23,
		"HOUSES":       24,
	},
	9: { // pink3
		"OWNER":        25,
		"IS_MORTGAGED": 26,
		"HOUSES":       27,
	},
	10: { // utility1
		"OWNER":        28,
		"IS_MORTGAGED": 29,
	},
	11: { // orange1
		"OWNER":        30,
		"IS_MORTGAGED": 31,
		"HOUSES":       32,
	},
	12: { // orange2
		"OWNER":        33,
		"IS_MORTGAGED": 34,
		"HOUSES":       35,
	},
	13: { // orange3
		"OWNER":        36,
		"IS_MORTGAGED": 37,
		"HOUSES":       38,
	},
	14: { // railway3
		"OWNER":        39,
		"IS_MORTGAGED": 40,
	},
	15: { // red1
		"OWNER":        41,
		"IS_MORTGAGED": 42,
		"HOUSES":       43,
	},
	16: { // red2
		"OWNER":        44,
		"IS_MORTGAGED": 45,
		"HOUSES":       46,
	},
	17: { // red3
		"OWNER":        47,
		"IS_MORTGAGED": 48,
		"HOUSES":       49,
	},
	18: { // yellow1
		"OWNER":        50,
		"IS_MORTGAGED": 51,
		"HOUSES":       52,
	},
	19: { // utility2
		"OWNER":        53,
		"IS_MORTGAGED": 54,
	},
	20: { // yellow2
		"OWNER":        55,
		"IS_MORTGAGED": 56,
		"HOUSES":       57,
	},
	21: { // yellow3
		"OWNER":        58,
		"IS_MORTGAGED": 59,
		"HOUSES":       60,
	},
	22: { // railway4
		"OWNER":        61,
		"IS_MORTGAGED": 62,
	},
	23: { // green1
		"OWNER":        63,
		"IS_MORTGAGED": 64,
		"HOUSES":       65,
	},
	24: { // green2
		"OWNER":        66,
		"IS_MORTGAGED": 67,
		"HOUSES":       68,
	},
	25: { // green3
		"OWNER":        69,
		"IS_MORTGAGED": 70,
		"HOUSES":       71,
	},
	26: { // blue1
		"OWNER":        72,
		"IS_MORTGAGED": 73,
		"HOUSES":       74,
	},
	27: { // blue2
		"OWNER":        75,
		"IS_MORTGAGED": 76,
		"HOUSES":       77,
	},
}

var playerInputs = map[int]map[string]int{
	0: {
		"IS_ALIVE":   78,
		"IS_JAILED":  79,
		"POSITION":   80,
		"MONEY":      81,
		"JAIL_CARDS": 82,
	},
	1: {
		"IS_ALIVE":   83,
		"IS_JAILED":  84,
		"POSITION":   85,
		"MONEY":      86,
		"JAIL_CARDS": 87,
	},
	2: {
		"IS_ALIVE":   88,
		"IS_JAILED":  89,
		"POSITION":   90,
		"MONEY":      91,
		"JAIL_CARDS": 92,
	},
	3: {
		"IS_ALIVE":   93,
		"IS_JAILED":  94,
		"POSITION":   95,
		"MONEY":      96,
		"JAIL_CARDS": 97,
	},
}

// Inputs dedicated to the player making the decision, information is redundant
var currPlayerInputs = map[string]int{
	"IS_ALIVE":   98,
	"IS_JAILED":  99,
	"POSITION":   100,
	"MONEY":      101,
	"JAIL_CARDS": 102,
}

var baseInputs = map[string]int{
	"PLAYER_ID":        103,
	"CURR_PLAYER":      104,
	"ROUND":            105, // Current round number
	"DECISION_CONTEXT": 106, // Current decision context, for example bidding decision, buying decision
	"PROPERTY_ID":      107, // In case of property-related decisions like bidding
	"PRICE":            108, // In case of price-related decisions; normalized to 0.0 - 1.0, where 1.0 is MAX_MONEY
	"CURR_BID":         109, // In case of bidding
	"CURR_BID_WINNER":  110, // In case of bidding
	"CHARGE":           111, // In case of charge that would result in player going bankrupt
	"SELL_OFFER_TRIES": 112, // In case of standard actions
	"BUY_OFFER_TRIES":  113, // In case of standard actions
}

var availableStdActionInputs = map[monopoly.StdAction]int{
	monopoly.NOACTION:  114,
	monopoly.MORTGAGE:  115,
	monopoly.BUYOUT:    116,
	monopoly.SELLOFFER: 117,
	monopoly.BUYOFFER:  118,
	monopoly.BUYHOUSE:  119,
	monopoly.SELLHOUSE: 120,
}

var availableJailActionInputs = map[monopoly.JailAction]int{
	monopoly.ROLL_DICE: 121,
	monopoly.BAIL:      122,
	monopoly.CARD:      123,
}

type DecisionContext int

const (
	BUY_DECISION DecisionContext = iota
	BIDDING_DECISION
	JAIL_DECISION
	BUY_FROM_PLAYER
	SELL_TO_PLAYER
	STD_ACTION
)

var outputs = map[string]int{
	"BUY_DECISION":    0, // yes / no
	"BID_DECISION":    1, // yes / no
	"BUY_FROM_PLAYER": 2, // yes / no
	"SELL_TO_PLAYER":  3, // yes / no

	// jail actions; highest score is the result (if applicable)
	"JAIL_ROLL_DICE": 4,
	"JAIL_BAIL":      5,
	"JAIL_CARD":      6,

	// standard actions; highest score is the result (if applicable)
	"NO_ACTION":  7,
	"MORTGAGE":   8,
	"BUYOUT":     9,
	"SELL_OFFER": 10,
	"BUY_OFFER":  11,
	"BUY_HOUSE":  12,
	"SELL_HOUSE": 13,

	// in case of player-related actions; highest score is the result (if applicable)
	"PLAYER_1": 14,
	"PLAYER_2": 15,
	"PLAYER_3": 16,
	"PLAYER_4": 17,

	"PRICE": 18, // in case of price-related actions; normalized to 0.0 - 1.0, where 1.0 is MAX_MONEY
}

func GetStdActionOutputValues(output []float64) map[monopoly.StdAction]float64 {
	return map[monopoly.StdAction]float64{
		monopoly.NOACTION:  output[outputs["NO_ACTION"]],
		monopoly.MORTGAGE:  output[outputs["MORTGAGE"]],
		monopoly.BUYOUT:    output[outputs["BUYOUT"]],
		monopoly.SELLOFFER: output[outputs["SELL_OFFER"]],
		monopoly.BUYOFFER:  output[outputs["BUY_OFFER"]],
		monopoly.BUYHOUSE:  output[outputs["BUY_HOUSE"]],
		monopoly.SELLHOUSE: output[outputs["SELL_HOUSE"]],
	}

}

func GetPlayerOutputValues(output []float64) map[int]float64 {
	return map[int]float64{
		0: output[outputs["PLAYER_1"]],
		1: output[outputs["PLAYER_2"]],
		2: output[outputs["PLAYER_3"]],
		3: output[outputs["PLAYER_4"]],
	}
}

func GetPriceOutputValue(output []float64) int {
	out := output[outputs["PRICE"]]
	return int(math.Round(out * float64(cfg.MAX_MONEY)))
}

func GetJailOutputValues(output []float64) map[monopoly.JailAction]float64 {
	return map[monopoly.JailAction]float64{
		monopoly.ROLL_DICE: output[outputs["JAIL_ROLL_DICE"]],
		monopoly.BAIL:      output[outputs["JAIL_BAIL"]],
		monopoly.CARD:      output[outputs["JAIL_CARD"]],
	}
}

type MonopolySensors []float64

func NewMonopolySensors() MonopolySensors {
	return make([]float64, 124)
}

func (s MonopolySensors) LoadState(state monopoly.GameState, playerID int) {
	for idx, player := range state.Players {
		s.loadPlayerState(idx, player)
		if idx == playerID {
			s.loadCurrentPlayerState(player)
		}
	}
	for idx, property := range state.Properties {
		s.loadPropertyState(idx, property)
	}
	s[baseInputs["PLAYER_ID"]] = normalize(playerID, 0, cfg.LAST_PLAYER_ID, true)
	s[baseInputs["CURR_PLAYER"]] = normalize(state.CurrentPlayerIdx, 0, cfg.LAST_PLAYER_ID, true)
	s[baseInputs["ROUND"]] = normalize(state.Round, 0, cfg.MAX_ROUNDS, false)
}

func (s MonopolySensors) loadPlayerState(playerId int, player *monopoly.Player) {
	s[playerInputs[playerId]["IS_ALIVE"]] = fromBool(!player.IsBankrupt)
	s[playerInputs[playerId]["IS_JAILED"]] = fromBool(player.IsJailed)
	s[playerInputs[playerId]["POSITION"]] = normalize(player.CurrentPosition, 0, cfg.LAST_FIELD_ID, false)
	s[playerInputs[playerId]["MONEY"]] = normalize(player.Money, 0, cfg.MAX_MONEY, false)
	s[playerInputs[playerId]["JAIL_CARDS"]] = normalize(player.JailCards, 0, cfg.MAX_JAIL_CARDS, false)
}

func (s MonopolySensors) loadPropertyState(propertyId int, property *monopoly.Property) {
	s[propertyInputs[propertyId]["IS_MORTGAGED"]] = fromBool(property.IsMortgaged)
	if property.Owner != nil {
		s[propertyInputs[propertyId]["OWNER"]] = normalize(property.Owner.ID, 0, cfg.LAST_PLAYER_ID, true)
	}
	if property.CanBuildHouse {
		s[propertyInputs[propertyId]["HOUSES"]] = normalize(property.Houses, 0, cfg.MAX_HOUSES, false)
	}
}

func (s MonopolySensors) loadCurrentPlayerState(player *monopoly.Player) {
	s[currPlayerInputs["IS_ALIVE"]] = fromBool(!player.IsBankrupt)
	s[currPlayerInputs["IS_JAILED"]] = fromBool(player.IsJailed)
	s[currPlayerInputs["POSITION"]] = normalize(player.CurrentPosition, 0, cfg.LAST_FIELD_ID, false)
	s[currPlayerInputs["MONEY"]] = normalize(player.Money, 0, cfg.MAX_MONEY, false)
	s[currPlayerInputs["JAIL_CARDS"]] = normalize(player.JailCards, 0, cfg.MAX_JAIL_CARDS, false)
}

func (s MonopolySensors) LoadDecisionContext(ctx DecisionContext) {
	s[baseInputs["DECISION_CONTEXT"]] = normalize(int(ctx), 0, int(STD_ACTION), false)
}

func (s MonopolySensors) LoadPropertyId(propertyId int) {
	s[baseInputs["PROPERTY_ID"]] = normalize(propertyId, 0, cfg.LAST_PROPERTY_ID, true)
}

func (s MonopolySensors) LoadPrice(price int) {
	s[baseInputs["PRICE"]] = normalize(price, 0, cfg.MAX_MONEY, false)
}

func (s MonopolySensors) LoadBiddingInputs(currentBid int, currentBidWinner int) {
	s[baseInputs["CURR_BID"]] = normalize(currentBid, 0, cfg.MAX_MONEY, false)
	s[baseInputs["CURR_BID_WINNER"]] = normalize(currentBidWinner, 0, cfg.LAST_PLAYER_ID, true)
}

func (s MonopolySensors) LoadCharge(charge int) {
	s[baseInputs["CHARGE"]] = normalize(charge, 0, cfg.MAX_MONEY, false)
}

func (s MonopolySensors) LoadSellOfferTries(tries int) {
	s[baseInputs["SELL_OFFER_TRIES"]] = normalize(tries, 0, cfg.MAX_OFFER_TRIES, false)
}

func (s MonopolySensors) LoadBuyOfferTries(tries int) {
	s[baseInputs["BUY_OFFER_TRIES"]] = normalize(tries, 0, cfg.MAX_OFFER_TRIES, false)
}

func (s MonopolySensors) LoadAvailableStdActions(actions []monopoly.StdAction) {
	for _, action := range actions {
		s[availableStdActionInputs[action]] = 1.0
	}
}

func (s MonopolySensors) LoadAvailableJailActions(actions []monopoly.JailAction) {
	for _, action := range actions {
		s[availableJailActionInputs[action]] = 1.0
	}
}

func fromBool(value bool) float64 {
	if value {
		return 1.0
	}
	return 0.0
}

func normalize(value int, min int, max int, shift bool) float64 {
	if max == min {
		return 0.0
	}
	dividend := float64(value - min)
	divisor := float64(max - min)
	if shift {
		dividend++
		divisor++
	}
	normalizedVal := dividend / divisor
	if normalizedVal < 0.0 {
		return 0.0
	}
	if normalizedVal > 1.0 {
		return 1.0
	}
	return normalizedVal
}
