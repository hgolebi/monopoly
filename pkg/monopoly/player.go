package monopoly

type Player struct {
	name           string
	money          int
	properties     []*Property
	currenPosition int
	isBankrupt     bool
	isJailed       bool
	jailCards      int
	roundsInJail   int
}

func (p *Player) AddMoney(count int) {
	panic("")
}

func (p *Player) SetPosition(pos int) {
	p.currenPosition = pos
}
