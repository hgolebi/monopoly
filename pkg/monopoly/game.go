package monopoly

import (
	"fmt"
	"math/rand"
	"os"
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
}

func (g *Game) initGame() {
	g.io = &ConsoleCLI{}
	g.logger = &ConsoleLogger{}
	g.logger.Log("Initializing game...")
	g.round = 1
	g.currentPlayerIdx = 0

	g.players = []*Player{
		NewPlayer("player1", 1500),
		NewPlayer("player2", 1500),
	}

	g.properties = []*Property{
		NewProperty(1, 0, "Mediterranean Avenue", 60, 50, true, "Brown"),
		NewProperty(3, 1, "Baltic Avenue", 60, 50, true, "Brown"),
		NewProperty(5, 2, "Reading Railroad", 200, 0, false, RAILROAD),
		NewProperty(6, 3, "Oriental Avenue", 100, 50, true, "Light Blue"),
		NewProperty(8, 4, "Vermont Avenue", 100, 50, true, "Light Blue"),
		NewProperty(9, 5, "Connecticut Avenue", 120, 50, true, "Light Blue"),
		NewProperty(11, 6, "St. Charles Place", 140, 100, true, "Pink"),
		NewProperty(12, 7, "Electric Company", 150, 0, false, UTILITY),
		NewProperty(13, 8, "States Avenue", 140, 100, true, "Pink"),
		NewProperty(14, 9, "Virginia Avenue", 160, 100, true, "Pink"),
		NewProperty(15, 10, "Pennsylvania Railroad", 200, 0, false, RAILROAD),
		NewProperty(16, 11, "St. James Place", 180, 100, true, "Orange"),
		NewProperty(18, 12, "Tennessee Avenue", 180, 100, true, "Orange"),
		NewProperty(19, 13, "New York Avenue", 200, 100, true, "Orange"),
		NewProperty(21, 14, "Kentucky Avenue", 220, 150, true, "Red"),
		NewProperty(23, 15, "Indiana Avenue", 220, 150, true, "Red"),
		NewProperty(24, 16, "Illinois Avenue", 240, 150, true, "Red"),
		NewProperty(25, 17, "B&O Railroad", 200, 0, false, RAILROAD),
		NewProperty(26, 18, "Atlantic Avenue", 260, 150, true, "Yellow"),
		NewProperty(27, 19, "Ventnor Avenue", 260, 150, true, "Yellow"),
		NewProperty(28, 20, "Water Works", 150, 0, false, UTILITY),
		NewProperty(29, 21, "Marvin Gardens", 280, 150, true, "Yellow"),
		NewProperty(31, 22, "Pacific Avenue", 300, 200, true, "Green"),
		NewProperty(32, 23, "North Carolina Avenue", 300, 200, true, "Green"),
		NewProperty(34, 24, "Pennsylvania Avenue", 320, 200, true, "Green"),
		NewProperty(35, 25, "Short Line", 200, 0, false, RAILROAD),
		NewProperty(37, 26, "Park Place", 350, 200, true, "Dark Blue"),
		NewProperty(39, 27, "Boardwalk", 400, 200, true, "Dark Blue"),
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
	}

	g.logger.Log("Game initialized successfully.")

}

func (g *Game) getState() GameState {
	return GameState{
		Players:          g.players,
		Fields:           g.fields,
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
	g.initGame()
	for {
		g.logger.Log(fmt.Sprintf("Round %d", g.round))
		for idx, player := range g.players {
			g.currentPlayerIdx = idx
			g.checkForWinner()
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

func (g *Game) checkForWinner() {
	active_players := g.getActivePlayers()
	if len(active_players) == 0 {
		g.endDraw()
	} else if len(active_players) == 1 {
		g.endWinner(g.players[active_players[0]])
	} else if g.round > g.settings.MAX_ROUNDS {
		g.endRoundLimit()
	}
}

func (g *Game) endRoundLimit() {
	g.logger.Log("Game ended due to round limit reached.")
	winner := g.players[0]
	max_net_worth := g.calculateNetWorth(winner)
	for _, player := range g.players {
		net_worth := g.calculateNetWorth(player)
		if net_worth > max_net_worth {
			max_net_worth = net_worth
			winner = player
		}
	}
	g.logger.Log(fmt.Sprintf("Winner is %s with net worth of %d", winner.Name, max_net_worth))
	os.Exit(0)
}

func (g *Game) endWinner(winner *Player) {
	g.logger.Log(fmt.Sprintf("Game ended. Winner is %s", winner.Name))
	os.Exit(0)
}

func (g *Game) endDraw() {
	g.logger.Log("Game ended in a draw. No players left.")
	os.Exit(0)
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
	if d1 == d2 {
		g.logger.Log("Player rolled doubles, taking another turn")
		g.makeMove(moves_in_a_row+1, 0, 0)
	}
}

func (g *Game) takeAction() {
	player := g.getCurrPlayer()
	field := g.fields[player.CurrentPosition]
	field.Action(g)
}

func (g *Game) jailPlayer() {
	player := g.getCurrPlayer()
	player.IsJailed = true
	player.CurrentPosition = g.settings.JAIL_POSITION
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
	g.logger.Log(fmt.Sprintf("Jail cards left: %d", player.JailCards))
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

	action_list.Actions = []StdAction{NOACTION}
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
	action_details := g.io.GetStdAction(g.currentPlayerIdx, g.getState(), action_list)
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
	player := g.getCurrPlayer()
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

func (g *Game) calculateNetWorth(player *Player) int {
	net_worth := player.Money
	for _, property := range player.Properties {
		if property.IsMortgaged {
			continue
		}
		net_worth += property.Price / 2
		net_worth += property.Houses * (property.HousePrice / 2)
	}
	return net_worth
}

func (g *Game) chargePlayer(player *Player, amount int, target *Player) {
	target_name := "Bank"
	if target != nil {
		target_name = target.Name
	}
	if player.Money >= amount {
		g.logger.Log(fmt.Sprintf("%s charged %d money from player %s", target_name, amount, player.Name))
		player.Charge(amount, target)
		return
	}
	g.logger.Log(fmt.Sprintf("Player %s does not have enough money to pay %d.", player.Name, amount))

	net_worth := g.calculateNetWorth(player)
	if net_worth < amount {
		player.Charge(amount, target)
		g.logger.Log(fmt.Sprintf("Player %s is bankrupt. All his properties go to %s", player.Name, target_name))
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
		action_details := g.io.GetStdAction(g.currentPlayerIdx, g.getState(), action_list)
		g.resolveStandardAction(action_details)
	}
	player.Charge(amount, target)
}

func (g *Game) mortgage(propertyId int) {
	property := g.properties[propertyId]
	player := g.getCurrPlayer()
	player.AddMoney(property.Price / 2)
	property.IsMortgaged = true
	g.logger.Log(fmt.Sprintf("Player %s mortgaged property %s for %d money", player.Name, property.Name, property.Price/2))
}

func (g *Game) sellHouse(propertyId int) {
	property := g.properties[propertyId]
	player := g.getCurrPlayer()
	player.AddMoney(property.HousePrice / 2)
	property.Houses--
	g.logger.Log(fmt.Sprintf("Player %s sold a house on property %s for %d money", player.Name, property.Name, property.HousePrice/2))
}

func (g *Game) buyHouse(propertyId int) {
	property := g.properties[propertyId]
	player := g.getCurrPlayer()
	g.chargePlayer(player, property.HousePrice, nil)
	property.Houses++
	g.logger.Log(fmt.Sprintf("Player %s bought a house on property %s for %d money", player.Name, property.Name, property.HousePrice))
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
	g.logger.Log(fmt.Sprintf("Player %s bought out property %s for %d money", player.Name, property.Name, int(float64(property.Price)*1.1)))
}

func (g *Game) doForNoActionField() {}

func (g *Game) doForChest() {
	//TODO
}

func (g *Game) doForChance() {
	//TODO
}

func (g *Game) doForTaxField(f *TaxField) {
	player := g.getCurrPlayer()
	g.chargePlayer(player, f.Tax, nil)
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
		g.chargePlayer(player, amount, p.Owner)
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
	player.Charge(p.Price, nil)
	player.AddProperty(p)

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
	active_players := g.getActivePlayers()
	bid_player_iterator := slices.Index[[]int](active_players, first_player_id)
	curr_price := g.settings.MIN_PRICE
	auction_winner := -1
	for len(active_players) > 1 {
		bidding_player_id := active_players[bid_player_iterator]
		bidding_player := g.players[bidding_player_id]
		bid_offer := g.io.BiddingDecision(bidding_player_id, g.getState(), property.PropertyIndex, curr_price)
		if bid_offer <= curr_price {
			g.logger.Log(fmt.Sprintf("Player %s did not bid.", bidding_player.Name))
			active_players = append(active_players[:bid_player_iterator], active_players[bid_player_iterator+1:]...)
		} else if bid_offer > bidding_player.Money {
			g.logger.Log(fmt.Sprintf("Player %s does not have enough money to bid %d. Going bankrupt.", bidding_player.Name, bid_offer))
			bidding_player.Charge(bid_offer, nil)
			active_players = append(active_players[:bid_player_iterator], active_players[bid_player_iterator+1:]...)
		} else {
			g.logger.Log(fmt.Sprintf("Player %s bid %d", bidding_player.Name, bid_offer))
			curr_price = bid_offer
			auction_winner = bidding_player_id
		}
		bid_player_iterator++
		if bid_player_iterator >= len(active_players) {
			bid_player_iterator = 0
		}
	}
	if auction_winner == -1 {
		g.logger.Log("Auction ended without any bids.")
		return
	}
	winner := g.players[auction_winner]
	g.logger.Log(fmt.Sprintf("Player %s won the auction with a bid of %d", winner.Name, curr_price))
	winner.Charge(curr_price, nil)
	winner.AddProperty(property)
}
