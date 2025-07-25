package monopoly

import (
	"container/list"
	"fmt"
	"math/rand"
	"slices"
)

const RAILROAD = "Railroad"
const UTILITY = "Utility"

type GameSettings struct {
	MAX_ROUNDS       int
	START_PASS_MONEY int
	JAIL_POSITION    int
	JAIL_BAIL        int
	MAX_HOUSES       int
	MIN_PRICE        int
	MAX_OFFER_TRIES  int
}

type Game struct {
	players          []*Player
	fields           []Field
	properties       []*Property
	charge_map       map[int][]int
	sets             map[string][]int
	currentPlayerIdx int
	round            int
	settings         GameSettings
	io               IMonopoly_IO
	logger           Logger
	buy_offer_tries  int
	sell_offer_tries int
}

func NewGame(players_count int, io IMonopoly_IO, logger Logger) *Game {
	g := &Game{}
	g.io = io
	g.logger = logger
	g.logger.Init()
	g.logger.Log("Initializing game...")
	g.round = 1
	g.currentPlayerIdx = 0

	if players_count < 2 || players_count > 4 {
		panic("Players count must be between 2 and 4")
	}

	g.players = []*Player{
		NewPlayer(0, "player1", 1500),
		NewPlayer(1, "player2", 1500),
		NewPlayer(2, "player3", 1500),
		NewPlayer(3, "player4", 1500),
	}[:players_count]
	g.io.Init(players_count)

	g.properties = []*Property{
		NewProperty(1, 0, "Brown1", 60, 50, true, "Brown"),
		NewProperty(3, 1, "Brown2", 60, 50, true, "Brown"),
		NewProperty(5, 2, "Railroad1", 200, 0, false, RAILROAD),
		NewProperty(6, 3, "LightBlue1", 100, 50, true, "Light Blue"),
		NewProperty(8, 4, "LightBlue2", 100, 50, true, "Light Blue"),
		NewProperty(9, 5, "LightBlue3", 120, 50, true, "Light Blue"),
		NewProperty(11, 6, "Pink1", 140, 100, true, "Pink"),
		NewProperty(12, 7, "Utility1", 150, 0, false, UTILITY),
		NewProperty(13, 8, "Pink2", 140, 100, true, "Pink"),
		NewProperty(14, 9, "Pink3", 160, 100, true, "Pink"),
		NewProperty(15, 10, "Railroad2", 200, 0, false, RAILROAD),
		NewProperty(16, 11, "Orange1", 180, 100, true, "Orange"),
		NewProperty(18, 12, "Orange2", 180, 100, true, "Orange"),
		NewProperty(19, 13, "Orange3", 200, 100, true, "Orange"),
		NewProperty(21, 14, "Red1", 220, 150, true, "Red"),
		NewProperty(23, 15, "Red2", 220, 150, true, "Red"),
		NewProperty(24, 16, "Red3", 240, 150, true, "Red"),
		NewProperty(25, 17, "Railroad3", 200, 0, false, RAILROAD),
		NewProperty(26, 18, "Yellow1", 260, 150, true, "Yellow"),
		NewProperty(27, 19, "Yellow2", 260, 150, true, "Yellow"),
		NewProperty(28, 20, "Utility2", 150, 0, false, UTILITY),
		NewProperty(29, 21, "Yellow3", 280, 150, true, "Yellow"),
		NewProperty(31, 22, "Green1", 300, 200, true, "Green"),
		NewProperty(32, 23, "Green2", 300, 200, true, "Green"),
		NewProperty(34, 24, "Green3", 320, 200, true, "Green"),
		NewProperty(35, 25, "Railroad4", 200, 0, false, RAILROAD),
		NewProperty(37, 26, "DarkBlue1", 350, 200, true, "Dark Blue"),
		NewProperty(39, 27, "DarkBlue2", 400, 200, true, "Dark Blue"),
	}

	g.fields = []Field{
		&NoActionField{FieldIndex: 0, Name: "GO"},
		g.properties[0],
		&Chest{FieldIndex: 2},
		g.properties[1],
		&TaxField{FieldIndex: 4, Name: "Income Tax", Tax: 200},
		g.properties[2],
		g.properties[3],
		&Chance{FieldIndex: 7},
		g.properties[4],
		g.properties[5],
		&NoActionField{FieldIndex: 10, Name: "Jail / Just Visiting"},
		g.properties[6],
		g.properties[7],
		g.properties[8],
		g.properties[9],
		g.properties[10],
		g.properties[11],
		&Chest{FieldIndex: 17},
		g.properties[12],
		g.properties[13],
		&NoActionField{FieldIndex: 20, Name: "Free Parking"},
		g.properties[14],
		&Chance{FieldIndex: 22},
		g.properties[15],
		g.properties[16],
		g.properties[17],
		g.properties[18],
		g.properties[19],
		g.properties[20],
		g.properties[21],
		&GoToJailField{FieldIndex: 30},
		g.properties[22],
		g.properties[23],
		&Chest{FieldIndex: 33},
		g.properties[24],
		g.properties[25],
		&Chance{FieldIndex: 36},
		g.properties[26],
		&TaxField{FieldIndex: 38, Name: "Luxury Tax", Tax: 100},
		g.properties[27],
	}

	g.sets = map[string][]int{
		"Brown":      {0, 1},
		"Light Blue": {3, 4, 5},
		"Pink":       {6, 8, 9},
		"Orange":     {11, 12, 13},
		"Red":        {14, 15, 16},
		"Yellow":     {18, 19, 20, 21},
		"Green":      {22, 23, 24},
		"Dark Blue":  {26, 27},
		RAILROAD:     {2, 10, 17, 25},
		UTILITY:      {7, 20},
	}

	g.charge_map = map[int][]int{
		// Brown
		0: {2, 4, 10, 30, 90, 160, 250},  // Mediterranean Avenue
		1: {4, 8, 20, 60, 180, 320, 450}, // Baltic Avenue

		// Light Blue
		3: {6, 12, 30, 90, 270, 400, 550},  // Oriental Avenue
		4: {6, 12, 30, 90, 270, 400, 550},  // Vermont Avenue
		5: {8, 16, 40, 100, 300, 450, 600}, // Connecticut Avenue

		// Pink
		6: {10, 20, 50, 150, 450, 625, 750}, // St. Charles Place
		8: {10, 20, 50, 150, 450, 625, 750}, // States Avenue
		9: {12, 24, 60, 180, 500, 700, 900}, // Virginia Avenue

		// Orange
		11: {14, 28, 70, 200, 550, 750, 950},  // St. James Place
		12: {14, 28, 70, 200, 550, 750, 950},  // Tennessee Avenue
		13: {16, 32, 80, 220, 600, 800, 1000}, // New York Avenue

		// Red
		14: {18, 36, 90, 250, 700, 875, 1050},  // Kentucky Avenue
		15: {18, 36, 90, 250, 700, 875, 1050},  // Indiana Avenue
		16: {20, 40, 100, 300, 750, 925, 1100}, // Illinois Avenue

		// Yellow
		18: {22, 44, 110, 330, 800, 975, 1150},  // Atlantic Avenue
		19: {22, 44, 110, 330, 800, 975, 1150},  // Ventnor Avenue
		21: {24, 48, 120, 360, 850, 1025, 1200}, // Marvin Gardens

		// Green
		22: {26, 52, 130, 390, 900, 1100, 1275},  // Pacific Avenue
		23: {26, 52, 130, 390, 900, 1100, 1275},  // North Carolina Avenue
		24: {28, 56, 150, 450, 1000, 1200, 1400}, // Pennsylvania Avenue

		// Dark Blue
		26: {35, 70, 175, 500, 1100, 1300, 1500},  // Park Place
		27: {50, 100, 200, 600, 1400, 1700, 2000}, // Boardwalk

		// Railroads (only 4 values: 1 to 4 railroads)
		2:  {25, 50, 100, 200}, // Reading Railroad
		10: {25, 50, 100, 200}, // Pennsylvania Railroad
		17: {25, 50, 100, 200}, // B&O Railroad
		25: {25, 50, 100, 200}, // Short Line

		// Utilities (special: rent = dice × multiplier)
		7:  {4, 10}, // Electric Company: 4× or 10× dice roll
		20: {4, 10}, // Water Works
	}

	g.settings = GameSettings{
		MAX_ROUNDS:       50,
		START_PASS_MONEY: 200,
		JAIL_POSITION:    10,
		JAIL_BAIL:        50,
		MAX_HOUSES:       5,
		MIN_PRICE:        10,
		MAX_OFFER_TRIES:  3,
	}

	g.logger.Log("Game initialized successfully.")
	return g
}

func (g *Game) getState() GameState {
	return GameState{
		Players:          g.players,
		Properties:       g.properties,
		Round:            g.round,
		CurrentPlayerIdx: g.currentPlayerIdx,
	}
}

func (g *Game) getActivePlayers() []int {
	active_players := []int{}
	for idx, player := range g.players {
		if !player.IsBankrupt {
			active_players = append(active_players, idx)
		}
	}
	return active_players
}

func (g *Game) getCurrPlayer() *Player {
	return g.players[g.currentPlayerIdx]
}

func (g *Game) Start() {
	finishFlag := false
	for finishFlag {
		g.logger.Log(fmt.Sprintf("Round %d", g.round))
		for idx, player := range g.players {
			g.logger.LogState(g.getState())
			g.currentPlayerIdx = idx
			finishFlag := g.checkForWinner()
			if finishFlag {
				break
			}
			if player.IsBankrupt {
				continue
			}
			field_name := g.fields[player.CurrentPosition].GetName()
			g.logger.Log(fmt.Sprintf("Player %s's turn. Current position: %s", player.Name, field_name))
			if player.IsJailed {
				g.handleJail()
				continue
			}
			g.makeMove(1, 0, 0)
		}
		g.round++
	}
}

func (g *Game) checkForWinner() bool {
	active_players := g.getActivePlayers()
	if len(active_players) == 0 {
		g.endDraw()
		return true
	} else if len(active_players) == 1 {
		g.endWinner(g.players[active_players[0]])
		return true
	} else if g.round > g.settings.MAX_ROUNDS {
		g.endRoundLimit()
		return true
	}
	return false
}

func (g *Game) endRoundLimit() {
	g.logger.Log("Game ended due to round limit reached.")
	winner_id := -1
	max_net_worth := -1
	for _, player_id := range g.getActivePlayers() {
		net_worth := g.calculateNetWorth(g.players[player_id])
		if net_worth > max_net_worth {
			max_net_worth = net_worth
			winner_id = player_id
		}
	}
	winner := g.players[winner_id]
	g.logger.Log(fmt.Sprintf("Winner is %s with net worth of %d", winner.Name, max_net_worth))
	g.io.Finish(ROUND_LIMIT, winner_id, g.getState())
}

func (g *Game) endWinner(winner *Player) {
	g.logger.Log(fmt.Sprintf("Game ended. Winner is %s", winner.Name))
	g.io.Finish(WIN, winner.ID, g.getState())
}

func (g *Game) endDraw() {
	g.logger.Log("Game ended in a draw. No players left.")
	g.io.Finish(DRAW, -1, g.getState())
}

func (g *Game) makeMove(moves_in_a_row int, d1 int, d2 int) {
	if d1 == 0 {
		if d2 != 0 {
			panic("d2 should be 0 if d1 is 0")
		}
		d1, d2 = g.rollDice()
	}
	if moves_in_a_row >= 3 && d1 == d2 {
		g.logger.Log("Player rolled doubles 3 times in a row, going to jail")
		g.jailPlayer()
		return
	}
	g.movePlayer(d1 + d2)
	g.takeAction()
	player := g.getCurrPlayer()
	if player.IsJailed {
		return
	}
	g.standardActions()
	if player.IsBankrupt {
		return
	}
	if d1 == d2 {
		g.logger.Log("Player rolled doubles, taking another turn")
		g.makeMove(moves_in_a_row+1, 0, 0)
	}
}

func (g *Game) takeAction() {
	player := g.getCurrPlayer()
	field := g.fields[player.CurrentPosition]
	field.Action(g)
	g.logger.LogState(g.getState())
}

func (g *Game) jailPlayer() {
	player := g.getCurrPlayer()
	player.IsJailed = true
	player.CurrentPosition = g.settings.JAIL_POSITION
	player.roundsInJail = 0
	g.logger.Log("Player is going to jail.")
}

func (g *Game) movePlayer(count int) {
	player := g.getCurrPlayer()
	curr_pos := player.CurrentPosition
	new_pos := curr_pos + count
	for new_pos > len(g.fields)-1 {
		player.AddMoney(g.settings.START_PASS_MONEY)
		g.logger.Log(fmt.Sprintf("Player passed GO and received %d money", g.settings.START_PASS_MONEY))
		new_pos = new_pos - len(g.fields)
	}
	player.SetPosition(new_pos)
	field_name := g.fields[new_pos].GetName()
	g.logger.Log(fmt.Sprintf("Player %s moved to %s", player.Name, field_name))
}

func (g *Game) rollDice() (dice1 int, dice2 int) {
	d1, d2 := rand.Intn(6)+1, rand.Intn(6)+1
	g.logger.Log(fmt.Sprintf("Rolled dice: %d, %d", d1, d2))
	return d1, d2
}

func (g *Game) handleJail() {
	g.logger.Log("Player is jailed.")
	player := g.getCurrPlayer()
	g.standardActions()
	if player.IsBankrupt {
		return
	}
	var action_list []JailAction
	action_list = append(action_list, BAIL)
	if player.JailCards > 0 {
		action_list = append(action_list, CARD)
	}
	if player.roundsInJail < 3 {
		action_list = append(action_list, ROLL_DICE)
	}
	action_details := g.io.GetJailAction(g.currentPlayerIdx, g.getState(), action_list)
	switch action_details {
	case ROLL_DICE:
		g.logger.Log("Player chose to roll dice to get out of jail.")
		g.jailRollDice()
		return
	case BAIL:
		g.logger.Log("Player chose to pay bail to get out of jail.")
		g.jailBail()
		return
	case CARD:
		g.logger.Log("Player chose to use a jail card to get out of jail.")
		g.jailCard()
		return
	default:
		panic("unknown action: " + fmt.Sprint(action_details))
	}

}

func (g *Game) jailRollDice() {
	player := g.getCurrPlayer()
	d1, d2 := g.rollDice()
	if d1 == d2 {
		g.logger.Log("Player rolled doubles, getting out of jail.")
		player.IsJailed = false
		player.roundsInJail = 0
		g.makeMove(1, d1, d2)
	} else {
		player.roundsInJail++
	}
}

func (g *Game) jailBail() {
	player := g.getCurrPlayer()
	g.chargePlayer(g.currentPlayerIdx, g.settings.JAIL_BAIL, nil)
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
	g.logger.Log(fmt.Sprintf("Jail cards left: %d", player.JailCards))
	player.IsJailed = false
	player.roundsInJail = 0
	g.makeMove(1, 0, 0)
}

func (g *Game) standardActions() {
	action_list := FullActionList{}

	action_list.MortgageList = g.getMortgageList(g.currentPlayerIdx)
	action_list.BuyOutList = g.getBuyOutList(g.currentPlayerIdx)
	action_list.SellPropertyList = g.getSellPropertyList(g.currentPlayerIdx)
	action_list.BuyPropertyList = g.getBuyPropertyList(g.currentPlayerIdx)
	action_list.BuyHouseList = g.getBuyHouseList(g.currentPlayerIdx)
	action_list.SellHouseList = g.getSellHouseList(g.currentPlayerIdx)

	action_list.Actions = []StdAction{NOACTION}
	if len(action_list.MortgageList) > 0 {
		action_list.Actions = append(action_list.Actions, MORTGAGE)
	}
	if len(action_list.BuyOutList) > 0 {
		action_list.Actions = append(action_list.Actions, BUYOUT)
	}
	if len(action_list.SellPropertyList) > 0 && g.sell_offer_tries < g.settings.MAX_OFFER_TRIES {
		action_list.Actions = append(action_list.Actions, SELLOFFER)
	}
	if len(action_list.BuyPropertyList) > 0 && g.buy_offer_tries < g.settings.MAX_OFFER_TRIES {
		action_list.Actions = append(action_list.Actions, BUYOFFER)
	}
	if len(action_list.BuyHouseList) > 0 {
		action_list.Actions = append(action_list.Actions, BUYHOUSE)
	}
	if len(action_list.SellHouseList) > 0 {
		action_list.Actions = append(action_list.Actions, SELLHOUSE)
	}
	action_details := g.io.GetStdAction(g.currentPlayerIdx, g.getState(), action_list)
	if action_details.Action == NOACTION {
		return
	}
	g.resolveStandardAction(g.currentPlayerIdx, action_details, action_list)
	if g.players[g.currentPlayerIdx].IsBankrupt {
		return
	}
	g.logger.LogState(g.getState())
	g.standardActions()
}

func (g *Game) resolveStandardAction(player_id int, action_details ActionDetails, available FullActionList) {
	player := g.players[player_id]
	if !slices.Contains(available.Actions, action_details.Action) {
		g.logger.Log(fmt.Sprintf("Action %s not available, going bankrupt", StdActionNames[action_details.Action]))
		g.bankrupt(player, nil)
	}
	switch action_details.Action {
	case MORTGAGE:
		if !slices.Contains(available.MortgageList, action_details.PropertyId) {
			g.logger.Log(fmt.Sprintf("Action %s not available for property %d, going bankrupt", StdActionNames[action_details.Action], action_details.PropertyId))
			g.bankrupt(player, nil)
			return
		}
		g.mortgage(player_id, action_details.PropertyId)
		return
	case SELLHOUSE:
		if !slices.Contains(available.SellHouseList, action_details.PropertyId) {
			g.logger.Log(fmt.Sprintf("Action %s not available for property %d, going bankrupt", StdActionNames[action_details.Action], action_details.PropertyId))
			g.bankrupt(player, nil)
			return
		}
		g.sellHouse(player_id, action_details.PropertyId)
		return
	case BUYHOUSE:
		if !slices.Contains(available.BuyHouseList, action_details.PropertyId) {
			g.logger.Log(fmt.Sprintf("Action %s not available for property %d, going bankrupt", StdActionNames[action_details.Action], action_details.PropertyId))
			g.bankrupt(player, nil)
			return
		}
		g.buyHouse(player_id, action_details.PropertyId)
		return
	case SELLOFFER:
		if g.sell_offer_tries >= g.settings.MAX_OFFER_TRIES {
			g.logger.Log("Max sell offer tries reached, going bankrupt.")
			g.bankrupt(player, nil)
			return
		}
		if !slices.Contains(available.SellPropertyList, action_details.PropertyId) {
			g.logger.Log(fmt.Sprintf("Action %s not available for property %d, going bankrupt", StdActionNames[action_details.Action], action_details.PropertyId))
			g.bankrupt(player, nil)
			return
		}
		if action_details.PlayerId == player_id || g.players[action_details.PlayerId].IsBankrupt {
			g.logger.Log(fmt.Sprintf("Invalid player %d for sell offer, going bankrupt", action_details.PlayerId))
			g.bankrupt(player, nil)
			return
		}
		if action_details.Price < 0 {
			g.logger.Log(fmt.Sprintf("Invalid price %d for sell offer, going bankrupt", action_details.Price))
			g.bankrupt(player, nil)
			return
		}
		if action_details.Price > g.players[action_details.PlayerId].Money {
			g.logger.Log(fmt.Sprintf("Player %d cannot afford to buy property for %d", action_details.PlayerId, action_details.Price))
			g.sell_offer_tries++
			return
		}
		g.sendSellOffer(player_id, action_details.PlayerId, action_details.PropertyId, action_details.Price)
		return
	case BUYOFFER:
		if g.buy_offer_tries >= g.settings.MAX_OFFER_TRIES {
			g.logger.Log("Max buy offer tries reached, going bankrupt.")
			g.bankrupt(player, nil)
			return
		}
		if !slices.Contains(available.BuyPropertyList, action_details.PropertyId) {
			g.logger.Log(fmt.Sprintf("Action %s not available for property %d, going bankrupt", StdActionNames[action_details.Action], action_details.PropertyId))
			g.bankrupt(player, nil)
			return
		}
		if action_details.Price < 0 {
			g.logger.Log(fmt.Sprintf("Invalid price %d for buy offer, going bankrupt", action_details.Price))
			g.bankrupt(player, nil)
			return
		}
		g.sendBuyOffer(player_id, action_details.PropertyId, action_details.Price)
		return
	case BUYOUT:
		if !slices.Contains(available.BuyOutList, action_details.PropertyId) {
			g.logger.Log(fmt.Sprintf("Action %s not available for property %d, going bankrupt", StdActionNames[action_details.Action], action_details.PropertyId))
			g.bankrupt(player, nil)
			return
		}
		g.buyOut(player_id, action_details.PropertyId)
		return
	}

}

func (g *Game) getMortgageList(player_id int) []int {
	mortgage_list := []int{}
	properties := g.players[player_id].Properties
	for _, id := range properties {
		property := g.properties[id]
		if !property.IsMortgaged && !g.checkHouses(property) {
			mortgage_list = append(mortgage_list, property.PropertyIndex)
		}
	}
	return mortgage_list
}

func (g *Game) getBuyOutList(player_id int) []int {
	buyout_list := []int{}
	properties := g.players[player_id].Properties
	for _, id := range properties {
		property := g.properties[id]
		if property.IsMortgaged {
			buyout_list = append(buyout_list, property.PropertyIndex)
		}
	}
	return buyout_list
}

func (g *Game) getSellPropertyList(player_id int) []int {
	sell_list := []int{}
	properties := g.players[player_id].Properties
	for _, id := range properties {
		property := g.properties[id]
		if !g.checkHouses(property) {
			sell_list = append(sell_list, property.PropertyIndex)
		}
	}
	return sell_list
}

func (g *Game) getBuyPropertyList(player_id int) []int {
	buy_list := []int{}
	player := g.players[player_id]
	for _, property := range g.properties {
		if property.Owner != nil && property.Owner != player && !g.checkHouses(property) {
			buy_list = append(buy_list, property.PropertyIndex)
		}
	}
	return buy_list
}

func (g *Game) getBuyHouseList(player_id int) []int {
	buy_list := []int{}
	player := g.players[player_id]
	for _, set := range g.sets {
		has_full_set := true
		var temp_list []int
		for _, propertyIdx := range set {
			property := g.properties[propertyIdx]
			if !property.CanBuildHouse {
				has_full_set = false
				break
			}
			if property.Owner != player {
				has_full_set = false
				break
			}
			if property.Houses < g.settings.MAX_HOUSES {
				temp_list = append(temp_list, propertyIdx)
			}
		}
		if has_full_set {
			buy_list = append(buy_list, temp_list...)
		}
	}
	return buy_list
}

func (g *Game) getSellHouseList(player_id int) []int {
	sell_list := []int{}
	properties := g.players[player_id].Properties
	for _, id := range properties {
		property := g.properties[id]
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

func (g *Game) calculateNetWorth(player *Player) int {
	net_worth := player.Money
	for _, id := range player.Properties {
		property := g.properties[id]
		if property.IsMortgaged {
			continue
		}
		net_worth += property.Price / 2
		net_worth += property.Houses * (property.HousePrice / 2)
	}
	return net_worth
}

func (g *Game) chargePlayer(player_id int, amount int, target *Player) {
	player := g.players[player_id]
	target_name := "Bank"
	if target != nil {
		target_name = target.Name
	}
	if player.Money >= amount {
		g.logger.Log(fmt.Sprintf("%s charged %d money from player %s", target_name, amount, player.Name))
		g.charge(player, amount, target)
		return
	}
	g.logger.Log(fmt.Sprintf("Player %s does not have enough money to pay %d.", player.Name, amount))

	net_worth := g.calculateNetWorth(player)
	if net_worth < amount {
		g.charge(player, amount, target)
		g.logger.Log(fmt.Sprintf("Player %s is bankrupt. All his properties go to %s", player.Name, target_name))
		return
	}
	for player.Money < amount {
		action_list := FullActionList{}
		action_list.MortgageList = g.getMortgageList(player_id)
		action_list.SellHouseList = g.getSellHouseList(player_id)
		if len(action_list.MortgageList) > 0 {
			action_list.Actions = append(action_list.Actions, MORTGAGE)
		}
		if len(action_list.SellHouseList) > 0 {
			action_list.Actions = append(action_list.Actions, SELLHOUSE)
		}
		state := g.getState()
		state.Charge = amount
		action_details := g.io.GetStdAction(player_id, state, action_list)
		g.resolveStandardAction(player_id, action_details, action_list)
	}
	g.charge(player, amount, target)
}

func (g *Game) mortgage(player_id int, propertyId int) {
	property := g.properties[propertyId]
	player := g.players[player_id]
	player.AddMoney(property.Price / 2)
	property.IsMortgaged = true
	g.logger.Log(fmt.Sprintf("Player %s mortgaged property %s for %d money", player.Name, property.Name, property.Price/2))
}

func (g *Game) sellHouse(player_id int, propertyId int) {
	property := g.properties[propertyId]
	player := g.players[player_id]
	player.AddMoney(property.HousePrice / 2)
	property.Houses--
	g.logger.Log(fmt.Sprintf("Player %s sold a house on property %s for %d money", player.Name, property.Name, property.HousePrice/2))
}

func (g *Game) buyHouse(player_id int, propertyId int) {
	property := g.properties[propertyId]
	player := g.players[player_id]
	g.chargePlayer(player_id, property.HousePrice, nil)
	property.Houses++
	g.logger.Log(fmt.Sprintf("Player %s bought a house on property %s for %d money", player.Name, property.Name, property.HousePrice))
}

func (g *Game) sendSellOffer(player_id int, target_id int, property_id int, price int) {
	seller := g.players[player_id]
	buyer := g.players[target_id]
	property := g.properties[property_id]
	g.logger.Log(fmt.Sprintf("Player %s sent a sell offer to %s for property %s with price %d", seller.Name, buyer.Name, property.Name, price))
	accepted := g.io.BuyFromPlayerDecision(target_id, g.getState(), property_id, price)
	if accepted {
		g.sell_offer_tries = 0
		if buyer.Money < price {
			g.logger.Log(fmt.Sprintf("Player %s does not have enough money to buy property %s for %d. Going bankrupt.", buyer.Name, property.Name, price))
			g.bankrupt(buyer, nil)
			return
		}
		g.charge(buyer, price, seller)
		g.transferProperty(seller, buyer, property_id)
	}

}

func (g *Game) sendBuyOffer(player_id int, property_id int, price int) {
	buyer := g.players[player_id]
	property := g.properties[property_id]
	seller := property.Owner
	if seller == nil {
		g.logger.Log(fmt.Sprintf("Property %s is not owned by anyone, going bankrupt", property.Name))
		g.bankrupt(buyer, nil)
		g.buy_offer_tries = 0
		return
	}
	g.logger.Log(fmt.Sprintf("Player %s sent a buy offer to %s for property %s with price %d", seller.Name, buyer.Name, property.Name, price))
	accepted := g.io.SellToPlayerDecision(seller.ID, g.getState(), property_id, price)
	if accepted {
		g.buy_offer_tries = 0
		if buyer.Money < price {
			g.logger.Log(fmt.Sprintf("Player %s does not have enough money to buy property %s for %d. Going bankrupt.", buyer.Name, property.Name, price))
			g.bankrupt(buyer, nil)
			return
		}
		g.charge(buyer, price, seller)
		g.transferProperty(seller, buyer, property_id)
	}
}

func (g *Game) buyOut(player_id int, propertyId int) {
	property := g.properties[propertyId]
	player := g.players[player_id]
	g.chargePlayer(player_id, int(float64(property.Price)*1.1), nil)
	property.IsMortgaged = false
	g.logger.Log(fmt.Sprintf("Player %s bought out property %s for %d money", player.Name, property.Name, int(float64(property.Price)*1.1)))
}

func (g *Game) doForNoActionField() {}

func (g *Game) doForChest() {
	action := rand.Intn(7)
	g.resolveChanceOrChest(action)
}

func (g *Game) doForChance() {
	action := rand.Intn(8)
	g.resolveChanceOrChest(action)
}

func (g *Game) resolveChanceOrChest(action int) {
	player := g.getCurrPlayer()

	switch action {
	case 0: // Player receives money from the bank
		amount := rand.Intn(151) + 50 // 50-200
		g.logger.Log(fmt.Sprintf("Chest: Player %s receives %d from the bank.", player.Name, amount))
		player.AddMoney(amount)
	case 1: // Player pays money to the bank
		amount := rand.Intn(101) + 50 // 50-150
		g.logger.Log(fmt.Sprintf("Chest: Player %s pays %d to the bank.", player.Name, amount))
		g.chargePlayer(g.currentPlayerIdx, amount, nil)
	case 2: // Each player pays money to the current player
		amount := rand.Intn(11) + 10 // 10-20
		g.logger.Log(fmt.Sprintf("Chest: Each player pays %d to %s.", amount, player.Name))
		for idx, p := range g.players {
			if idx != g.currentPlayerIdx && !p.IsBankrupt {
				g.chargePlayer(idx, amount, g.players[g.currentPlayerIdx])
			}
		}
	case 3: // Current player pays money to each other player
		amount := rand.Intn(11) + 10 // 10-20
		g.logger.Log(fmt.Sprintf("Chest: %s pays %d to each other player.", player.Name, amount))
		for idx, p := range g.players {
			if idx == g.currentPlayerIdx && !p.IsBankrupt {
				g.chargePlayer(g.currentPlayerIdx, amount, p)
			}
		}
	case 4: // Player goes directly to jail
		g.logger.Log(fmt.Sprintf("Chest: Player %s goes directly to jail.", player.Name))
		g.jailPlayer()
	case 5: // Player receives a Get Out of Jail Free card
		player.JailCards++
		g.logger.Log(fmt.Sprintf("Chest: Player %s receives a Get Out of Jail Free card.", player.Name))
	case 6: // Player moves to GO
		g.logger.Log(fmt.Sprintf("Chest: Player %s moves to GO and receives %d money.", player.Name, g.settings.START_PASS_MONEY))
		player.SetPosition(0)
		player.AddMoney(g.settings.START_PASS_MONEY)
	case 7: // Player moves to a specific field
		field_index := rand.Intn(len(g.fields))
		field := g.fields[field_index]
		g.logger.Log(fmt.Sprintf("Chest: Player %s moves to field %s.", player.Name, field.GetName()))
		player.SetPosition(field_index)
		field.Action(g)
	}
}

func (g *Game) doForTaxField(f *TaxField) {
	g.chargePlayer(g.currentPlayerIdx, f.Tax, nil)
}

func (g *Game) doForGoToJailField() {
	g.jailPlayer()
}

func (g *Game) doForProperty(p *Property) {
	player := g.getCurrPlayer()
	if p.Owner == player {
		g.logger.Log("This property is already owned by the player.")
		return
	}

	if p.Owner != nil {
		amount := g.checkCharge(p)
		g.logger.Log(fmt.Sprintf("This property is owned by %s. Charge for tresspassing: %d$ ", p.Owner.Name, amount))
		g.chargePlayer(g.currentPlayerIdx, amount, p.Owner)
		return
	}

	g.logger.Log("This property is not owned by anyone.")
	if player.Money < p.Price {
		g.auction(p, g.currentPlayerIdx)
		return
	}
	wantToBuy := g.io.BuyDecision(g.currentPlayerIdx, g.getState(), p.PropertyIndex)
	if !wantToBuy {
		g.auction(p, g.currentPlayerIdx)
		return
	}
	g.logger.Log(fmt.Sprintf("Player %s bought property %s for %d money", player.Name, p.Name, p.Price))
	g.charge(player, p.Price, nil)
	g.addProperty(player, p.PropertyIndex)

}

func (g *Game) checkCharge(p *Property) int {
	if p.IsMortgaged {
		return 0
	}
	charges := g.charge_map[p.PropertyIndex]
	charge_idx := -1
	if p.Set == RAILROAD {
		for _, propertyIdx := range g.sets[RAILROAD] {
			if g.properties[propertyIdx].Owner == p.Owner {
				charge_idx++
			}
		}
		return charges[charge_idx]
	}
	if p.Set == UTILITY {
		for _, propertyIdx := range g.sets[UTILITY] {
			if g.properties[propertyIdx].Owner == p.Owner {
				charge_idx++
			}
		}
		return charges[charge_idx]
	}

	set := g.sets[p.Set]
	has_full_set := true
	for _, propertyIdx := range set {
		if g.properties[propertyIdx].Owner != p.Owner {
			has_full_set = false
			break
		}
	}
	if !has_full_set {
		return charges[0]
	}
	charge_idx = 1
	charge_idx += p.Houses
	return charges[charge_idx]
}

func (g *Game) auction(property *Property, first_player_id int) {
	g.logger.Log(fmt.Sprintf("Auctioning property %s", property.Name))
	queue := list.New()
	for _, player := range g.players[first_player_id:] {
		if !player.IsBankrupt {
			queue.PushBack(player.ID)
		}
	}
	for _, player := range g.players[:first_player_id] {
		if !player.IsBankrupt {
			queue.PushBack(player.ID)
		}
	}
	curr_price := g.settings.MIN_PRICE
	auction_winner := -1
	for queue.Len() > 0 {
		bidderID := queue.Front().Value.(int)
		queue.Remove(queue.Front())
		if auction_winner == bidderID {
			break
		}
		bidder := g.players[bidderID]
		bid_offer := g.io.BiddingDecision(bidderID, g.getState(), property.PropertyIndex, curr_price)
		if bid_offer <= curr_price {
			g.logger.Log(fmt.Sprintf("Player %s did not bid.", bidder.Name))
		} else if bid_offer > bidder.Money {
			g.logger.Log(fmt.Sprintf("Player %s does not have enough money to bid %d. Going bankrupt.", bidder.Name, bid_offer))
			g.bankrupt(bidder, nil)
		} else {
			g.logger.Log(fmt.Sprintf("Player %s bid %d", bidder.Name, bid_offer))
			curr_price = bid_offer
			auction_winner = bidderID
			queue.PushBack(bidderID)
		}
	}
	if auction_winner == -1 {
		g.logger.Log("Auction ended without any bids.")
		return
	}
	winner := g.players[auction_winner]
	g.logger.Log(fmt.Sprintf("Player %s won the auction with a bid of %d", winner.Name, curr_price))
	g.charge(winner, curr_price, nil)
	g.addProperty(winner, property.PropertyIndex)
}

func (g *Game) addProperty(player *Player, property_id int) {
	if g.properties[property_id].Owner != nil {
		panic("Property is already owned")
	}
	property := g.properties[property_id]
	property.Owner = player
	player.AddProperty(property_id)
}

func (g *Game) charge(player *Player, amount int, target *Player) {
	if player.Money < amount {
		g.bankrupt(player, target)
		return
	}
	player.RemoveMoney(amount)
	if target != nil {
		target.AddMoney(amount)
	}
}

func (g *Game) bankrupt(player *Player, creditor *Player) {
	player.IsBankrupt = true
	if creditor != nil {
		creditor.AddMoney(max(0, player.Money))
		for _, property := range player.Properties {
			g.transferProperty(player, creditor, property)
		}
	} else {
		for _, property := range player.Properties {
			g.properties[property].Owner = nil
			g.properties[property].IsMortgaged = false
			g.properties[property].Houses = 0
			active_players := g.getActivePlayers()
			g.auction(g.properties[property], active_players[rand.Intn(len(active_players))])
		}
	}
	player.Properties = []int{}
	player.CurrentPosition = -1
	player.Money = -1
}

func (g *Game) transferProperty(player *Player, target *Player, property_id int) {
	property := g.properties[property_id]
	if property.Owner != player {
		panic("Property does not belong to player")
	}
	property.Owner = target
	player.RemoveProperty(property_id)
	target.AddProperty(property_id)
}
