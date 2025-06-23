package monopoly

import (
	"fmt"
	"math/rand"
)

type GameSettings struct {
	MAX_ROUNDS       int
	START_PASS_MONEY int
	JAIL_POSITION    int
	JAIL_BAIL        int
	MAX_HOUSES       int
}

type Game struct {
	players          []*Player
	fields           []Field
	properties       []*Property
	sets             map[string][]int
	currentPlayerIdx int
	round            int
	settings         GameSettings
	io               IMonopoly_IO
}

func (g *Game) initGame() {
	//TODO
}

func (g *Game) getState() GameState {
	return GameState{
		Players:          g.players,
		Fields:           g.fields,
		Round:            g.round,
		CurrentPlayerIdx: g.currentPlayerIdx,
	}
}

func (g *Game) Start() {
	g.initGame()
	for {
		for idx, player := range g.players {
			g.currentPlayerIdx = idx
			g.checkForWinner()
			if player.IsBankrupt {
				continue
			}
			if player.IsJailed {
				g.handleJail()
				continue
			}
			g.makeMove(1, 0, 0)
		}
		g.round++
	}
}

func (g *Game) checkForWinner() {
	players_alive := 0
	for _, player := range g.players {
		if !player.IsBankrupt {
			players_alive++
		}
	}
	if players_alive == 0 {
		g.endDraw()
	} else if players_alive == 1 {
		g.endWinner()
	} else if g.round > g.settings.MAX_ROUNDS {
		g.endRoundLimit()
	}
}

func (g *Game) endRoundLimit() {
	panic("unimplemented")
}

func (g *Game) endWinner() {
	panic("unimplemented")
}

func (g *Game) endDraw() {
	panic("unimplemented")
}

func (g *Game) makeMove(moves_in_a_row int, d1 int, d2 int) {
	if d1 == 0 {
		if d2 != 0 {
			panic("d2 should be 0 if d1 is 0")
		}
		d1, d2 = g.rollDice()
	}
	if moves_in_a_row >= 3 && d1 == d2 {
		g.jailPlayer()
		return
	}
	g.movePlayer(d1 + d2)
	g.takeAction()
	player := g.players[g.currentPlayerIdx]
	if player.IsJailed {
		return
	}
	if d1 == d2 {
		g.makeMove(moves_in_a_row+1, 0, 0)
	}
}

func (g *Game) takeAction() {
	panic("unimplemented")
}

func (g *Game) jailPlayer() {
	player := g.players[g.currentPlayerIdx]
	player.IsJailed = true
	player.CurrentPosition = g.settings.JAIL_POSITION
}

func (g *Game) movePlayer(count int) {
	player := g.players[g.currentPlayerIdx]
	curr_pos := player.CurrentPosition
	new_pos := curr_pos + count
	for new_pos > len(g.fields)-1 {
		player.AddMoney(g.settings.START_PASS_MONEY)
		new_pos = new_pos - len(g.fields)
	}
	player.SetPosition(new_pos)
}

func (g *Game) rollDice() (dice1 int, dice2 int) {
	return rand.Intn(6) + 1, rand.Intn(6) + 1
}

func (g *Game) handleJail() {
	player := g.players[g.currentPlayerIdx]
	g.standardActions()
	var action_list = FullActionList{
		Actions: []Action{},
	}
	action_list.Actions = append(action_list.Actions, JAIL_BAIL)
	if player.JailCards > 0 {
		action_list.Actions = append(action_list.Actions, JAIL_CARD)
	}
	if player.roundsInJail < 3 {
		action_list.Actions = append(action_list.Actions, JAIL_ROLL_DICE)
	}
	action_details := g.io.GetAction(action_list, g.getState())
	switch action_details.Action {
	case JAIL_ROLL_DICE:
		g.jailRollDice()
		return
	case JAIL_BAIL:
		g.jailBail()
		return
	case JAIL_CARD:
		g.jailCard()
		return
	default:
		panic("unknown action: " + fmt.Sprint(action_details.Action))
	}

}

func (g *Game) jailRollDice() {
	player := g.players[g.currentPlayerIdx]
	d1, d2 := g.rollDice()
	if d1 == d2 {
		player.IsJailed = false
		player.roundsInJail = 0
		g.makeMove(1, d1, d2)
	} else {
		player.roundsInJail++
	}
}

func (g *Game) jailBail() {
	player := g.players[g.currentPlayerIdx]
	player.Charge(g.settings.JAIL_BAIL)
	if player.IsBankrupt {
		return
	}
	player.IsJailed = false
	player.roundsInJail = 0
	g.makeMove(1, 0, 0)
}

func (g *Game) jailCard() {
	player := g.players[g.currentPlayerIdx]
	if player.JailCards <= 0 {
		panic("no jail cards left")
	}
	player.JailCards--
	player.IsJailed = false
	player.roundsInJail = 0
	g.makeMove(1, 0, 0)
}

func (g *Game) standardActions() {
	action_list := FullActionList{}

	action_list.MortgageList = g.getMortgageList()
	action_list.BuyOutList = g.getBuyOutList()
	action_list.SellPropertyList = g.getSellPropertyList()
	action_list.BuyPropertyList = g.getBuyPropertyList()
	action_list.BuyHouseList = g.getBuyHouseList()
	action_list.SellHouseList = g.getSellHouseList()

	action_list.Actions = []Action{NOACTION}
	if len(action_list.MortgageList) > 0 {
		action_list.Actions = append(action_list.Actions, MORTGAGE)
	}
	if len(action_list.BuyOutList) > 0 {
		action_list.Actions = append(action_list.Actions, BUYOUT)
	}
	if len(action_list.SellPropertyList) > 0 {
		action_list.Actions = append(action_list.Actions, SELLOFFER)
	}
	if len(action_list.BuyPropertyList) > 0 {
		action_list.Actions = append(action_list.Actions, BUYOFFER)
	}
	if len(action_list.BuyHouseList) > 0 {
		action_list.Actions = append(action_list.Actions, BUYHOUSE)
	}
	if len(action_list.SellHouseList) > 0 {
		action_list.Actions = append(action_list.Actions, SELLHOUSE)
	}
	action_details := g.io.GetAction(action_list, g.getState())
	if action_details.Action == NOACTION {
		return
	}
	g.resolveStandardAction(action_details)
	g.standardActions()
}

func (g *Game) getMortgageList() []int {
	mortgage_list := []int{}
	properties := g.players[g.currentPlayerIdx].Properties
	for _, property := range properties {
		if !property.IsMortgaged && !g.checkHouses(property) {
			mortgage_list = append(mortgage_list, property.Index)
		}
	}
	return mortgage_list
}

func (g *Game) getBuyOutList() []int {
	buyout_list := []int{}
	properties := g.players[g.currentPlayerIdx].Properties
	for _, property := range properties {
		if property.IsMortgaged {
			buyout_list = append(buyout_list, property.Index)
		}
	}
	return buyout_list
}

func (g *Game) getSellPropertyList() []int {
	sell_list := []int{}
	properties := g.players[g.currentPlayerIdx].Properties
	for _, property := range properties {
		if !g.checkHouses(property) {
			sell_list = append(sell_list, property.Index)
		}
	}
	return sell_list
}

func (g *Game) getBuyPropertyList() []int {
	buy_list := []int{}
	for _, property := range g.properties {
		if property.Owner >= 0 && property.Owner != g.currentPlayerIdx && !g.checkHouses(property) {
			buy_list = append(buy_list, property.Index)
		}
	}
	return buy_list
}

func (g *Game) getBuyHouseList() []int {
	buy_list := []int{}
	temp_list := []int{}
	for _, set := range g.players[g.currentPlayerIdx].Sets {
		foundMortgaged := false
		for _, propertyIdx := range g.sets[set] {
			if g.properties[propertyIdx].IsMortgaged {
				foundMortgaged = true
				break
			}
		}
		if !foundMortgaged {
			temp_list = append(temp_list, g.sets[set]...)
		}
	}
	for _, propertyIdx := range temp_list {
		if g.properties[propertyIdx].Houses < g.settings.MAX_HOUSES {
			buy_list = append(buy_list, propertyIdx)
		}
	}
	return buy_list
}

func (g *Game) getSellHouseList() []int {
	sell_list := []int{}
	properties := g.players[g.currentPlayerIdx].Properties
	for _, property := range properties {
		if property.Houses > 0 {
			sell_list = append(sell_list, property.Index)
		}
	}
	return sell_list
}

func (g *Game) checkHouses(property *Property) bool {
	set := g.sets[property.Set]
	for _, propertyIdx := range set {
		if g.properties[propertyIdx].Houses > 0 {
			return true
		}
	}
	return false
}
