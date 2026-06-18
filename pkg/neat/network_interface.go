package neatnetwork

import (
	"math"
	cfg "monopoly/pkg/config"
	"monopoly/pkg/monopoly"
)

type InputID int
type OutputID int

const (
	IN_PLAYER_IS_JAILED InputID = iota
	IN_PLAYER_POSITION
	IN_PLAYER_MONEY

	IN_PROPERTY_ID
	IN_PROPERTY_TYPE // street/railroad/utility
	IN_PROPERTY_PRICE

	IN_SET_ID
	IN_SET_PROPERTIES_NEEDED   // how many properties player needs to complete the set
	IN_SET_PROPERTIES_OCCUPIED // how many properties in the set are owned by other players

	// enemy is the player that is being targeted by the current action (buying from, selling to, bidding against)
	IN_ENEMY_INVOLVED
	IN_ENEMY_MONEY
	IN_ENEMY_SET_PROPERTIES_NEEDED   // how many properties in the set the enemy player needs to complete the set
	IN_ENEMY_SET_PROPERTIES_OCCUPIED // how many properties in the set are owned by players other than the enemy player

	IN_ENEMY_SELL_COUNTEROFFER // price offered by the enemy player for selling the property to the agent (if the current action is buying from the enemy player)
	IN_ENEMY_BUY_COUNTEROFFER  // price offered by the enemy player for buying the property from the agent (if the current action is selling to the enemy player)
	IN_NEGOTIATION_LAST_TRY    // if current offer is last try (either because max rounds of negotiation reached or because there is an impasse)

	IN_AVAILABLE_PROPERTIES // how many properties are still available for purchase (not owned by any player)
	IN_ROUND                // current round number

	IN_BUY_PRICE  // price for the current action of buying (including bidding)
	IN_SELL_PRICE // price for the current action of selling

	INPUT_COUNT
)

const (
	OUT_BUY_PROPERTY OutputID = iota
	OUT_SELL_PROPERTY
	OUT_MORTGAGE
	OUT_BUYOUT
	OUT_BUY_HOUSE

	OUTPUT_COUNT
)

type MonopolySensors []float64

func NewMonopolySensors() MonopolySensors {
	return make([]float64, int(INPUT_COUNT))
}

func (s MonopolySensors) LoadState(state monopoly.GameState, playerID int, propertyID int) {
	s.loadPlayerInputs(state, playerID)
	s.loadPropertyInputs(state, propertyID)
	s.loadSetInputs(state, propertyID, playerID)
	s.loadAvailablePropertiesInput(state)
	s[IN_ROUND] = normalize(state.Round, 0, cfg.MAX_ROUNDS, false)

}

func (s MonopolySensors) loadPlayerInputs(state monopoly.GameState, playerID int) {
	player := state.Players[playerID]
	s[IN_PLAYER_IS_JAILED] = fromBool(player.IsJailed)
	s[IN_PLAYER_POSITION] = normalize(player.CurrentPosition, 0, cfg.LAST_FIELD_ID, false)
	s[IN_PLAYER_MONEY] = normalize(player.Money, 0, cfg.MAX_MONEY, false)
}

func (s MonopolySensors) loadPropertyInputs(state monopoly.GameState, propertyID int) {
	property := state.Properties[propertyID]
	s[IN_PROPERTY_ID] = normalize(propertyID, 0, cfg.LAST_PROPERTY_ID, false)
	if property.SetIndex == monopoly.RailroadSetIndex {
		s[IN_PROPERTY_TYPE] = 0.0
	} else if property.SetIndex == monopoly.UtilitySetIndex {
		s[IN_PROPERTY_TYPE] = 0.5
	} else {
		s[IN_PROPERTY_TYPE] = 1.0
	}

	price := property.Price
	if property.IsMortgaged {
		price = price + property.GetBuyoutPrice()
	}
	s[IN_PROPERTY_PRICE] = normalize(price, 0, cfg.MAX_MONEY, false)
}

func (s MonopolySensors) loadSetInputs(state monopoly.GameState, propertyId int, playerID int) {
	property := state.Properties[propertyId]
	setId := property.SetIndex
	lastSetId := len(monopoly.Sets) - 1
	s[IN_SET_ID] = normalize(setId, 0, lastSetId, false)
	setCount, ownedByPlayer, ownedByOthers := getSetDetails(state, setId, playerID)
	s[IN_SET_PROPERTIES_NEEDED] = normalize(setCount-ownedByPlayer, 0, lastSetId, false)
	s[IN_SET_PROPERTIES_OCCUPIED] = normalize(ownedByOthers, 0, lastSetId, false)
}

func (s MonopolySensors) loadAvailablePropertiesInput(state monopoly.GameState) {
	availableProperties := 0
	for _, prop := range state.Properties {
		if prop.Owner == nil {
			availableProperties++
		}
	}
	s[IN_AVAILABLE_PROPERTIES] = normalize(availableProperties, 0, cfg.LAST_PROPERTY_ID+1, false)
}

func (s MonopolySensors) LoadEnemyInputs(state monopoly.GameState, enemyPlayerID int, propertyID int) {
	enemy := state.Players[enemyPlayerID]
	s[IN_ENEMY_INVOLVED] = fromBool(true)
	s[IN_ENEMY_MONEY] = normalize(enemy.Money, 0, cfg.MAX_MONEY, false)
	property := state.Properties[propertyID]
	setId := property.SetIndex
	lastSetId := len(monopoly.Sets) - 1
	setCount, ownedByEnemy, ownedByOthers := getSetDetails(state, setId, enemyPlayerID)
	s[IN_ENEMY_SET_PROPERTIES_NEEDED] = normalize(setCount-ownedByEnemy, 0, lastSetId, false)
	s[IN_ENEMY_SET_PROPERTIES_OCCUPIED] = normalize(ownedByOthers, 0, lastSetId, false)
}

func (s MonopolySensors) LoadBuyPriceInput(price int) {
	s[IN_BUY_PRICE] = normalize(price, 0, cfg.MAX_MONEY, false)
}

func (s MonopolySensors) LoadSellPriceInput(price int) {
	s[IN_SELL_PRICE] = normalize(price, 0, cfg.MAX_MONEY, false)
}

func (s MonopolySensors) LoadBuyCounterofferInputs(offer int, isLastTry bool) {
	s[IN_ENEMY_BUY_COUNTEROFFER] = normalize(offer, 0, cfg.MAX_MONEY, false)
	s[IN_NEGOTIATION_LAST_TRY] = fromBool(isLastTry)
}

func (s MonopolySensors) LoadSellCounterofferInputs(offer int, isLastTry bool) {
	s[IN_ENEMY_SELL_COUNTEROFFER] = normalize(offer, 0, cfg.MAX_MONEY, false)
	s[IN_NEGOTIATION_LAST_TRY] = fromBool(isLastTry)
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

func activation(x float64) float64 {
	// f(x) = x*(1-((1-x)/(0.8))^2)
	return x * (1 - math.Pow((1-x)/0.8, 2))
}
