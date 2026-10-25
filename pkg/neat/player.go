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
	// ForGame returns an instance of the player for a single game. Instances share the score statistics,
	// but each one has its own state, so they can play different games concurrently.
	ForGame() (MonopolyPlayer, error)
}

// playerStats are the score statistics of an organism, shared by all of its game instances.
type playerStats struct {
	mutex sync.Mutex
	score int
	wins  int
	draws int
}

type NEATMonopolyPlayer struct {
	fast      *fastNetwork     // compiled network, nil for networks with recurrent links
	network   *network.Network // used when fast is nil; activation state lives in the nodes, so it must not be shared between games
	organism  *genetics.Organism
	max_depth int
	stats     *playerStats
	log       *BotDecisionLogger
	auction   auctionBids // per game: the network's answers in the auction in progress
}

// auctionBids remembers the network's answers within one auction. The inputs do not change during an auction,
// only the price, so with a monotonic answer an accepted price answers every lower one and a rejected price every higher one.
type auctionBids struct {
	auctionId    int
	propertyId   int
	round        int
	sensors      MonopolySensors
	acceptedUpTo int // highest price accepted so far, -1 if none
	rejectedFrom int // lowest price rejected so far, above MAX_MONEY if none
	lookahead    int // until a price is rejected, the next query asks this many steps above the bid; doubles after every accepted query
}

func NewNEATMonopolyPlayer(organism *genetics.Organism) (*NEATMonopolyPlayer, error) {
	player := &NEATMonopolyPlayer{
		organism: organism,
		stats:    &playerStats{},
	}
	fast, err := compileFastNetwork(organism.Genotype)
	if err == nil {
		player.fast = fast
		return player, nil
	}
	if !errors.Is(err, errRecurrentNetwork) {
		neat.DebugLog(fmt.Sprintf("Organism %d uses the goNEAT network: %v", organism.Genotype.Id, err))
	}

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
	player.network, player.max_depth = network, max_depth
	return player, nil
}

// ForGame returns an instance of the player with its own network state for one game.
// A compiled network only needs new values; otherwise the network is built again from the genome.
// Genesis modifies the genome, so it must not be called concurrently for the same organism.
func (p *NEATMonopolyPlayer) ForGame() (MonopolyPlayer, error) {
	if p.fast != nil {
		return &NEATMonopolyPlayer{
			fast:      p.fast.newInstance(),
			organism:  p.organism,
			max_depth: p.max_depth,
			stats:     p.stats,
			log:       p.log,
		}, nil
	}
	network, err := p.organism.Genotype.Genesis(p.organism.Genotype.Id)
	if err != nil {
		return nil, fmt.Errorf("error building network for organism %d: %w", p.organism.Genotype.Id, err)
	}
	return &NEATMonopolyPlayer{
		network:   network,
		organism:  p.organism,
		max_depth: p.max_depth,
		stats:     p.stats,
		log:       p.log,
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
	p.stats.mutex.Lock()
	defer p.stats.mutex.Unlock()
	return p.stats.score
}

func (p *NEATMonopolyPlayer) GetOrganism() *genetics.Organism {
	return p.organism
}

// decisionHook is called on every network activation. It is nil outside of the performance tests,
// which set it while no game is running to count activations per decision path and network.
var decisionHook func(p *NEATMonopolyPlayer)

func (p *NEATMonopolyPlayer) GetDecision(input []float64) []float64 {
	if decisionHook != nil {
		decisionHook(p)
	}
	if p.fast != nil {
		output, err := p.fast.activate(input)
		if err != nil {
			panic("Error activating network: " + err.Error())
		}
		return output
	}
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
	if p.log != nil {
		p.log.Warn("▶ GetStdAction", "player", player, "round", state.Round, "money", state.Players[player].Money)
	}

	if state.Charge > 0 {
		action := getChargeAction(player, state, availableActions)
		if p.log != nil {
			p.log.Info("→ result (charge)", "charge", state.Charge, "action", action.Action, "propertyId", action.PropertyId)
		}
		return action
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

// getBuyOutCandidates returns mortgaged properties the network is asked to buy out, and whether raising money
// (mortgaging) would unblock a buyout.
//   - Streets: base rent without a full set rarely pays back the buyout, so only full sets (needed to build houses).
//   - Railroads and utilities never get houses and their rent grows with the number owned, so every one is a candidate.
//     They are bought out only with cash at hand — raising money could mortgage one railroad to buy out another.
func getBuyOutCandidates(playerId int, state monopoly.GameState, buyOutList []int) (candidates []int, needMoney bool) {
	player := state.Players[playerId]
	for _, propertyId := range buyOutList {
		property := state.Properties[propertyId]
		isStreet := property.SetIndex != monopoly.RailroadSetIndex && property.SetIndex != monopoly.UtilitySetIndex
		if isStreet {
			setCount, ownedByPlayer, _ := getSetDetails(state, property.SetIndex, playerId)
			if setCount != ownedByPlayer {
				continue
			}
		}
		if property.GetBuyoutPrice() >= player.Money {
			needMoney = needMoney || isStreet
			continue
		}
		candidates = append(candidates, propertyId)
	}
	return candidates, needMoney
}

func getChargeAction(playerId int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails {
	mortgageList := slices.Sorted(slices.Values(availableActions.MortgageList))
	for _, propertyId := range mortgageList {
		setCount, ownedByPlayer, _ := getSetDetails(state, state.Properties[propertyId].SetIndex, playerId)
		if setCount != ownedByPlayer {
			return monopoly.ActionDetails{Action: monopoly.MORTGAGE, PropertyId: propertyId}
		}
	}
	if len(mortgageList) > 0 {
		return monopoly.ActionDetails{Action: monopoly.MORTGAGE, PropertyId: mortgageList[0]}
	}

	if len(availableActions.SellHouseList) > 0 {
		return monopoly.ActionDetails{Action: monopoly.SELLHOUSE, PropertyId: slices.Min(availableActions.SellHouseList)}
	}
	return monopoly.ActionDetails{Action: monopoly.NOACTION}
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
			Players:    opponents(player, state), // the engine skips the ones who cannot afford the starting price
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
		setCount, ownedByPlayer, ownedByOthers := getSetDetails(state, state.Properties[propertyId].SetIndex, player)
		if setCount == ownedByPlayer {
			continue
		}
		property := state.Properties[propertyId]

		sensors := NewMonopolySensors()
		sensors.LoadState(state, player, propertyId)
		sensors.LoadSellPriceInput(property.Price / 2)

		if p.log != nil {
			pl := state.Players[player]
			availableProps := countAvailableProperties(state)
			p.log.Info("  mortgage candidate", "propertyId", propertyId, "name", property.Name, "mortgageValue", property.Price/2)
			p.log.Debug("game state",
				"round", state.Round,
				"player_jailed", pl.IsJailed,
				"player_position", pl.CurrentPosition,
				"player_money", pl.Money,
				"property", property.Name,
				"property_price", property.Price,
				"mortgage_value", property.Price/2,
				"set_id", property.SetIndex,
				"set_needed", setCount-ownedByPlayer,
				"set_occupied", ownedByOthers,
				"available_properties", availableProps,
			)
			p.log.Info("  network inputs", "values", formatFloatSlice(sensors))
		}
		outputList := p.GetDecision(sensors)
		if p.log != nil {
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

// sellCandidate is a property the bot can put up for a sell auction.
type sellCandidate struct {
	propertyId int
	minPrice   int // below this price mortgaging is better than selling
	maxPrice   int // nominal value, capped by the cash of the richest opponent
	buyerId    int // opponent who gains the most from buying it, shown to the network as the enemy
}

// getSellCandidates returns the properties the bot may sell in an auction:
//   - properties in the player's full sets are skipped, like in tryMortgage, because selling one breaks the set;
//   - mortgaged properties are allowed (the buyer takes over the mortgage), so their value is half the price;
//   - a property is skipped when no opponent can pay its minimum price.
func getSellCandidates(playerId int, state monopoly.GameState, sellList []int) []sellCandidate {
	var candidates []sellCandidate
	for _, propertyId := range sellList {
		property := state.Properties[propertyId]
		setCount, ownedByPlayer, _ := getSetDetails(state, property.SetIndex, playerId)
		if setCount == ownedByPlayer {
			continue
		}
		minPrice, nominal := property.Price/2, property.Price
		if property.IsMortgaged {
			minPrice, nominal = cfg.PRICE_STEP, property.Price/2
		}

		buyerId, buyerNeeds, richest := -1, 0, 0
		for _, opponentId := range opponents(playerId, state) {
			money := state.Players[opponentId].Money
			richest = max(richest, money)
			if money < minPrice {
				continue
			}
			_, ownedByOpponent, _ := getSetDetails(state, property.SetIndex, opponentId)
			needs := setCount - ownedByOpponent
			if buyerId == -1 || needs < buyerNeeds || (needs == buyerNeeds && money > state.Players[buyerId].Money) {
				buyerId, buyerNeeds = opponentId, needs
			}
		}
		if buyerId == -1 {
			continue
		}
		candidates = append(candidates, sellCandidate{
			propertyId: propertyId,
			minPrice:   minPrice,
			maxPrice:   max(minPrice, min(nominal, richest)),
			buyerId:    buyerId,
		})
	}
	return candidates
}

// opponents returns the players that are still in the game, except playerId.
func opponents(playerId int, state monopoly.GameState) []int {
	var ids []int
	for id, player := range state.Players {
		if id != playerId && !player.IsBankrupt {
			ids = append(ids, id)
		}
	}
	return ids
}

// trySellProperty picks the property to put up for a sell auction and its starting price.
// The network is asked once per candidate at the candidate's maximum price, the argmax above 0.5 is chosen,
// and the starting price is the lowest price the network still accepts. The seller has to sell to the highest
// bidder, so the starting price is the only protection; buyers who value the property more bid it up.
func (p *NEATMonopolyPlayer) trySellProperty(player int, state monopoly.GameState, availableActions monopoly.FullActionList) (decision bool, propertyId int, price int) {
	if !slices.Contains(availableActions.Actions, monopoly.SELLOFFER) {
		return false, 0, 0
	}
	candidates := getSellCandidates(player, state, availableActions.SellPropertyList)
	if len(candidates) == 0 {
		return false, 0, 0
	}
	if p.log != nil {
		p.log.Info("▶ trySellProperty", "candidates", len(candidates))
	}
	outputs := make(map[int]float64)
	chosen := make(map[int]sellCandidate)
	sensorsOf := make(map[int]MonopolySensors)
	for _, c := range candidates {
		sensors := p.sellToPlayerSensors(player, state, c.propertyId, c.buyerId, 0, false)
		sensors.LoadSellPriceInput(c.maxPrice)
		outputList := p.GetDecision(sensors)
		if p.log != nil {
			p.log.Info("  sell candidate", "propertyId", c.propertyId, "maxPrice", c.maxPrice, "buyer", c.buyerId,
				"OUT_SELL_PROPERTY", fmt.Sprintf("%.4f", outputList[OUT_SELL_PROPERTY]))
		}
		if outputList[OUT_SELL_PROPERTY] > 0.5 {
			outputs[c.propertyId] = outputList[OUT_SELL_PROPERTY]
			chosen[c.propertyId] = c
			sensorsOf[c.propertyId] = sensors
		}
	}
	decision, propertyId, _ = getBestDecision(outputs)
	if !decision {
		if p.log != nil {
			p.log.Info("→ result", "decision", "NO")
		}
		return false, 0, 0
	}
	c := chosen[propertyId]
	sensors := sensorsOf[propertyId]
	accept := p.priceAcceptor(sensors, sensors.LoadSellPriceInput, OUT_SELL_PROPERTY)
	price, _ = lowestAcceptedPrice(c.minPrice, c.maxPrice, func(price int) bool {
		return price == c.maxPrice || accept(price) // the maximum price was accepted above
	})
	if p.log != nil {
		p.log.Info("→ result", "decision", "YES", "propertyId", propertyId, "price", price)
	}
	return true, propertyId, price
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
		setCount, ownedByPlayer, ownedByOthers := getSetDetails(state, property.SetIndex, playerId)

		sensors := NewMonopolySensors()
		sensors.LoadState(state, playerId, propertyId)
		sensors.LoadBuyPriceInput(property.HousePrice)

		if p.log != nil {
			availableProps := countAvailableProperties(state)
			p.log.Info("  house candidate", "propertyId", propertyId, "name", property.Name, "housePrice", property.HousePrice)
			p.log.Debug("game state",
				"round", state.Round,
				"player_jailed", player.IsJailed,
				"player_position", player.CurrentPosition,
				"player_money", player.Money,
				"property", property.Name,
				"property_price", property.Price,
				"house_price", property.HousePrice,
				"set_id", property.SetIndex,
				"set_needed", setCount-ownedByPlayer,
				"set_occupied", ownedByOthers,
				"available_properties", availableProps,
			)
			p.log.Info("  network inputs", "values", formatFloatSlice(sensors))
		}
		outputList := p.GetDecision(sensors)
		if p.log != nil {
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
	candidates, needMoney := getBuyOutCandidates(playerId, state, availableActions.BuyOutList)
	for _, propertyId := range candidates {
		property := state.Properties[propertyId]
		setCount, ownedByPlayer, ownedByOthers := getSetDetails(state, property.SetIndex, playerId)
		buyoutPrice := property.GetBuyoutPrice()

		sensors := NewMonopolySensors()
		sensors.LoadState(state, playerId, propertyId)
		sensors.LoadBuyPriceInput(buyoutPrice)

		if p.log != nil {
			availableProps := countAvailableProperties(state)
			p.log.Info("  buyout candidate", "propertyId", propertyId, "name", property.Name, "buyoutPrice", buyoutPrice)
			p.log.Debug("game state",
				"round", state.Round,
				"player_jailed", player.IsJailed,
				"player_position", player.CurrentPosition,
				"player_money", player.Money,
				"property", property.Name,
				"property_price", property.Price,
				"buyout_price", buyoutPrice,
				"set_id", property.SetIndex,
				"set_needed", setCount-ownedByPlayer,
				"set_occupied", ownedByOthers,
				"available_properties", availableProps,
			)
			p.log.Info("  network inputs", "values", formatFloatSlice(sensors))
		}
		outputList := p.GetDecision(sensors)
		if p.log != nil {
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
	money := state.Players[playerId].Money
	outputs := make(map[int]float64)
	sensorsOf := make(map[int]MonopolySensors)
	for propertyId := range keyProperties {
		property := state.Properties[propertyId]
		minPrice := property.Price / 2
		if minPrice >= money {
			needMoney = true
			continue
		}
		sensors := p.buyFromPlayerSensors(playerId, state, propertyId, -1, false)
		sensors.LoadBuyPriceInput(minPrice)
		outputList := p.GetDecision(sensors)
		if p.log != nil {
			p.log.Info("  key property candidate", "propertyId", propertyId, "name", property.Name, "minPrice", minPrice,
				"OUT_BUY_PROPERTY", fmt.Sprintf("%.4f", outputList[OUT_BUY_PROPERTY]))
		}
		if outputList[OUT_BUY_PROPERTY] > 0.5 {
			outputs[propertyId] = outputList[OUT_BUY_PROPERTY]
			sensorsOf[propertyId] = sensors
		}
	}
	decision, propertyId, _ = getBestDecision(outputs)
	if !decision {
		if p.log != nil {
			p.log.Info("→ result", "decision", "NO", "needMoney", needMoney)
		}
		return decision, propertyId, 0, needMoney
	}
	// opening offer: the highest price the network accepts (it accepted the minimum price above)
	minPrice := state.Properties[propertyId].Price / 2
	sensors := sensorsOf[propertyId]
	accept := p.priceAcceptor(sensors, sensors.LoadBuyPriceInput, OUT_BUY_PROPERTY)
	price, _ = highestAcceptedPrice(minPrice, min(cfg.MAX_MONEY, money), func(price int) bool {
		return price == minPrice || accept(price)
	})
	if p.log != nil {
		p.log.Info("→ result", "decision", "YES", "propertyId", propertyId, "price", price)
	}
	return decision, propertyId, price, needMoney
}

// buyFromPlayerSensors loads the inputs for buying propertyId from its owner, without the price.
// The price is set by the caller, so one set of sensors serves a whole price search.
func (p *NEATMonopolyPlayer) buyFromPlayerSensors(player int, state monopoly.GameState, propertyId int, sellerOffer int, isLastTry bool) MonopolySensors {
	property := state.Properties[propertyId]
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player, propertyId)
	sensors.LoadEnemyInputs(state, property.Owner.ID, propertyId)
	if sellerOffer > 0 {
		sensors.LoadBuyCounterofferInputs(sellerOffer, isLastTry)
	}

	if p.log != nil {
		pl := state.Players[player]
		enemy := state.Players[property.Owner.ID]
		setCount, ownedByPlayer, ownedByOthers := getSetDetails(state, property.SetIndex, player)
		_, enemyOwned, _ := getSetDetails(state, property.SetIndex, property.Owner.ID)
		p.log.Debug("game state (buy from player)",
			"round", state.Round,
			"player_jailed", pl.IsJailed,
			"player_position", pl.CurrentPosition,
			"player_money", pl.Money,
			"property", property.Name,
			"property_price", property.Price,
			"seller_offer", sellerOffer,
			"is_last_try", isLastTry,
			"set_id", property.SetIndex,
			"set_needed", setCount-ownedByPlayer,
			"set_occupied", ownedByOthers,
			"available_properties", countAvailableProperties(state),
			"enemy_money", enemy.Money,
			"enemy_set_needed", setCount-enemyOwned,
		)
	}
	return sensors
}

// priceAcceptor returns a function that loads a price into sensors with loadPrice, activates the network
// and tells whether the output is above 0.5. Only the price input changes between calls.
func (p *NEATMonopolyPlayer) priceAcceptor(sensors MonopolySensors, loadPrice func(int), output OutputID) func(price int) bool {
	return func(price int) bool {
		loadPrice(price)
		outputList := p.GetDecision(sensors)
		if p.log != nil {
			p.log.Info("  price query", "price", price, "inputs", formatFloatSlice(sensors), "outputs", formatFloatSlice(outputList),
				"decision", boolToYesNo(outputList[output] > 0.5))
		}
		return outputList[output] > 0.5
	}
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
	} else if slices.Contains(available, monopoly.BAIL) && state.Players[player].Money >= cfg.JAIL_BAIL {
		action = monopoly.BAIL
	} else {
		// Paying bail without cash would force mortgaging — rolling is free
		action = monopoly.ROLL_DICE
	}
	if p.log != nil {
		p.log.Error("→ result", "action", action)
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
		availableProps := countAvailableProperties(state)
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
		p.log.Error("→ result", "OUT_BUY_PROPERTY", fmt.Sprintf("%.4f", outputList[OUT_BUY_PROPERTY]), "decision", boolToYesNo(outputList[OUT_BUY_PROPERTY] > 0.5))
	}
	return outputList[OUT_BUY_PROPERTY] > 0.5
}

func (p *NEATMonopolyPlayer) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int {
	property := state.Properties[propertyId]
	if p.log != nil {
		p.log.Warn("▶ BiddingDecision", "propertyId", propertyId, "name", property.Name, "currentPrice", currentPrice, "currentWinner", currentWinner)
	}
	bid := currentPrice + cfg.PRICE_STEP
	if currentWinner < 0 {
		bid = currentPrice // nobody has bid yet, so the starting price is a valid bid
	}
	if bid > state.Players[player].Money || !p.acceptsBid(player, state, propertyId, bid, currentWinner) {
		if p.log != nil {
			p.log.Error("→ result", "bid", bid, "decision", "PASS (bid=0)")
		}
		return 0
	}
	if p.log != nil {
		p.log.Error("→ result", "bid", bid)
	}
	return bid
}

// acceptsBid tells whether the network wants to pay bid in the auction in progress. Within one auction it reuses
// the earlier answers, and each query asks about a price ahead of the bid (1, 2, 4, ... steps after accepted
// queries), so the bot's first bid costs one activation, as before, and a bidding war of d steps O(log d)
// activations instead of d. With a monotonic answer the decisions are the same as asking at every bid.
// The inputs are loaded at the bot's first bid in the auction, so a later change of the leading bidder is not seen.
// Without an auction id (a caller other than the engine's auction) every call activates the network.
func (p *NEATMonopolyPlayer) acceptsBid(player int, state monopoly.GameState, propertyId int, bid int, currentWinner int) bool {
	a := &p.auction
	if state.AuctionId == 0 || a.auctionId != state.AuctionId || a.propertyId != propertyId || a.round != state.Round {
		sensors := NewMonopolySensors()
		sensors.LoadState(state, player, propertyId)
		if currentWinner >= 0 {
			sensors.LoadEnemyInputs(state, currentWinner, propertyId)
		}
		*a = auctionBids{
			auctionId:    state.AuctionId,
			propertyId:   propertyId,
			round:        state.Round,
			sensors:      sensors,
			acceptedUpTo: -1,
			rejectedFrom: cfg.MAX_MONEY + 1,
			lookahead:    1,
		}
	}
	accept := p.priceAcceptor(a.sensors, a.sensors.LoadBuyPriceInput, OUT_BUY_PROPERTY)
	return a.accepts(bid, min(cfg.MAX_MONEY, state.Players[player].Money), accept)
}

// accepts answers from the remembered answers or asks accept at a price between bid and maxPrice (bid <= maxPrice).
func (a *auctionBids) accepts(bid int, maxPrice int, accept func(price int) bool) bool {
	for {
		if bid <= a.acceptedUpTo {
			return true
		}
		if bid >= a.rejectedFrom {
			return false
		}
		var probe int
		if a.rejectedFrom > maxPrice {
			// no rejected price yet: look further ahead after every accepted probe
			probe = min(bid+(a.lookahead-1)*cfg.PRICE_STEP, maxPrice)
		} else {
			// the answer changes between bid and the lowest rejected price: probe the middle
			probe = bid + (a.rejectedFrom-bid)/cfg.PRICE_STEP/2*cfg.PRICE_STEP
		}
		if accept(probe) {
			a.acceptedUpTo = probe
			a.lookahead *= 2
		} else {
			a.rejectedFrom = probe
		}
	}
}

func (p *NEATMonopolyPlayer) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, sellerOffer int, tradingPartnerId int) (bool, int) {
	property := state.Properties[propertyId]
	if p.log != nil {
		p.log.Warn("▶ BuyFromPlayerDecision", "propertyId", propertyId, "name", property.Name, "sellerOffer", sellerOffer, "tradingPartnerId", tradingPartnerId)
	}
	if sellerOffer > state.Players[player].Money {
		if p.log != nil {
			p.log.Error("→ result", "decision", "NO (cannot afford)", "sellerOffer", sellerOffer)
		}
		return false, 0
	}
	// The bot either accepts the counteroffer or withdraws. With a monotonic answer the highest acceptable price
	// reaches sellerOffer exactly when sellerOffer itself is accepted, so a single activation decides.
	isLastTry := state.NegotiationRound >= cfg.MAX_TRADE_ROUNDS || state.SellerImpasse
	sensors := p.buyFromPlayerSensors(player, state, propertyId, sellerOffer, isLastTry)
	if !p.priceAcceptor(sensors, sensors.LoadBuyPriceInput, OUT_BUY_PROPERTY)(sellerOffer) {
		if p.log != nil {
			p.log.Error("→ result", "decision", "NO", "sellerOffer", sellerOffer)
		}
		return false, 0
	}
	if p.log != nil {
		p.log.Error("→ result", "decision", "YES", "price", sellerOffer)
	}
	return true, sellerOffer
}

func (p *NEATMonopolyPlayer) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, buyerOffer int, tradingPartnerId int) (bool, int) {
	property := state.Properties[propertyId]
	if p.log != nil {
		p.log.Warn("▶ SellToPlayerDecision", "propertyId", propertyId, "name", property.Name, "buyerOffer", buyerOffer, "tradingPartnerId", tradingPartnerId)
	}
	isLastTry := state.NegotiationRound >= cfg.MAX_TRADE_ROUNDS || state.BuyerImpasse
	sensors := p.sellToPlayerSensors(player, state, propertyId, tradingPartnerId, buyerOffer, isLastTry)
	accept := p.priceAcceptor(sensors, sensors.LoadSellPriceInput, OUT_SELL_PROPERTY)
	// the lowest acceptable price: buyerOffer accepts the offer, a higher price is a counteroffer
	price, sell := lowestAcceptedPrice(buyerOffer, max(buyerOffer, cfg.MAX_MONEY-cfg.PRICE_STEP), accept)
	if p.log != nil {
		p.log.Error("→ result", "decision", boolToYesNo(sell), "price", price)
	}
	return sell, price
}

// sellToPlayerSensors loads the inputs for selling propertyId to buyerId, without the price.
// counteroffer is the buyer's offer in a negotiation, 0 when there is none.
func (p *NEATMonopolyPlayer) sellToPlayerSensors(playerId int, state monopoly.GameState, propertyId int, buyerId int, counteroffer int, isLastTry bool) MonopolySensors {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, playerId, propertyId)
	sensors.LoadEnemyInputs(state, buyerId, propertyId)
	if counteroffer > 0 {
		sensors.LoadSellCounterofferInputs(counteroffer, isLastTry)
	}

	if p.log != nil {
		property := state.Properties[propertyId]
		pl := state.Players[playerId]
		buyer := state.Players[buyerId]
		setCount, ownedByPlayer, ownedByOthers := getSetDetails(state, property.SetIndex, playerId)
		_, buyerOwned, _ := getSetDetails(state, property.SetIndex, buyerId)
		p.log.Debug("game state (sell to player)",
			"round", state.Round,
			"player_jailed", pl.IsJailed,
			"player_position", pl.CurrentPosition,
			"player_money", pl.Money,
			"property", property.Name,
			"property_price", property.Price,
			"buyer_offer", counteroffer,
			"is_last_try", isLastTry,
			"set_id", property.SetIndex,
			"set_needed", setCount-ownedByPlayer,
			"set_occupied", ownedByOthers,
			"available_properties", countAvailableProperties(state),
			"enemy_money", buyer.Money,
			"enemy_set_needed", setCount-buyerOwned,
		)
	}
	return sensors
}

func (p *NEATMonopolyPlayer) AddScore(points int) {
	p.stats.mutex.Lock()
	p.stats.score += points
	p.stats.mutex.Unlock()
}

func (p *NEATMonopolyPlayer) AddWin() {
	p.stats.mutex.Lock()
	p.stats.wins++
	p.stats.mutex.Unlock()
}

func (p *NEATMonopolyPlayer) AddDraw() {
	p.stats.mutex.Lock()
	p.stats.draws++
	p.stats.mutex.Unlock()
}

func (p *NEATMonopolyPlayer) GetWins() int {
	p.stats.mutex.Lock()
	defer p.stats.mutex.Unlock()
	return p.stats.wins
}
func (p *NEATMonopolyPlayer) GetDraws() int {
	p.stats.mutex.Lock()
	defer p.stats.mutex.Unlock()
	return p.stats.draws
}

func getBestDecision(decisions map[int]float64) (bestFound bool, bestId int, bestValue float64) {
	if len(decisions) == 0 {
		return false, 0, 0.0
	}

	bestId = -1
	bestValue = -1.0
	for id, value := range decisions {
		// ties go to the lower id, so the choice does not depend on the map iteration order
		if value > bestValue || (value == bestValue && id < bestId) {
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

func countAvailableProperties(state monopoly.GameState) int {
	count := 0
	for _, prop := range state.Properties {
		if prop.Owner == nil {
			count++
		}
	}
	return count
}
