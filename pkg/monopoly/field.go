package monopoly

type Field interface {
	Action(*Game)
}

type Property struct {
	name  string
	price int
	value int
	tax   int
	owner *Player
}

func (p *Property) Action(game *Game) {
	game.doForProperty(p)
}
