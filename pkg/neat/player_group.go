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

// func (t *NEATPlayerGroup) OLDFinish(f monopoly.FinishOption, winner int, state monopoly.GameState) {
// 	pointsMap := map[int]int{
// 		0: 0,
// 		1: 0,
// 		2: 0,
// 		3: 0,
// 	}
// 	if t.gameFinished {
// 		return
// 	}
// 	t.gameFinished = true
// 	second_place := -1
// 	highest_round := -1
// 	for i, player := range state.Players {
// 		pointsMap[i] += cfg.ROUND_SCORE * player.RoundsPlayed
// 		if player.IsBankrupt {
// 			if player.RoundsPlayed > highest_round {
// 				highest_round = player.RoundsPlayed
// 				second_place = i
// 			}
// 			if player.RoundsPlayed <= cfg.PUNISHMENT_FIRST_THRESHOLD {
// 				pointsMap[player.ID] += cfg.HIGHEST_PUNISHMENT
// 			} else if player.RoundsPlayed <= cfg.PUNISHMENT_SECOND_THRESHOLD {
// 				pointsMap[player.ID] += cfg.SECOND_HIGHEST_PUNISHMENT
// 			} else {
// 				pointsMap[player.ID] += cfg.MAX_ROUNDS - player.RoundsPlayed
// 			}
// 		}
// 	}

// 	switch f {
// 	case monopoly.ROUND_LIMIT:
// 		pointsMap[winner] += cfg.ROUND_LIMIT_WINNER_SCORE
// 	case monopoly.WIN:
// 		for _, propertyId := range state.Players[winner].Properties {
// 			property := state.Properties[propertyId]
// 			if !property.IsMortgaged {
// 				pointsMap[winner] += cfg.POINT_PER_PROPERTY
// 				pointsMap[winner] += cfg.POINTS_PER_HOUSE * property.Houses
// 			}
// 		}
// 		pointsMap[winner] = cfg.FIRST_PLACE_SCORE
// 		if second_place != -1 {
// 			pointsMap[second_place] += cfg.SECOND_PLACE_SCORE
// 		}
// 	}
// 	for i, p := range t.players {
// 		p.AddScore(pointsMap[i])
// 		fmt.Printf("Player %s scored %d points\n", p.GetName(), pointsMap[i])
// 	}
// }

func (t *NEATPlayerGroup) Finish(f monopoly.FinishOption, winner int, state monopoly.GameState) {
	pointsMap := map[int]int{
		0: 0,
		1: 0,
		2: 0,
		3: 0,
	}
	if t.gameFinished {
		return
	}
	t.gameFinished = true
	for i, player := range state.Players {
		pointsMap[i] += cfg.ROUND_SCORE * player.RoundsPlayed
		// fmt.Printf("Max properties: %d", player.MaxProperties)
		pointsMap[i] += player.MaxProperties * cfg.POINTS_PER_PROPERTY
	}
	if f == monopoly.WIN {
		pointsMap[winner] = cfg.FIRST_PLACE_SCORE
	}
	for i, p := range t.players {
		p.AddScore(pointsMap[i])
	}
}
