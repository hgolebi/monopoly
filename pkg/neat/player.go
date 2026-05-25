package neatnetwork

import (
	"errors"
	"fmt"
	cfg "monopoly/pkg/config"
	"monopoly/pkg/monopoly"
	"slices"
	"sync"

	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
	"github.com/yaricom/goNEAT/v4/neat/network"
)

type MonopolyPlayer interface {
	GetStdAction(player int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails
	GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction
	BuyDecision(player int, state monopoly.GameState, propertyId int) bool
	BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, sellerOffer int) (bool, int)
	SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, buyerOffer int) (bool, int)
	BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int
	AddScore(points int)
	AddWin()
	AddDraw()
	GetWins() int
	GetDraws() int
	GetName() string
	GetId() int
	GetScore() int
	GetOrganism() *genetics.Organism
	ResetScore()
}

type NEATMonopolyPlayer struct {
	network   *network.Network
	organism  *genetics.Organism
	max_depth int
	score     int
	mutex     sync.Mutex
	wins      int
	draws     int
}

func NewNEATMonopolyPlayer(organism *genetics.Organism) (*NEATMonopolyPlayer, error) {
	network, err := organism.Phenotype()
	if err != nil {
		errorMsg := fmt.Sprintf("Error getting phenotype for organism %d: %v\n", organism.Genotype.Id, err)
		return nil, fmt.Errorf(errorMsg)
	}
	max_depth, err := network.MaxActivationDepthWithCap(0)
	if err != nil {
		return nil, err
	}
	if max_depth <= 0 {
		return nil, errors.New("Invalid network depth: " + fmt.Sprint(max_depth))
	}

	return &NEATMonopolyPlayer{
		network:   network,
		organism:  organism,
		max_depth: max_depth,
		score:     0,
	}, nil
}

func (p *NEATMonopolyPlayer) GetName() string {
	return fmt.Sprintf("Bot%d", p.organism.Genotype.Id)
}

func (p *NEATMonopolyPlayer) GetId() int {
	return p.organism.Genotype.Id
}

func (p *NEATMonopolyPlayer) GetScore() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.score
}

func (p *NEATMonopolyPlayer) GetOrganism() *genetics.Organism {
	return p.organism
}

func (p *NEATMonopolyPlayer) GetDecision(input []float64) []float64 {
	err := p.network.LoadSensors(input)
	if err != nil {
		panic("Error loading sensors: " + err.Error())
	}
	success, err := p.network.ForwardSteps(p.max_depth)
	if err != nil {
		neat.DebugLog(fmt.Sprintf("Error during forward steps for organism %d: %v", p.organism.Genotype.Id, err))
	}
	if !success {
		neat.DebugLog(fmt.Sprintf("Forward steps failed for organism %d", p.organism.Genotype.Id))
	}
	var output []float64
	for _, node := range p.network.Outputs {
		output = append(output, node.Activation)
	}
	return output
}

func (p *NEATMonopolyPlayer) GetStdAction(player int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails {
	decision, property, needMoney := p.getBuyHouseDecision(player, state, availableActions)
	if decision {
		return monopoly.ActionDetails{
			Action:     monopoly.BUYHOUSE,
			PropertyId: property,
		}
	} else if needMoney {
		success, action := p.tryToGetMoney(player, state, availableActions)
		if success {
			return action
		}
	}

	decision, property, needMoney = p.getBuyOutDecision(player, state, availableActions)
	if decision {
		return monopoly.ActionDetails{
			Action:     monopoly.BUYOUT,
			PropertyId: property,
		}
	} else if needMoney {
		success, action := p.tryToGetMoney(player, state, availableActions)
		if success {
			return action
		}
	}

	decision, property, price, needMoney := p.getBuyKeyPropertyDecision(player, state, availableActions)
	if decision {
		return monopoly.ActionDetails{
			Action:     monopoly.BUYOFFER,
			PropertyId: property,
			Price:      price,
		}
	} else if needMoney {
		success, action := p.tryToGetMoney(player, state, availableActions)
		if success {
			return action
		}
	}

	return monopoly.ActionDetails{}
}

func (p *NEATMonopolyPlayer) tryToGetMoney(player int, state monopoly.GameState, availableActions monopoly.FullActionList) (success bool, action monopoly.ActionDetails) {
	success, property := p.tryMortgage(player, state, availableActions)
	if success {
		return true, monopoly.ActionDetails{
			Action:     monopoly.MORTGAGE,
			PropertyId: property,
		}
	}

	success, property, price := p.trySellProperty(player, state, availableActions)
	if success {
		return true, monopoly.ActionDetails{
			Action:     monopoly.SELLOFFER,
			PropertyId: property,
			Price:      price,
		}
	}

	return false, monopoly.ActionDetails{}
}

func (p *NEATMonopolyPlayer) tryMortgage(player int, state monopoly.GameState, availableActions monopoly.FullActionList) (success bool, propertyId int) {
	if len(availableActions.MortgageList) == 0 {
		return false, 0
	}
	outputs := make(map[int]float64)
	for _, propertyId := range availableActions.MortgageList {
		setCount, ownedByPlayer, _ := getSetDetails(state, state.Properties[propertyId].SetIndex, player)
		if setCount == ownedByPlayer {
			continue
		}
		property := state.Properties[propertyId]

		sensors := NewMonopolySensors()
		sensors.LoadState(state, player, propertyId)
		sensors.LoadSellPriceInput(property.Price / 2)

		outputList := p.GetDecision(sensors)
		if outputList[OUT_MORTGAGE] > 0.5 {
			outputs[propertyId] = outputList[OUT_MORTGAGE]
		}
	}
	decision, propertyId, _ := getBestDecision(outputs)
	return decision, propertyId
}

func (p *NEATMonopolyPlayer) trySellProperty(player int, state monopoly.GameState, availableActions monopoly.FullActionList) (decision bool, propertyId int, price int) {
	return false, 0, 0 // TO DO
}

func (p *NEATMonopolyPlayer) getBuyHouseDecision(playerId int, state monopoly.GameState, availableActions monopoly.FullActionList) (decision bool, propertyId int, needMoney bool) {
	if len(availableActions.BuyHouseList) == 0 {
		return false, 0, false
	}
	outputs := make(map[int]float64)

	player := state.Players[playerId]
	for _, propertyId := range availableActions.BuyHouseList {
		property := state.Properties[propertyId]
		if property.HousePrice >= player.Money {
			needMoney = true
			continue
		}
		sensors := NewMonopolySensors()
		sensors.LoadState(state, playerId, propertyId)
		sensors.LoadBuyPriceInput(property.HousePrice)
		outputList := p.GetDecision(sensors)
		if outputList[OUT_BUY_HOUSE] > 0.5 {
			outputs[propertyId] = outputList[OUT_BUY_HOUSE]
		}
	}
	decision, propertyId, _ = getBestDecision(outputs)
	return decision, propertyId, needMoney

}

func (p *NEATMonopolyPlayer) getBuyOutDecision(playerId int, state monopoly.GameState, availableActions monopoly.FullActionList) (decision bool, propertyId int, needMoney bool) {
	if len(availableActions.BuyOutList) == 0 {
		return false, 0, false
	}
	outputs := make(map[int]float64)

	player := state.Players[playerId]
	for _, propertyId := range availableActions.BuyOutList {
		setCount, ownedByPlayer, _ := getSetDetails(state, state.Properties[propertyId].SetIndex, playerId)
		if setCount != ownedByPlayer {
			continue
		}
		property := state.Properties[propertyId]
		buyoutPrice := property.GetBuyoutPrice()
		if buyoutPrice >= player.Money {
			needMoney = true
			continue
		}
		sensors := NewMonopolySensors()
		sensors.LoadState(state, playerId, propertyId)
		sensors.LoadBuyPriceInput(buyoutPrice)
		outputList := p.GetDecision(sensors)
		if outputList[OUT_BUYOUT] > 0.5 {
			outputs[propertyId] = outputList[OUT_BUYOUT]
		}
	}
	decision, propertyId, _ = getBestDecision(outputs)
	return decision, propertyId, needMoney
}

func (p *NEATMonopolyPlayer) getBuyKeyPropertyDecision(playerId int, state monopoly.GameState, availableActions monopoly.FullActionList) (decision bool, propertyId int, price int, needMoney bool) {
	keyProperties := findKeyPropertiesToBuy(state, playerId)
	outputs := make(map[int]float64)
	for propertyId := range keyProperties {
		property := state.Properties[propertyId]
		minPrice := property.Price / 2
		if minPrice >= state.Players[playerId].Money {
			needMoney = true
			continue
		}
		sensors := NewMonopolySensors()
		sensors.LoadState(state, playerId, propertyId)
		sensors.LoadBuyPriceInput(minPrice)
		sensors.LoadEnemyInputs(state, property.Owner.ID, propertyId)

		outputList := p.GetDecision(sensors)
		if outputList[OUT_BUY_PROPERTY] > 0.5 {
			outputs[propertyId] = outputList[OUT_BUY_PROPERTY]
		}
	}
	decision, propertyId, _ = getBestDecision(outputs)
	property := state.Properties[propertyId]
	price = p.findMaxBuyPrice(state, playerId, propertyId, property.Price/2)
	return decision, propertyId, price, needMoney

}

func (p *NEATMonopolyPlayer) findMaxBuyPrice(state monopoly.GameState, playerId int, propertyId int, minPrice int) int {
	price := minPrice
	decision := true
	for decision {
		price += 10
		if price >= cfg.MAX_MONEY || price >= state.Players[playerId].Money {
			return price - 10
		}
		decision = p.GetBuyFromPlayerDecision(playerId, state, propertyId, price)
	}
	return price - 10
}

func (p *NEATMonopolyPlayer) GetBuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player, propertyId)
	sensors.LoadBuyPriceInput(price)
	sensors.LoadEnemyInputs(state, state.Properties[propertyId].Owner.ID, propertyId)
	outputList := p.GetDecision(sensors)
	return outputList[OUT_BUY_PROPERTY] > 0.5
}

func (p *NEATMonopolyPlayer) GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction {
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

func (p *NEATMonopolyPlayer) BuyDecision(player int, state monopoly.GameState, propertyId int) bool {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player, propertyId)
	price := state.Properties[propertyId].Price
	sensors.LoadBuyPriceInput(price)
	outputList := p.GetDecision(sensors)
	return outputList[OUT_BUY_PROPERTY] > 0.5
}

func (p *NEATMonopolyPlayer) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player, propertyId)
	if currentWinner >= 0 {
		sensors.LoadEnemyInputs(state, currentWinner, propertyId)
	}
	bid := currentPrice + 10
	sensors.LoadBuyPriceInput(bid)

	outputList := p.GetDecision(sensors)
	if outputList[OUT_BUY_PROPERTY] <= 0.5 {
		return 0
	}
	return bid
}

func (p *NEATMonopolyPlayer) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, sellerOffer int) (bool, int) {
	// Called during BUYOFFER negotiation: seller has made a counteroffer.
	// Returns (false, _) to withdraw immediately.
	// Returns (true, price): price >= sellerOffer = accept, price <= buyerPrice = impasse, in between = raise offer.
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player, propertyId)
	sensors.LoadBuyPriceInput(sellerOffer)
	property := state.Properties[propertyId]
	sensors.LoadEnemyInputs(state, property.Owner.ID, propertyId)

	isLastTry := state.NegotiationRound >= cfg.MAX_TRADE_ROUNDS || state.SellerImpasse
	sensors.LoadBuyCounterofferInputs(sellerOffer, isLastTry)

	outputList := p.GetDecision(sensors)
	if outputList[OUT_BUY_PROPERTY] <= 0.5 {
		return false, 0 // Withdraw
	}
	return true, sellerOffer // Accept seller's offer
}

func (p *NEATMonopolyPlayer) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, buyerOffer int) (bool, int) {
	// Called during BUYOFFER negotiation: buyer has made an offer.
	// Returns (false, _) to hard-reject immediately.
	// Returns (true, price): price <= buyerOffer = accept, price >= lastSellerPrice = impasse, in between = lower counteroffer.
	isLastTry := state.NegotiationRound >= cfg.MAX_TRADE_ROUNDS || state.BuyerImpasse
	sell, price := p.findMinSellPrice(state, player, propertyId, buyerOffer, buyerOffer, isLastTry)

	return sell, price
}

func (p *NEATMonopolyPlayer) findMinSellPrice(state monopoly.GameState, playerId int, propertyId int, minPrice int, counteroffer int, isLastTry bool) (bool, int) {
	price := minPrice
	decision := p.getSellDecision(playerId, state, propertyId, price, counteroffer, isLastTry)
	for !decision {
		price += 10
		if price >= cfg.MAX_MONEY {
			return false, 0
		}
		decision = p.getSellDecision(playerId, state, propertyId, price, counteroffer, isLastTry)
	}
	return true, price
}

func (p *NEATMonopolyPlayer) getSellDecision(playerId int, state monopoly.GameState, propertyId int, price int, counteroffer int, isLastTry bool) bool {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, playerId, propertyId)
	sensors.LoadSellPriceInput(price)
	property := state.Properties[propertyId]
	sensors.LoadEnemyInputs(state, property.Owner.ID, propertyId)

	if counteroffer > 0 {
		sensors.LoadSellCounterofferInputs(counteroffer, isLastTry)
	}
	outputList := p.GetDecision(sensors)
	return outputList[OUT_SELL_PROPERTY] > 0.5
}

func (p *NEATMonopolyPlayer) AddScore(points int) {
	p.mutex.Lock()
	p.score += points
	p.mutex.Unlock()
}

func (p *NEATMonopolyPlayer) AddWin() {
	p.mutex.Lock()
	p.wins++
	p.mutex.Unlock()
}

func (p *NEATMonopolyPlayer) AddDraw() {
	p.mutex.Lock()
	p.draws++
	p.mutex.Unlock()
}

func (p *NEATMonopolyPlayer) GetWins() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.wins
}
func (p *NEATMonopolyPlayer) GetDraws() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.draws
}

func (p *NEATMonopolyPlayer) ResetScore() {
	p.mutex.Lock()
	p.score = 0
	p.mutex.Unlock()
}

func getBestDecision(decisions map[int]float64) (bestFound bool, bestId int, bestValue float64) {
	if len(decisions) == 0 {
		return false, 0, 0.0
	}

	bestId = -1
	bestValue = -1.0
	for id, value := range decisions {
		if value > bestValue {
			bestValue = value
			bestId = id
		}
	}
	return true, bestId, bestValue
}

func getSetDetails(state monopoly.GameState, setId int, playerId int) (propertiesInSet int, ownedByPlayer int, ownedByOthers int) {
	propertiesInSet = len(monopoly.Sets[setId])
	for _, propertyId := range monopoly.Sets[setId] {
		property := state.Properties[propertyId]
		if property.Owner == state.Players[playerId] {
			ownedByPlayer++
		} else if property.Owner != nil {
			ownedByOthers++
		}
	}
	return propertiesInSet, ownedByPlayer, ownedByOthers
}

func findKeyPropertiesToBuy(state monopoly.GameState, playerId int) map[int]bool {
	keyProperties := make(map[int]bool)
	player := state.Players[playerId]
	for _, propertyId := range player.Properties {
		setCount, ownedByPlayer, ownedByOthers := getSetDetails(state, state.Properties[propertyId].SetIndex, playerId)
		if ownedByPlayer == setCount-1 && ownedByOthers == 1 {
			keyProperties[propertyId] = true
		}
	}
	return keyProperties
}
