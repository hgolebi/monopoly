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

func (g *Game) getCurrPlayer() *Player {
	return g.players[g.currentPlayerIdx]
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
	player := g.getCurrPlayer()
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
	player := g.getCurrPlayer()
	player.IsJailed = true
	player.CurrentPosition = g.settings.JAIL_POSITION
}

func (g *Game) movePlayer(count int) {
	player := g.getCurrPlayer()
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
	player := g.getCurrPlayer()
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
	player := g.getCurrPlayer()
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
	player := g.getCurrPlayer()
	g.chargePlayer(player, g.settings.JAIL_BAIL, nil)
	if player.IsBankrupt {
		return
	}
	player.IsJailed = false
	player.roundsInJail = 0
	g.makeMove(1, 0, 0)
}

func (g *Game) jailCard() {
	player := g.getCurrPlayer()
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

func (g *Game) resolveStandardAction(action_details ActionDetails) {
	switch action_details.Action {
	case MORTGAGE:
		g.mortgage(action_details.PropertyId)
		return
	case SELLHOUSE:
		g.sellHouse(action_details.PropertyId)
		return
	case BUYHOUSE:
		g.buyHouse(action_details.PropertyId)
		return
	case SELLOFFER:
		g.sendSellOffer(action_details.PlayerId, action_details.PropertyId, action_details.Price)
		return
	case BUYOFFER:
		g.sendBuyOffer(action_details.PlayerId, action_details.PropertyId, action_details.Price)
		return
	case BUYOUT:
		g.buyOut(action_details.PropertyId)
		return
	}

}

func (g *Game) getMortgageList() []int {
	mortgage_list := []int{}
	properties := g.players[g.currentPlayerIdx].Properties
	for _, property := range properties {
		if !property.IsMortgaged && !g.checkHouses(property) {
			mortgage_list = append(mortgage_list, property.PropertyIndex)
		}
	}
	return mortgage_list
}

func (g *Game) getBuyOutList() []int {
	buyout_list := []int{}
	properties := g.players[g.currentPlayerIdx].Properties
	for _, property := range properties {
		if property.IsMortgaged {
			buyout_list = append(buyout_list, property.PropertyIndex)
		}
	}
	return buyout_list
}

func (g *Game) getSellPropertyList() []int {
	sell_list := []int{}
	properties := g.players[g.currentPlayerIdx].Properties
	for _, property := range properties {
		if !g.checkHouses(property) {
			sell_list = append(sell_list, property.PropertyIndex)
		}
	}
	return sell_list
}

func (g *Game) getBuyPropertyList() []int {
	buy_list := []int{}
	player := g.getCurrPlayer()
	for _, property := range g.properties {
		if property.Owner != nil && property.Owner != player && !g.checkHouses(property) {
			buy_list = append(buy_list, property.PropertyIndex)
		}
	}
	return buy_list
}

func (g *Game) getBuyHouseList() []int {
	buy_list := []int{}
	temp_list := []int{}
	for _, set := range g.players[g.currentPlayerIdx].Sets {
		can_build_houses := true
		for _, propertyIdx := range g.sets[set] {
			property := g.properties[propertyIdx]
			if property.IsMortgaged || !property.CanBuildHouse {
				can_build_houses = false
				break
			}
		}
		if can_build_houses {
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
			sell_list = append(sell_list, property.PropertyIndex)
		}
	}
	return sell_list
}

func (g *Game) checkHouses(property *Property) bool {
	if !property.CanBuildHouse {
		return false
	}
	set := g.sets[property.Set]
	for _, propertyIdx := range set {
		if g.properties[propertyIdx].Houses > 0 {
			return true
		}
	}
	return false
}

func (g *Game) chargePlayer(player *Player, amount int, target *Player) {
	if player.Money >= amount {
		player.Charge(amount, target)
		return
	}

	net_worth := g.calculateNetWorth(player)
	if net_worth < amount {
		player.Charge(amount, target)
		return
	}
	for player.Money < amount {
		action_list := FullActionList{}
		action_list.MortgageList = g.getMortgageList()
		action_list.SellHouseList = g.getSellHouseList()
		if len(action_list.MortgageList) > 0 {
			action_list.Actions = append(action_list.Actions, MORTGAGE)
		}
		if len(action_list.SellHouseList) > 0 {
			action_list.Actions = append(action_list.Actions, SELLHOUSE)
		}
		action_details := g.io.GetAction(action_list, g.getState())
		g.resolveStandardAction(action_details)
	}
	player.Charge(amount, target)
}

func (g *Game) mortgage(propertyId int) {
	property := g.properties[propertyId]
	player := g.getCurrPlayer()
	player.AddMoney(property.Price / 2)
	property.IsMortgaged = true
}

func (g *Game) sellHouse(propertyId int) {
	property := g.properties[propertyId]
	player := g.getCurrPlayer()
	player.AddMoney(property.HousePrice / 2)
	property.Houses--
}

func (g *Game) buyHouse(propertyId int) {
	property := g.properties[propertyId]
	player := g.getCurrPlayer()
	g.chargePlayer(player, property.HousePrice, nil)
	property.Houses++
}

func (g *Game) sendSellOffer(id int, param2 int, price int) {
	panic("unimplemented")
}

func (g *Game) sendBuyOffer(id int, param2 int, price int) {
	panic("unimplemented")
}

func (g *Game) buyOut(propertyId int) {
	property := g.properties[propertyId]
	player := g.getCurrPlayer()
	g.chargePlayer(player, int(float64(property.Price)*1.1), nil)
	property.IsMortgaged = false
}

func (g *Game) doForNoActionField() {
	return
}

func (g *Game) doForTaxField(f *TaxField) {
	player := g.getCurrPlayer()
	g.chargePlayer(player, f.Tax, nil)
	g.standardActions()
}

func (g *Game) doForGoToJailField() {
	g.jailPlayer()
}

func (g *Game) doForProperty(p *Property) {
	g.propertyAction(p)
	g.standardActions()
}

func (g *Game) propertyAction(p *Property) {
	player := g.getCurrPlayer()
	if p.Owner == player {
		return
	}

	if p.Owner != nil {
		amount := g.checkCharge(p)
		g.chargePlayer(player, amount, p.Owner)
		return
	}

	if player.Money < p.Price {
		g.auction(p, g.currentPlayerIdx)
		return
	}

	actions := FullActionList{
		Actions: []Action{NOACTION},
	}
	if player.Money >= p.Price {
		actions.Actions = append(actions.Actions, BUY)
	}
	action_details := g.io.GetAction(actions, g.getState())
	if action_details.Action == BUY {
		player.Charge(p.Price, nil)
		player.AddProperty(p)
		return
	}

	g.auction(p, g.currentPlayerIdx)
	return
}
