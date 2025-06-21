package monopoly

import (
	"math/rand"
)

type GameSettings struct {
	MAX_ROUNDS       int
	START_PASS_MONEY int
	JAIL_POSITION    int
	JAIL_BAIL        int
}

type Game struct {
	players          []*Player
	fields           []Field
	alivePlayers     []int
	currentPlayerIdx int
	round            int
	settings         GameSettings
	io               IMonopoly_IO
}

func (g *Game) initGame() {
	//TODO
}

func (g *Game) Start() {
	g.initGame()
	for {
		for idx, player := range g.players {
			g.currentPlayerIdx = idx
			g.checkForWinner()
			if player.isBankrupt {
				continue
			}
			if player.isJailed {
				g.handleJail()
				continue
			}
			g.makeMove(1)
		}
		g.round++
	}
}

func (g *Game) checkForWinner() {
	players_alive := 0
	for _, player := range g.players {
		if !player.isBankrupt {
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

func (g *Game) makeMove(moves_in_a_row int) {
	d1, d2 := g.rollDice()
	if moves_in_a_row >= 3 && d1 == d2 {
		g.jailPlayer()
		return
	}
	g.movePlayer(d1 + d2)
	g.takeAction()
	player := g.players[g.currentPlayerIdx]
	if player.isJailed {
		return
	}
	if d1 == d2 {
		g.makeMove(moves_in_a_row + 1)
	}
}

func (g *Game) takeAction() {
	panic("unimplemented")
}

func (g *Game) jailPlayer() {
	player := g.players[g.currentPlayerIdx]
	player.isJailed = true
	player.currenPosition = g.settings.JAIL_POSITION
}

func (g *Game) movePlayer(count int) {
	player := g.players[g.currentPlayerIdx]
	curr_pos := player.currenPosition
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

func (g *Game) doForProperty(property *Property) {

}

func (g *Game) handleJail() {
	player := g.players[g.currentPlayerIdx]
	g.standardActions()
	var action_list = FullActionList{
		Actions: []Action{},
	}
	action_list.Actions = append(action_list.Actions, JAIL_BAIL)
	if player.jailCards > 0 {
		action_list.Actions = append(action_list.Actions, JAIL_CARD)
	}
	if player.roundsInJail < 3 {
		action_list.Actions = append(action_list.Actions, JAIL_ROLL_DICE)
	}

}

func (g *Game) standardActions() {
	panic("unimplemented")
}
