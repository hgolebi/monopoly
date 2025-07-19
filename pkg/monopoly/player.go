package monopoly

type Player struct {
	ID              int
	Name            string
	Money           int
	Properties      []int
	CurrentPosition int
	IsBankrupt      bool
	IsJailed        bool
	JailCards       int
	roundsInJail    int
}

func NewPlayer(name string, money int) *Player {
	if money < 0 {
		panic("Money cannot be negative")
	}
	return &Player{
		Name:            name,
		Money:           money,
		Properties:      []int{},
		CurrentPosition: 0,
		IsBankrupt:      false,
		IsJailed:        false,
		JailCards:       0,
		roundsInJail:    0,
	}
}

func (p *Player) AddMoney(amount int) {
	if amount < 0 {
		panic("Cannot add negative amount")
	}
	p.Money += amount
}

func (p *Player) RemoveMoney(amount int) {
	if amount < 0 {
		panic("Cannot remove negative amount")
	}
	p.Money -= amount
}

func (p *Player) SetPosition(pos int) {
	p.CurrentPosition = pos
}

func (p *Player) AddProperty(propertyIndex int) {
	for _, prop := range p.Properties {
		if prop == propertyIndex {
			panic("Property already owned by player")
		}
	}
	p.Properties = append(p.Properties, propertyIndex)
}

func (p *Player) RemoveProperty(propertyIndex int) {
	for i, prop := range p.Properties {
		if prop == propertyIndex {
			p.Properties = append(p.Properties[:i], p.Properties[i+1:]...)
			return
		}
	}
	panic("Property not owned by player")
}
