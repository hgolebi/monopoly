package neatnetwork

import (
	"fmt"
	cfg "monopoly/pkg/config"
	"monopoly/pkg/monopoly"
)

type NEATPlayerGroup struct {
	Id           int
	players      []MonopolyPlayer
	gameFinished bool
}

func NewNEATPlayerGroup(id int, players []MonopolyPlayer) (*NEATPlayerGroup, error) {
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

func (t *NEATPlayerGroup) Init() []string {
	player_names := make([]string, len(t.players))
	for i, player := range t.players {
		player_names[i] = player.GetName()
	}
	return player_names
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

func (t *NEATPlayerGroup) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, sellerOffer int, tradingPartnerId int) (bool, int) {
	if player < 0 || player >= len(t.players) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.BuyFromPlayerDecision(player, state, propertyId, sellerOffer, tradingPartnerId)
}

func (t *NEATPlayerGroup) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, buyerOffer int, tradingPartnerId int) (bool, int) {
	if player < 0 || player >= len(t.players) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.SellToPlayerDecision(player, state, propertyId, buyerOffer, tradingPartnerId)
}

func (t *NEATPlayerGroup) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int {
	if player < 0 || player >= len(t.players) {
		panic("Invalid player index")
	}

	p := t.players[player]
	return p.BiddingDecision(player, state, propertyId, currentPrice, currentWinner)
}

// Finish scores the game. The winner (last player not bankrupt) gets FIRST_PLACE_SCORE points.
// When the game ends by the round limit, the player with the highest net worth (chosen by the engine)
// gets DRAW_WINNER_SCORE points and all players alive get a draw. A draw without survivors gives no points.
func (t *NEATPlayerGroup) Finish(f monopoly.FinishOption, winner int, state monopoly.GameState) {
	if t.gameFinished {
		return
	}
	t.gameFinished = true
	switch f {
	case monopoly.WIN:
		t.players[winner].AddScore(cfg.FIRST_PLACE_SCORE)
		t.players[winner].AddWin()
	case monopoly.ROUND_LIMIT:
		t.players[winner].AddScore(cfg.DRAW_WINNER_SCORE)
		for i, p := range state.Players {
			if !p.IsBankrupt {
				t.players[i].AddDraw()
			}
		}
	}
}
