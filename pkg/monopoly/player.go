package monopoly

type Player struct {
	Name            string
	Money           int
	Properties      []*Property
	Sets            []string
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

func (p *Player) Charge(count int) {
	p.Money -= count
	if p.Money < 0 {
		p.GoBankrupt()
	}
}

func (p *Player) GoBankrupt() {
	p.IsBankrupt = true
	p.Properties = nil
	p.CurrentPosition = 0
	p.Money = 0
}
