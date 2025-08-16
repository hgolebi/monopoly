package neatnetwork

import (
	"fmt"
	"monopoly/pkg/monopoly"

	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

type NEATPlayerGroup struct {
	Id        int
	organisms []*genetics.Organism
	players   []*NEATMonopolyPlayer
	points    []int
}

func NewNEATPlayerGroup(id int, organisms []*genetics.Organism) (*NEATPlayerGroup, error) {
	if len(organisms) <= 0 || len(organisms) > 4 {
		errorMsg := fmt.Sprintf("Invalid number of organisms: %d. Expected between 1 and 4.", len(organisms))
		return nil, fmt.Errorf(errorMsg)
	}
	players := make([]*NEATMonopolyPlayer, 0, len(organisms))
	for i := 0; i < len(organisms); i++ {
		phenotype, err := organisms[i].Phenotype()
		if err != nil {
			errorMsg := fmt.Sprintf("Error getting phenotype for organism %d: %v\n", i, err)
			return nil, fmt.Errorf(errorMsg)
		}
		players = append(players, &NEATMonopolyPlayer{network: phenotype, player_id: i})
	}
	return &NEATPlayerGroup{
		Id:        id,
		organisms: organisms,
		players:   players,
		points:    make([]int, 0, len(organisms)),
	}, nil
}

func (t *NEATPlayerGroup) Init() int {
	return len(t.organisms)
}

func (t *NEATPlayerGroup) GetStdAction(player int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails {
	if player < 0 || player >= len(t.organisms) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.GetStdAction(player, state, availableActions)
}

func (t *NEATPlayerGroup) GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction {
	if player < 0 || player >= len(t.organisms) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.GetJailAction(player, state, available)
}

func (t *NEATPlayerGroup) BuyDecision(player int, state monopoly.GameState, propertyId int) bool {
	if player < 0 || player >= len(t.organisms) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.BuyDecision(player, state, propertyId)
}

func (t *NEATPlayerGroup) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	if player < 0 || player >= len(t.organisms) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.BuyFromPlayerDecision(player, state, propertyId, price)
}

func (t *NEATPlayerGroup) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	if player < 0 || player >= len(t.organisms) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.SellToPlayerDecision(player, state, propertyId, price)
}

func (t *NEATPlayerGroup) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int {
	if player < 0 || player >= len(t.organisms) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.BiddingDecision(player, state, propertyId, currentPrice, currentWinner)
}

func (t *NEATPlayerGroup) Finish(f monopoly.FinishOption, winner int, state monopoly.GameState) {}
