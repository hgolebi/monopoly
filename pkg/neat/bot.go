package neatnetwork

import (
	"math/rand/v2"
	"monopoly/pkg/config"
	"monopoly/pkg/monopoly"
	"slices"

	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

type SimplePlayerBot struct {
}

func (bot *SimplePlayerBot) GetName() string {
	return "SimpleBot"
}
func (bot *SimplePlayerBot) GetId() int {
	return -1
}
func (bot *SimplePlayerBot) GetScore() int {
	return 0
}
func (bot *SimplePlayerBot) GetOrganism() *genetics.Organism {
	return nil
}
func (bot *SimplePlayerBot) AddScore(points int) {}

func (bot *SimplePlayerBot) GetStdAction(player int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails {
	retValue := monopoly.ActionDetails{}
	if state.StdActionsUsed >= config.MAX_STD_ACTIONS {
		retValue.Action = monopoly.NOACTION
		return retValue
	}
	playerCash := state.Players[player].Money
	need_money := playerCash < 300

	// Buying houses
	if len(availableActions.BuyHouseList) > 0 {
		randIdx := rand.IntN(len(availableActions.BuyHouseList))
		propertyId := availableActions.BuyHouseList[randIdx]
		property := state.Properties[propertyId]
		if playerCash-property.HousePrice >= 200 {
			retValue.Action = monopoly.BUYHOUSE
			retValue.PropertyId = propertyId
			return retValue
		} else {
			need_money = true
		}
	}

	// Unmortgaging properties in full sets
	fullSetProperties := findPropertiesInFullSets(state, player)
	for _, propertyId := range fullSetProperties {
		if slices.Contains(availableActions.BuyOutList, propertyId) {
			propertyBuyOut := int(float64(state.Properties[propertyId].Price) * 1.1)
			if playerCash-propertyBuyOut >= 200 {
				retValue.Action = monopoly.BUYOUT
				retValue.PropertyId = propertyId
				return retValue
			} else {
				need_money = true
			}
		}
	}

	// Buying key properties
	if state.BuyOfferTries < config.MAX_OFFER_TRIES {
		keyProperties := findKeyProperties(state, player)
		for _, propertyId := range keyProperties {
			if slices.Contains(availableActions.BuyPropertyList, propertyId) {
				price := state.Properties[propertyId].Price / 2
				if playerCash-price >= 200 {
					retValue.Action = monopoly.BUYOFFER
					retValue.PropertyId = propertyId
					retValue.Price = price
					return retValue
				} else {
					need_money = true
				}
			}
		}
	}

	if need_money {
		unwantedProperties := findUnwantedProperties(state, player)
		randIdx := rand.IntN(len(unwantedProperties))
		propertyId := unwantedProperties[randIdx]

		// Selling properties
		if state.SellOfferTries < config.MAX_OFFER_TRIES && slices.Contains(availableActions.SellPropertyList, propertyId) {
			retValue.Action = monopoly.SELLOFFER
			retValue.PropertyId = propertyId
			property := state.Properties[propertyId]
			retValue.Price = int(float64(property.Price) * 1.5)
			retValue.Players = []int{0, 1, 2, 3}
			return retValue
		}

		// Mortgaging properties
		if slices.Contains(availableActions.MortgageList, propertyId) {
			retValue.Action = monopoly.MORTGAGE
			retValue.PropertyId = propertyId
			return retValue
		}
	}

	// Unmortgaging rest of properties
	for _, propertyId := range availableActions.BuyOutList {
		buyOut := int(float64(state.Properties[propertyId].Price) * 1.1)
		if playerCash-buyOut >= 200 {
			retValue.Action = monopoly.BUYOUT
			retValue.PropertyId = propertyId
			return retValue
		}
	}

	// Trying to buy properties for free
	if state.BuyOfferTries < config.MAX_OFFER_TRIES && len(availableActions.BuyPropertyList) > 0 {
		randIdx := rand.IntN(len(availableActions.BuyPropertyList))
		propertyId := availableActions.BuyPropertyList[randIdx]
		retValue.Action = monopoly.BUYOFFER
		retValue.PropertyId = propertyId
		retValue.Price = 0
		return retValue
	}

	retValue.Action = monopoly.NOACTION
	return retValue
}

func (bot *SimplePlayerBot) GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction {
	if !slices.Contains(available, monopoly.ROLL_DICE) {
		if slices.Contains(available, monopoly.CARD) {
			return monopoly.CARD
		}
		return monopoly.BAIL
	}
	if state.Round > 20 {
		return monopoly.ROLL_DICE
	}
	if slices.Contains(available, monopoly.CARD) {
		return monopoly.CARD
	}
	return monopoly.BAIL
}
func (bot *SimplePlayerBot) BuyDecision(player int, state monopoly.GameState, propertyId int) bool {
	if state.Players[player].Money-state.Properties[propertyId].Price >= 200 {
		return true
	}
	return false
}
func (bot *SimplePlayerBot) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	if state.Players[player].Money-price < 200 {
		return false
	}
	if price < state.Properties[propertyId].Price {
		return true
	}
	isKeyProperty := slices.Contains(findKeyProperties(state, player), propertyId)
	if isKeyProperty && price <= 2*state.Properties[propertyId].Price {
		return true
	}
	return false
}

func (bot *SimplePlayerBot) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	fullSetProperties := findPropertiesInFullSets(state, player)
	if slices.Contains(fullSetProperties, propertyId) {
		return false
	}
	unwantedProperties := findUnwantedProperties(state, player)
	if slices.Contains(unwantedProperties, propertyId) && price > state.Properties[propertyId].Price {
		return true
	}
	if price > 2*state.Properties[propertyId].Price {
		return true
	}
	return false
}

func (bot *SimplePlayerBot) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int {
	if state.Players[player].Money-currentPrice < 200 {
		return 0
	}
	isKeyProperty := slices.Contains(findKeyProperties(state, player), propertyId)
	if isKeyProperty {
		return currentPrice + 10
	}
	if currentPrice < state.Properties[propertyId].Price {
		return currentPrice + 1
	}
	return 0
}

func findKeyProperties(state monopoly.GameState, playerId int) []int {
	_, missing := getSetMaps(state, playerId)
	keyProperties := []int{}
	for _, properties := range missing {
		if len(properties) == 1 {
			keyProperties = append(keyProperties, properties[0])
		}
	}
	return keyProperties
}

func findUnwantedProperties(state monopoly.GameState, playerId int) []int {
	have, _ := getSetMaps(state, playerId)
	unwanted := []int{}
	for set, properties := range have {
		if len(properties) == 1 && set != "DarkBlue" && set != "Brown" {
			unwanted = append(unwanted, properties[0])
		}
	}
	return unwanted
}

func findPropertiesInFullSets(state monopoly.GameState, playerId int) []int {
	have, missing := getSetMaps(state, playerId)
	fullSetProperties := []int{}
	for set, properties := range missing {
		if len(properties) == 0 {
			fullSetProperties = append(fullSetProperties, have[set]...)
		}
	}
	return fullSetProperties
}

func getSetMaps(state monopoly.GameState, playerId int) (have map[string][]int, missing map[string][]int) {
	have = map[string][]int{
		"Brown":     {},
		"LightBlue": {},
		"Pink":      {},
		"Orange":    {},
		"Red":       {},
		"Yellow":    {},
		"Green":     {},
		"DarkBlue":  {},
	}
	missing = map[string][]int{
		"Brown":     {},
		"LightBlue": {},
		"Pink":      {},
		"Orange":    {},
		"Red":       {},
		"Yellow":    {},
		"Green":     {},
		"DarkBlue":  {},
	}
	for idx, property := range state.Properties {
		if property.Set == monopoly.RAILROAD || property.Set == monopoly.UTILITY {
			continue
		}
		if property.Owner == nil || property.Owner.ID != playerId {
			missing[property.Set] = append(missing[property.Set], idx)
		} else {
			have[property.Set] = append(have[property.Set], idx)
		}
	}
	return
}
