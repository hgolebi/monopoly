package neatnetwork

import (
	"fmt"
	"math/rand/v2"
	"monopoly/pkg/config"
	"monopoly/pkg/monopoly"
	"slices"
	"sync"

	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

type SimplePlayerBot struct {
	score int
	mutex sync.Mutex
	wins  int
	draws int
}

func (bot *SimplePlayerBot) GetName() string {
	return "SimpleBot"
}
func (bot *SimplePlayerBot) GetId() int {
	return -1
}
func (bot *SimplePlayerBot) GetScore() int {
	bot.mutex.Lock()
	defer bot.mutex.Unlock()
	return bot.score
}
func (bot *SimplePlayerBot) GetOrganism() *genetics.Organism {
	return nil
}
func (bot *SimplePlayerBot) AddScore(points int) {
	bot.mutex.Lock()
	bot.score += points
	bot.mutex.Unlock()
}

func (bot *SimplePlayerBot) AddWin() {
	bot.mutex.Lock()
	bot.wins++
	bot.mutex.Unlock()
}

func (bot *SimplePlayerBot) AddDraw() {
	bot.mutex.Lock()
	bot.draws++
	bot.mutex.Unlock()
}

func (bot *SimplePlayerBot) GetWins() int {
	bot.mutex.Lock()
	defer bot.mutex.Unlock()
	return bot.wins
}

func (bot *SimplePlayerBot) GetDraws() int {
	bot.mutex.Lock()
	defer bot.mutex.Unlock()
	return bot.draws
}

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
		keyProperties := FindKeyProperties(state, player)
		for _, propertyId := range keyProperties {
			if slices.Contains(availableActions.BuyPropertyList, propertyId) {
				price := state.Properties[propertyId].Price / 2
				if playerCash-price >= 200 {
					retValue.Action = monopoly.BUYOFFER
					retValue.PropertyId = propertyId
					retValue.Price = price
					if retValue.Price > 1000 {
						fmt.Println("PRICE HIGH !!!!!!!!!!!!!!!!")
						fmt.Printf("PropertyID: %d, Name: %s\n", propertyId, state.Properties[propertyId].Name)
						fmt.Printf("Offer: %d\n", retValue.Price)
						fmt.Printf("Property price: %d\n", state.Properties[propertyId].Price)
						fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
						panic("High price offer")
					}
					return retValue
				} else {
					need_money = true
				}
			}
		}
	}

	unwantedProperties := findUnwantedProperties(state, player)
	if need_money && len(unwantedProperties) > 0 {
		randIdx := rand.IntN(len(unwantedProperties))
		propertyId := unwantedProperties[randIdx]

		// Selling properties
		if state.SellOfferTries < config.MAX_OFFER_TRIES && slices.Contains(availableActions.SellPropertyList, propertyId) {
			retValue.Action = monopoly.SELLOFFER
			retValue.PropertyId = propertyId
			property := state.Properties[propertyId]
			retValue.Price = int(float64(property.Price) * 1.5)
			retValue.Players = []int{0, 1, 2, 3}
			if retValue.Price > 1000 {
				fmt.Println("PRICE HIGH !!!!!!!!!!!!!!!!")
				fmt.Printf("PropertyID: %d, Name: %s\n", propertyId, state.Properties[propertyId].Name)
				fmt.Printf("Offer: %d\n", retValue.Price)
				fmt.Printf("Property price: %d\n", state.Properties[propertyId].Price)
				fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
				panic("High price offer")
			}
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
func (bot *SimplePlayerBot) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, sellerOffer int) int {
	// Called when seller has made a counteroffer during BUYOFFER negotiation.
	// Returns 0 to withdraw, or a price >= sellerOffer to accept/raise the offer.
	property := state.Properties[propertyId]
	playerMoney := state.Players[player].Money

	if playerMoney-sellerOffer < 200 {
		return 0 // Cannot afford it
	}
	if sellerOffer <= property.Price {
		return sellerOffer // Accept — price is at or below catalogue value
	}
	isKeyProperty := slices.Contains(FindKeyProperties(state, player), propertyId)
	if isKeyProperty && sellerOffer <= 2*property.Price {
		return sellerOffer // Accept — key property within budget
	}
	// Repeat buyer's current offer as impasse signal (no better counteroffer to make)
	return state.NegotiationBuyerOffer
}

func (bot *SimplePlayerBot) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, buyerOffer int) int {
	// Called when buyer has made an offer during BUYOFFER negotiation.
	// Returns 0 to hard-reject, a price <= buyerOffer to accept, or a price > buyerOffer as a counteroffer.
	property := state.Properties[propertyId]

	fullSetProperties := findPropertiesInFullSets(state, player)
	if slices.Contains(fullSetProperties, propertyId) {
		return 0 // Never sell monopoly properties
	}
	unwantedProperties := findUnwantedProperties(state, player)
	if slices.Contains(unwantedProperties, propertyId) && buyerOffer > property.Price {
		return buyerOffer // Accept — unwanted property above catalogue price
	}
	if buyerOffer > 2*property.Price {
		return buyerOffer // Accept — excellent price
	}
	// Counteroffer: ask for 1.5× catalogue price
	minAcceptable := property.Price * 3 / 2
	if buyerOffer >= minAcceptable {
		return buyerOffer // Accept if already at or above minimum acceptable
	}
	// Signal impasse if already at minimum (repeat last seller offer)
	if state.NegotiationSellerOffer == minAcceptable {
		return minAcceptable // Repeat = seller impasse signal
	}
	return minAcceptable
}

func (bot *SimplePlayerBot) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int {
	if state.Players[player].Money-currentPrice < 200 {
		return 0
	}
	isKeyProperty := slices.Contains(FindKeyProperties(state, player), propertyId)
	if isKeyProperty {
		return currentPrice + 20
	}
	if currentPrice < state.Properties[propertyId].Price {
		return currentPrice + 10
	}
	return 0
}

func (bot *SimplePlayerBot) ResetScore() {
	bot.mutex.Lock()
	bot.score = 0
	bot.mutex.Unlock()
}

func FindKeyProperties(state monopoly.GameState, playerId int) []int {
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
	for setIdx, properties := range have {
		// Skip small sets (2 properties) — Brown and DarkBlue are too valuable to mark as unwanted
		if len(properties) == 1 && len(monopoly.Sets[setIdx]) > 2 {
			unwanted = append(unwanted, properties[0])
		}
	}
	return unwanted
}

func findPropertiesInFullSets(state monopoly.GameState, playerId int) []int {
	have, missing := getSetMaps(state, playerId)
	fullSetProperties := []int{}
	for setIdx, properties := range missing {
		if len(properties) == 0 {
			fullSetProperties = append(fullSetProperties, have[setIdx]...)
		}
	}
	return fullSetProperties
}

func getSetMaps(state monopoly.GameState, playerId int) (have map[int][]int, missing map[int][]int) {
	have = make(map[int][]int)
	missing = make(map[int][]int)
	// Initialize all street sets (skip Railroad and Utility)
	for setIdx := range monopoly.Sets {
		if setIdx == monopoly.RailroadSetIndex || setIdx == monopoly.UtilitySetIndex {
			continue
		}
		have[setIdx] = []int{}
		missing[setIdx] = []int{}
	}
	for idx, property := range state.Properties {
		if property.SetIndex == monopoly.RailroadSetIndex || property.SetIndex == monopoly.UtilitySetIndex {
			continue
		}
		if property.Owner == nil || property.Owner.ID != playerId {
			missing[property.SetIndex] = append(missing[property.SetIndex], idx)
		} else {
			have[property.SetIndex] = append(have[property.SetIndex], idx)
		}
	}
	return
}
