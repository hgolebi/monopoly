package neatnetwork

import (
	"errors"
	"fmt"
	"monopoly/pkg/monopoly"

	"github.com/yaricom/goNEAT/v4/neat/network"
)

type NEATMonopolyPlayer struct {
	network   *network.Network
	player_id int
	max_depth int
}

func NewNEATMonopolyPlayer(network *network.Network, player_id int) (*NEATMonopolyPlayer, error) {
	max_depth, err := network.MaxActivationDepthWithCap(0)
	if err != nil {
		return nil, err
	}
	if max_depth <= 0 {
		return nil, errors.New("Invalid network depth: " + fmt.Sprint(max_depth))
	}

	return &NEATMonopolyPlayer{
		network:   network,
		player_id: player_id,
		max_depth: max_depth,
	}, nil
}

func (p *NEATMonopolyPlayer) GetDecision(input []float64) []float64 {
	err := p.network.LoadSensors(input)
	if err != nil {
		panic("Error loading sensors: " + err.Error())
	}
	success, err := p.network.ForwardSteps(p.max_depth)
	if err != nil {
		panic("Error during forward steps: " + err.Error())
	}
	if !success {
		panic("Forward steps failed")
	}
	var output []float64
	for _, node := range p.network.Outputs {
		output = append(output, node.Activation)
	}
	return output
}

func (p *NEATMonopolyPlayer) GetStdAction(player int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player)
	sensors.LoadDecisionContext(STD_ACTION)
	sensors.LoadSellOfferTries(state.SellOfferTries)
	sensors.LoadBuyOfferTries(state.BuyOfferTries)
	if state.Charge > 0 {
		sensors.LoadCharge(state.Charge)

	}

	var result monopoly.ActionDetails
	propertyActions := transformAvailableActionsList(availableActions)
	for propertyId, availableActions := range propertyActions {
		sensors.LoadAvailableStdActions(availableActions)
		sensors.LoadPropertyId(propertyId)
		outputList := p.GetDecision(sensors)
		stdActionOutValues := GetStdActionOutputValues(outputList)
		var highest float64 = 0.0
		for _, action := range availableActions {
			if stdActionOutValues[action] > highest {
				highest = stdActionOutValues[action]
				result.Action = action
			}
		}
		if result.Action == monopoly.SELLOFFER || result.Action == monopoly.BUYOFFER {
			result.Price = GetPriceOutputValue(outputList)

			playerOutputs := GetPlayerOutputValues(outputList)
			var highest float64 = 0.0
			for playerId, value := range playerOutputs {
				if value > highest && !state.Players[playerId].IsBankrupt && playerId != player {
					highest = value
					result.PlayerId = playerId
				}
			}
		}
		if result.Action != monopoly.NOACTION {
			result.PropertyId = propertyId
			return result
		}
	}
	return result
}

func (p *NEATMonopolyPlayer) GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player)
	sensors.LoadDecisionContext(JAIL_DECISION)
	sensors.LoadAvailableJailActions(available)
	outputList := p.GetDecision(sensors)
	jailOutputs := GetJailOutputValues(outputList)
	highest := 0.0
	result := monopoly.ROLL_DICE
	for _, action := range available {
		if jailOutputs[action] > highest {
			highest = jailOutputs[action]
			result = action
		}
	}
	return result
}

func (p *NEATMonopolyPlayer) BuyDecision(player int, state monopoly.GameState, propertyId int) bool {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player)
	sensors.LoadDecisionContext(BUY_DECISION)
	sensors.LoadPropertyId(propertyId)
	sensors.LoadPrice(state.Properties[propertyId].Price)
	outputList := p.GetDecision(sensors)
	return outputList[outputs["BUY_DECISION"]] > 0.5
}

func (p *NEATMonopolyPlayer) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player)
	sensors.LoadDecisionContext(BUY_FROM_PLAYER)
	sensors.LoadPropertyId(propertyId)
	sensors.LoadPrice(price)

	outputList := p.GetDecision(sensors)
	return outputList[outputs["BUY_FROM_PLAYER"]] > 0.5
}

func (p *NEATMonopolyPlayer) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player)
	sensors.LoadDecisionContext(SELL_TO_PLAYER)
	sensors.LoadPropertyId(propertyId)
	sensors.LoadPrice(price)

	outputList := p.GetDecision(sensors)
	return outputList[outputs["SELL_TO_PLAYER"]] > 0.5
}

func (p *NEATMonopolyPlayer) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int {
	sensors := NewMonopolySensors()
	sensors.LoadState(state, player)
	sensors.LoadDecisionContext(BIDDING_DECISION)
	sensors.LoadPropertyId(propertyId)
	sensors.LoadBiddingInputs(currentPrice, currentWinner)
	outputList := p.GetDecision(sensors)
	decision := outputList[outputs["BIDDING_DECISION"]] > 0.5
	if !decision {
		return 0.0
	}
	return GetPriceOutputValue(outputList)
}

func transformAvailableActionsList(actions monopoly.FullActionList) map[int][]monopoly.StdAction {
	propertyActions := make(map[int][]monopoly.StdAction)
	for _, propertyID := range actions.MortgageList {
		propertyActions[propertyID] = append(propertyActions[propertyID], monopoly.MORTGAGE)
	}
	for _, propertyID := range actions.BuyOutList {
		propertyActions[propertyID] = append(propertyActions[propertyID], monopoly.BUYOUT)
	}
	for _, propertyID := range actions.SellPropertyList {
		propertyActions[propertyID] = append(propertyActions[propertyID], monopoly.SELLOFFER)
	}
	for _, propertyID := range actions.BuyPropertyList {
		propertyActions[propertyID] = append(propertyActions[propertyID], monopoly.BUYOFFER)
	}
	for _, propertyID := range actions.BuyHouseList {
		propertyActions[propertyID] = append(propertyActions[propertyID], monopoly.BUYHOUSE)
	}
	for _, propertyID := range actions.SellHouseList {
		propertyActions[propertyID] = append(propertyActions[propertyID], monopoly.SELLHOUSE)
	}

	return propertyActions
}
