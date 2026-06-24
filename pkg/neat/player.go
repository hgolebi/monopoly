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
	BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, sellerOffer int, tradingPartnerId int) (bool, int)
	SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, buyerOffer int, tradingPartnerId int) (bool, int)
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
	network    *network.Network
	organism   *genetics.Organism
	max_depth  int
	score      int
	mutex      sync.Mutex
	wins       int
	draws      int
	is_trading bool
	log        *BotDecisionLogger
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

func (p *NEATMonopolyPlayer) SetDecisionLogger(l *BotDecisionLogger) {
	p.log = l
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
	p.is_trading = false
	if p.log != nil {
		p.log.Warn("▶ GetStdAction", "player", player, "round", state.Round, "money", state.Players[player].Money)
	}

	decision, property, needMoney := p.getBuyHouseDecision(player, state, availableActions)
	if decision {
		if p.log != nil {
			p.log.Info("→ result", "action", "BUYHOUSE", "propertyId", property)
		}
		return monopoly.ActionDetails{
			Action:     monopoly.BUYHOUSE,
			PropertyId: property,
		}
	} else if needMoney {
		success, action := p.tryToGetMoney(player, state, availableActions)
		if success {
			if p.log != nil {
				p.log.Info("→ result (getMoney for BUYHOUSE)", "action", action.Action, "propertyId", action.PropertyId)
			}
			return action
		}
	}

	decision, property, needMoney = p.getBuyOutDecision(player, state, availableActions)
	if decision {
		if p.log != nil {
			p.log.Info("→ result", "action", "BUYOUT", "propertyId", property)
		}
		return monopoly.ActionDetails{
			Action:     monopoly.BUYOUT,
			PropertyId: property,
		}
	} else if needMoney {
		success, action := p.tryToGetMoney(player, state, availableActions)
		if success {
			if p.log != nil {
				p.log.Info("→ result (getMoney for BUYOUT)", "action", action.Action, "propertyId", action.PropertyId)
			}
			return action
		}
	}

	decision, property, price, needMoney := p.getBuyKeyPropertyDecision(player, state, availableActions)
	if decision {
		if p.log != nil {
			p.log.Info("→ result", "action", "BUYOFFER", "propertyId", property, "price", price)
		}
		return monopoly.ActionDetails{
			Action:     monopoly.BUYOFFER,
			PropertyId: property,
			Price:      price,
		}
	} else if needMoney {
		success, action := p.tryToGetMoney(player, state, availableActions)
		if success {
			if p.log != nil {
				p.log.Info("→ result (getMoney for BUYOFFER)", "action", action.Action, "propertyId", action.PropertyId)
			}
			return action
		}
	}

	if p.log != nil {
		p.log.Info("→ result", "action", "NONE")
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
	if p.log != nil {
		p.log.Info("▶ tryMortgage", "candidates", availableActions.MortgageList)
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
		if p.log != nil {
			p.log.Info("  mortgage candidate", "propertyId", propertyId, "name", property.Name, "mortgageValue", property.Price/2)
			p.log.Info("  network inputs", "values", formatFloatSlice(sensors))
			p.log.Info("  network outputs", "values", formatFloatSlice(outputList))
			p.log.Info("  OUT_MORTGAGE", "value", fmt.Sprintf("%.4f", outputList[OUT_MORTGAGE]), "decision", boolToYesNo(outputList[OUT_MORTGAGE] > 0.5))
		}
		if outputList[OUT_MORTGAGE] > 0.5 {
			outputs[propertyId] = outputList[OUT_MORTGAGE]
		}
	}
	decision, propertyId, _ := getBestDecision(outputs)
	if p.log != nil {
		p.log.Info("→ result", "decision", boolToYesNo(decision), "propertyId", propertyId)
	}
	return decision, propertyId
}

func (p *NEATMonopolyPlayer) trySellProperty(player int, state monopoly.GameState, availableActions monopoly.FullActionList) (decision bool, propertyId int, price int) {
	return false, 0, 0 // TO DO
}

func (p *NEATMonopolyPlayer) getBuyHouseDecision(playerId int, state monopoly.GameState, availableActions monopoly.FullActionList) (decision bool, propertyId int, needMoney bool) {
	if len(availableActions.BuyHouseList) == 0 {
		return false, 0, false
	}
	if p.log != nil {
		p.log.Info("▶ getBuyHouseDecision", "candidates", availableActions.BuyHouseList)
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
		if p.log != nil {
			p.log.Info("  house candidate", "propertyId", propertyId, "name", property.Name, "housePrice", property.HousePrice)
			p.log.Info("  network inputs", "values", formatFloatSlice(sensors))
			p.log.Info("  network outputs", "values", formatFloatSlice(outputList))
			p.log.Info("  OUT_BUY_HOUSE", "value", fmt.Sprintf("%.4f", outputList[OUT_BUY_HOUSE]), "decision", boolToYesNo(outputList[OUT_BUY_HOUSE] > 0.5))
		}
		if outputList[OUT_BUY_HOUSE] > 0.5 {
			outputs[propertyId] = outputList[OUT_BUY_HOUSE]
		}
	}
	decision, propertyId, _ = getBestDecision(outputs)
	if p.log != nil {
		p.log.Info("→ result", "decision", boolToYesNo(decision), "propertyId", propertyId, "needMoney", needMoney)
	}
	return decision, propertyId, needMoney

}

func (p *NEATMonopolyPlayer) getBuyOutDecision(playerId int, state monopoly.GameState, availableActions monopoly.FullActionList) (decision bool, propertyId int, needMoney bool) {
	if len(availableActions.BuyOutList) == 0 {
		return false, 0, false
	}
	if p.log != nil {
		p.log.Info("▶ getBuyOutDecision", "candidates", availableActions.BuyOutList)
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
		if p.log != nil {
			p.log.Info("  buyout candidate", "propertyId", propertyId, "name", property.Name, "buyoutPrice", buyoutPrice)
			p.log.Info("  network inputs", "values", formatFloatSlice(sensors))
			p.log.Info("  network outputs", "values", formatFloatSlice(outputList))
			p.log.Info("  OUT_BUYOUT", "value", fmt.Sprintf("%.4f", outputList[OUT_BUYOUT]), "decision", boolToYesNo(outputList[OUT_BUYOUT] > 0.5))
		}
		if outputList[OUT_BUYOUT] > 0.5 {
			outputs[propertyId] = outputList[OUT_BUYOUT]
		}
	}
	decision, propertyId, _ = getBestDecision(outputs)
	if p.log != nil {
		p.log.Info("→ result", "decision", boolToYesNo(decision), "propertyId", propertyId, "needMoney", needMoney)
	}
	return decision, propertyId, needMoney
}

func (p *NEATMonopolyPlayer) getBuyKeyPropertyDecision(playerId int, state monopoly.GameState, availableActions monopoly.FullActionList) (decision bool, propertyId int, price int, needMoney bool) {
	keyProperties := findKeyPropertiesToBuy(state, playerId, availableActions.BuyPropertyList)
	if p.log != nil {
		keys := make([]int, 0, len(keyProperties))
		for k := range keyProperties {
			keys = append(keys, k)
		}
		p.log.Info("▶ getBuyKeyPropertyDecision", "keyProperties", keys)
	}
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
		if p.log != nil {
			p.log.Info("  key property candidate", "propertyId", propertyId, "name", property.Name, "minPrice", minPrice)
			p.log.Info("  network inputs", "values", formatFloatSlice(sensors))
			p.log.Info("  network outputs", "values", formatFloatSlice(outputList))
			p.log.Info("  OUT_BUY_PROPERTY", "value", fmt.Sprintf("%.4f", outputList[OUT_BUY_PROPERTY]), "decision", boolToYesNo(outputList[OUT_BUY_PROPERTY] > 0.5))
		}
		if outputList[OUT_BUY_PROPERTY] > 0.5 {
			outputs[propertyId] = outputList[OUT_BUY_PROPERTY]
		}
	}
	decision, propertyId, _ = getBestDecision(outputs)
	property := state.Properties[propertyId]
	if !decision {
		if p.log != nil {
			p.log.Info("→ result", "decision", "NO", "needMoney", needMoney)
		}
		return decision, propertyId, 0, needMoney
	}
	price = p.findMaxBuyPrice(state, playerId, propertyId, property.Price/2, -1, false)
	if p.log != nil {
		p.log.Info("→ result", "decision", "YES", "propertyId", propertyId, "price", price)
	}
	return decision, propertyId, price, needMoney

}

func (p *NEATMonopolyPlayer) findMaxBuyPrice(state monopoly.GameState, playerId int, propertyId int, minPrice int, sellerOffer int, isLastTry bool) int {
	price := minPrice
	decision := p.getBuyFromPlayerDecision(playerId, state, propertyId, price, sellerOffer, isLastTry)
	maxPrice := cfg.MAX_MONEY
	if sellerOffer > 0 {
		maxPrice = sellerOffer
	}
	for decision {
		price += 10
		if price >= maxPrice {
			return maxPrice
		}
		if price >= state.Players[playerId].Money {
			return state.Players[playerId].Money
		}
		decision = p.getBuyFromPlayerDecision(playerId, state, propertyId, price, sellerOffer, isLastTry)
	}
	return price - 10
}

func (p *NEATMonopolyPlayer) getBuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, price int, sellerOffer int, isLastTry bool) bool {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player, propertyId)
	sensors.LoadBuyPriceInput(price)
	sensors.LoadEnemyInputs(state, state.Properties[propertyId].Owner.ID, propertyId)
	if sellerOffer > 0 {
		sensors.LoadBuyCounterofferInputs(sellerOffer, isLastTry)
	}
	outputList := p.GetDecision(sensors)
	if p.log != nil {
		p.log.Info("  getBuyFromPlayerDecision", "propertyId", propertyId, "price", price, "sellerOffer", sellerOffer, "isLastTry", isLastTry)
		p.log.Info("  network inputs", "values", formatFloatSlice(sensors))
		p.log.Info("  network outputs", "values", formatFloatSlice(outputList))
		p.log.Info("  OUT_BUY_PROPERTY", "value", fmt.Sprintf("%.4f", outputList[OUT_BUY_PROPERTY]), "decision", boolToYesNo(outputList[OUT_BUY_PROPERTY] > 0.5))
	}
	return outputList[OUT_BUY_PROPERTY] > 0.5
}

func (p *NEATMonopolyPlayer) GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction {
	if p.log != nil {
		p.log.Warn("▶ GetJailAction", "player", player, "round", state.Round, "available", available)
	}
	var action monopoly.JailAction
	if !slices.Contains(available, monopoly.ROLL_DICE) {
		if slices.Contains(available, monopoly.CARD) {
			action = monopoly.CARD
		} else {
			action = monopoly.BAIL
		}
	} else if state.Round > 20 {
		action = monopoly.ROLL_DICE
	} else if slices.Contains(available, monopoly.CARD) {
		action = monopoly.CARD
	} else {
		action = monopoly.BAIL
	}
	if p.log != nil {
		p.log.Info("→ result", "action", action)
	}
	return action
}

func (p *NEATMonopolyPlayer) BuyDecision(player int, state monopoly.GameState, propertyId int) bool {
	property := state.Properties[propertyId]
	price := property.Price
	if p.log != nil {
		p.log.Warn("▶ BuyDecision", "propertyId", propertyId, "name", property.Name, "price", price)
	}
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player, propertyId)
	sensors.LoadBuyPriceInput(price)
	if p.log != nil {
		pl := state.Players[player]
		setCount, ownedByPlayer, ownedByOthers := getSetDetails(state, property.SetIndex, player)
		availableProps := 0
		for _, prop := range state.Properties {
			if prop.Owner == nil {
				availableProps++
			}
		}
		propType := "street"
		if property.SetIndex == monopoly.RailroadSetIndex {
			propType = "railroad"
		} else if property.SetIndex == monopoly.UtilitySetIndex {
			propType = "utility"
		}
		p.log.Debug("game state",
			"round", state.Round,
			"player_jailed", pl.IsJailed,
			"player_position", pl.CurrentPosition,
			"player_money", pl.Money,
			"property", property.Name,
			"property_type", propType,
			"property_price", price,
			"set_id", property.SetIndex,
			"set_needed", setCount-ownedByPlayer,
			"set_occupied", ownedByOthers,
			"available_properties", availableProps,
		)
		p.log.Info("network inputs", "values", formatFloatSlice(sensors))
	}
	outputList := p.GetDecision(sensors)
	if p.log != nil {
		p.log.Info("network outputs", "values", formatFloatSlice(outputList))
		p.log.Info("→ result", "OUT_BUY_PROPERTY", fmt.Sprintf("%.4f", outputList[OUT_BUY_PROPERTY]), "decision", boolToYesNo(outputList[OUT_BUY_PROPERTY] > 0.5))
	}
	return outputList[OUT_BUY_PROPERTY] > 0.5
}

func (p *NEATMonopolyPlayer) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int {
	property := state.Properties[propertyId]
	if p.log != nil {
		p.log.Warn("▶ BiddingDecision", "propertyId", propertyId, "name", property.Name, "currentPrice", currentPrice, "currentWinner", currentWinner)
	}
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player, propertyId)
	if currentWinner >= 0 {
		sensors.LoadEnemyInputs(state, currentWinner, propertyId)
	}
	bid := currentPrice + 10
	sensors.LoadBuyPriceInput(bid)
	if p.log != nil {
		pl := state.Players[player]
		p.log.Debug("game state",
			"round", state.Round,
			"player_money", pl.Money,
			"bid", bid,
		)
		p.log.Info("network inputs", "values", formatFloatSlice(sensors))
	}
	outputList := p.GetDecision(sensors)
	if p.log != nil {
		p.log.Info("network outputs", "values", formatFloatSlice(outputList))
	}
	if outputList[OUT_BUY_PROPERTY] <= 0.5 {
		if p.log != nil {
			p.log.Info("→ result", "OUT_BUY_PROPERTY", fmt.Sprintf("%.4f", outputList[OUT_BUY_PROPERTY]), "decision", "PASS (bid=0)")
		}
		return 0
	}
	if p.log != nil {
		p.log.Info("→ result", "OUT_BUY_PROPERTY", fmt.Sprintf("%.4f", outputList[OUT_BUY_PROPERTY]), "bid", bid)
	}
	return bid
}

func (p *NEATMonopolyPlayer) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, sellerOffer int, tradingPartnerId int) (bool, int) {
	property := state.Properties[propertyId]
	if p.log != nil {
		p.log.Warn("▶ BuyFromPlayerDecision", "propertyId", propertyId, "name", property.Name, "sellerOffer", sellerOffer, "tradingPartnerId", tradingPartnerId)
	}
	isLastTry := state.NegotiationRound >= cfg.MAX_TRADE_ROUNDS || state.SellerImpasse
	price := p.findMaxBuyPrice(state, player, propertyId, property.Price/2, sellerOffer, isLastTry)
	if price < sellerOffer {
		if p.log != nil {
			p.log.Info("→ result", "decision", "NO", "maxPrice", price, "sellerOffer", sellerOffer)
		}
		return false, 0
	}
	if p.log != nil {
		p.log.Info("→ result", "decision", "YES", "price", price)
	}
	return true, price
}

func (p *NEATMonopolyPlayer) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, buyerOffer int, tradingPartnerId int) (bool, int) {
	property := state.Properties[propertyId]
	if p.log != nil {
		p.log.Warn("▶ SellToPlayerDecision", "propertyId", propertyId, "name", property.Name, "buyerOffer", buyerOffer, "tradingPartnerId", tradingPartnerId)
	}
	isLastTry := state.NegotiationRound >= cfg.MAX_TRADE_ROUNDS || state.BuyerImpasse
	sell, price := p.findMinSellPrice(state, player, propertyId, buyerOffer, buyerOffer, tradingPartnerId, isLastTry)
	if p.log != nil {
		p.log.Info("→ result", "decision", boolToYesNo(sell), "price", price)
	}
	return sell, price
}

func (p *NEATMonopolyPlayer) findMinSellPrice(state monopoly.GameState, playerId int, propertyId int, minPrice int, counteroffer int, buyerId int, isLastTry bool) (bool, int) {
	price := minPrice
	decision := p.getSellDecision(playerId, state, propertyId, price, counteroffer, buyerId, isLastTry)
	for !decision {
		price += 10
		if price >= cfg.MAX_MONEY {
			return false, 0
		}
		decision = p.getSellDecision(playerId, state, propertyId, price, counteroffer, buyerId, isLastTry)
	}
	return true, price
}

func (p *NEATMonopolyPlayer) getSellDecision(playerId int, state monopoly.GameState, propertyId int, price int, counteroffer int, buyerId int, isLastTry bool) bool {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, playerId, propertyId)
	sensors.LoadSellPriceInput(price)
	sensors.LoadEnemyInputs(state, buyerId, propertyId)

	if counteroffer > 0 {
		sensors.LoadSellCounterofferInputs(counteroffer, isLastTry)
	}
	outputList := p.GetDecision(sensors)
	if p.log != nil {
		p.log.Info("  getSellDecision", "propertyId", propertyId, "price", price, "counteroffer", counteroffer, "isLastTry", isLastTry)
		p.log.Info("  network inputs", "values", formatFloatSlice(sensors))
		p.log.Info("  network outputs", "values", formatFloatSlice(outputList))
		p.log.Info("  OUT_SELL_PROPERTY", "value", fmt.Sprintf("%.4f", outputList[OUT_SELL_PROPERTY]), "decision", boolToYesNo(outputList[OUT_SELL_PROPERTY] > 0.5))
	}
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

func findKeyPropertiesToBuy(state monopoly.GameState, playerId int, propertiesToBuy []int) map[int]bool {
	keyProperties := make(map[int]bool)
	for _, propertyId := range propertiesToBuy {
		setCount, ownedByPlayer, ownedByOthers := getSetDetails(state, state.Properties[propertyId].SetIndex, playerId)
		if ownedByPlayer == setCount-1 && ownedByOthers == 1 {
			keyProperties[propertyId] = true
		}
	}
	return keyProperties
}
