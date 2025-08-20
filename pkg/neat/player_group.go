package neatnetwork

import (
	"fmt"
	cfg "monopoly/pkg/config"
	"monopoly/pkg/monopoly"
)

type NEATPlayerGroup struct {
	Id           int
	players      []*NEATMonopolyPlayer
	gameFinished bool
}

func NewNEATPlayerGroup(id int, players []*NEATMonopolyPlayer) (*NEATPlayerGroup, error) {
	if len(players) <= 0 || len(players) > 4 {
		errorMsg := fmt.Sprintf("Invalid number of players: %d. Expected between 1 and 4.", len(players))
		return nil, fmt.Errorf(errorMsg)
	}
	return &NEATPlayerGroup{
		Id:           id,
		players:      players,
		gameFinished: false,
	}, nil
}

func (t *NEATPlayerGroup) Init() int {
	return len(t.players)
}

func (t *NEATPlayerGroup) GetStdAction(player int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails {
	if player < 0 || player >= len(t.players) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.GetStdAction(player, state, availableActions)
}

func (t *NEATPlayerGroup) GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction {
	if player < 0 || player >= len(t.players) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.GetJailAction(player, state, available)
}

func (t *NEATPlayerGroup) BuyDecision(player int, state monopoly.GameState, propertyId int) bool {
	if player < 0 || player >= len(t.players) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.BuyDecision(player, state, propertyId)
}

func (t *NEATPlayerGroup) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	if player < 0 || player >= len(t.players) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.BuyFromPlayerDecision(player, state, propertyId, price)
}

func (t *NEATPlayerGroup) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	if player < 0 || player >= len(t.players) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.SellToPlayerDecision(player, state, propertyId, price)
}

func (t *NEATPlayerGroup) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int {
	if player < 0 || player >= len(t.players) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.BiddingDecision(player, state, propertyId, currentPrice, currentWinner)
}

func (t *NEATPlayerGroup) Finish(f monopoly.FinishOption, winner int, state monopoly.GameState) {
	if t.gameFinished {
		return
	}
	t.gameFinished = true
	second_place := -1
	highest_round := -1
	for i, player := range state.Players {
		if player.IsBankrupt {
			if player.RoundWhenBankrupted > highest_round {
				highest_round = player.RoundWhenBankrupted
				second_place = i
			}
			if player.RoundWhenBankrupted <= cfg.PUNISHMENT_FIRST_THRESHOLD {
				t.players[player.ID].organism.Fitness += cfg.HIGHEST_PUNISHMENT
			} else if player.RoundWhenBankrupted <= cfg.PUNISHMENT_SECOND_THRESHOLD {
				t.players[player.ID].organism.Fitness += cfg.SECOND_HIGHEST_PUNISHMENT
			} else {
				t.players[player.ID].organism.Fitness += float64(cfg.MAX_ROUNDS - player.RoundWhenBankrupted)
			}
		}
	}

	switch f {
	case monopoly.ROUND_LIMIT:
		t.players[winner].organism.Fitness += cfg.ROUND_LIMIT_WINNER_SCORE
	case monopoly.WIN:
		t.players[winner].organism.Fitness += cfg.FIRST_PLACE_SCORE
		if second_place != -1 {
			t.players[second_place].organism.Fitness += cfg.SECOND_PLACE_SCORE
		}
	}
}
