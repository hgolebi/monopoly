package monopoly

type Field interface {
	Action(*Game)
}

type Property struct {
	Name  string
	Price int
	Value int
	Tax   int
	Owner *Player
}

func (p *Property) Action(game *Game) {
	game.doForProperty(p)
}
