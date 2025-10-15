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
	1: {
		"IS_ALIVE": 78,
		"MONEY":    79,
	},
	2: {
		"IS_ALIVE": 80,
		"MONEY":    81,
	},
	3: {
		"IS_ALIVE": 82,
		"MONEY":    83,
	},
}

// Inputs dedicated to the player making the decision, information is redundant
var currPlayerInputs = map[string]int{
	"IS_ALIVE":   84,
	"IS_JAILED":  85,
	"POSITION":   86,
	"MONEY":      87,
	"JAIL_CARDS": 88,
}

var baseInputs = map[string]int{
	"DECISION_CONTEXT": 89, // Current decision context, for example bidding decision, buying decision
	"PROPERTY_ID":      90, // In case of property-related decisions like bidding
	"PRICE":            91, // In case of price-related decisions; normalized to 0.0 - 1.0, where 1.0 is MAX_MONEY
	"CURR_BID":         92, // In case of bidding
	"CURR_BID_WINNER":  93, // In case of bidding
	"CHARGE":           94, // In case of charge that would result in player going bankrupt
}

var availableStdActionInputs = map[monopoly.StdAction]int{
	monopoly.NOACTION:  95,
	monopoly.MORTGAGE:  96,
	monopoly.BUYOUT:    97,
	monopoly.SELLOFFER: 98,
	monopoly.BUYOFFER:  99,
	monopoly.BUYHOUSE:  100,
	monopoly.SELLHOUSE: 101,
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

	// standard actions; highest score is the result (if applicable)
	"NO_ACTION":  4,
	"MORTGAGE":   5,
	"BUYOUT":     6,
	"SELL_OFFER": 7,
	"BUY_OFFER":  8,
	"BUY_HOUSE":  9,
	"SELL_HOUSE": 10,

	// in case of sell offer; if player is included in the offer;
	"PLAYER_1": 11, // yes / no
	"PLAYER_2": 12, // yes / no
	"PLAYER_3": 13, // yes / no

	"PRICE": 14, // in case of price-related actions; normalized to 0.0 - 1.0, where 1.0 is MAX_MONEY
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
		1: output[outputs["PLAYER_1"]],
		2: output[outputs["PLAYER_2"]],
		3: output[outputs["PLAYER_3"]],
	}
}

func GetPriceOutputValue(output []float64) int {
	out := output[outputs["PRICE"]]
	return int(math.Round(out * float64(cfg.MAX_MONEY)))
}

type MonopolySensors []float64

func NewMonopolySensors() MonopolySensors {
	return make([]float64, 102)
}

func (s MonopolySensors) LoadState(state monopoly.GameState, playerID int) {
	id := 0
	for index, player := range state.Players {
		if index == playerID {
			s.loadCurrentPlayerState(player)
		} else {
			s.loadPlayerState(id, player)
			id++
		}
	}
	for idx, property := range state.Properties {
		s.loadPropertyState(idx, property, playerID)
	}
}

func (s MonopolySensors) loadPlayerState(id int, player *monopoly.Player) {
	// id is not the same as in game
	s[playerInputs[id]["IS_ALIVE"]] = fromBool(!player.IsBankrupt)
	s[playerInputs[id]["MONEY"]] = normalize(player.Money, 0, cfg.MAX_MONEY, false)
}

func (s MonopolySensors) loadPropertyState(propertyId int, property *monopoly.Property, currPlayerId int) {
	s[propertyInputs[propertyId]["IS_MORTGAGED"]] = fromBool(property.IsMortgaged)
	if property.Owner != nil {
		s[propertyInputs[propertyId]["OWNER"]] = normalize(getNewPlayerId(property.Owner.ID, currPlayerId), 0, cfg.LAST_PLAYER_ID, true)
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

func (s MonopolySensors) LoadBiddingInputs(currentBid int, currentBidWinner int, currPlayerId int) {
	s[baseInputs["CURR_BID"]] = normalize(currentBid, 0, cfg.MAX_MONEY, false)
	s[baseInputs["CURR_BID_WINNER"]] = normalize(getNewPlayerId(currentBidWinner, currPlayerId), 0, cfg.LAST_PLAYER_ID, true)
}

func (s MonopolySensors) LoadCharge(charge int) {
	s[baseInputs["CHARGE"]] = normalize(charge, 0, cfg.MAX_MONEY, false)
}

func (s MonopolySensors) LoadAvailableStdActions(actions []monopoly.StdAction) {
	for _, action := range actions {
		s[availableStdActionInputs[action]] = 1.0
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

func getNewPlayerId(original int, currPlayerId int) int {
	if original == currPlayerId {
		return 0
	}
	if original < currPlayerId {
		return original + 1
	}
	return original
}

func getOriginalPlayerId(newId int, currPlayerId int) int {
	if newId == 0 {
		return currPlayerId
	}
	if newId <= currPlayerId {
		return newId - 1
	}
	return newId
}
