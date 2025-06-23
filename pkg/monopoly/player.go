package monopoly

type Player struct {
	Name            string
	Money           int
	Properties      []*Property
	CurrentPosition int
	IsBankrupt      bool
	IsJailed        bool
	JailCards       int
	roundsInJail    int
}

func (p *Player) AddMoney(count int) {
	panic("")
}

func (p *Player) SetPosition(pos int) {
	p.CurrentPosition = pos
}
