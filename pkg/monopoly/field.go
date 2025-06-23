package monopoly

type Field interface {
	Action(*Game)
}

type Property struct {
	Index       int
	Name        string
	Price       int
	Value       int
	Tax         int
	Owner       int
	IsMortgaged bool
	Houses      int
	Set         string
}

func (p *Property) Action(game *Game) {
	game.doForProperty(p)
}
