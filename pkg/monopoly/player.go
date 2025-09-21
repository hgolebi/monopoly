package monopoly

import "fmt"

type Player struct {
	ID                  int
	Name                string
	Money               int
	Properties          []int
	CurrentPosition     int
	IsBankrupt          bool
	IsJailed            bool
	JailCards           int
	RoundsInJail        int
	RoundWhenBankrupted int
}

func NewPlayer(id int, name string, money int) *Player {
	if money < 0 {
		panic(fmt.Sprintf("Initial money cannot be negative, got: %d", money))
	}
	return &Player{
		ID:              id,
		Name:            name,
		Money:           money,
		Properties:      []int{},
		CurrentPosition: 0,
		IsBankrupt:      false,
		IsJailed:        false,
		JailCards:       0,
		RoundsInJail:    0,
	}
}

func (p *Player) AddMoney(amount int) {
	if amount < 0 {
		panic(fmt.Sprintf("Cannot add negative amount to player %s, amount: %d", p.Name, amount))
	}
	p.Money += amount
}

func (p *Player) RemoveMoney(amount int) {
	if amount < 0 {
		panic(fmt.Sprintf("Cannot remove negative amount from player %s, amount: %d", p.Name, amount))
	}
	p.Money -= amount
}

func (p *Player) SetPosition(pos int) {
	p.CurrentPosition = pos
}

func (p *Player) AddProperty(propertyIndex int) {
	for _, prop := range p.Properties {
		if prop == propertyIndex {
			panic(fmt.Sprintf("Property %d already owned by player %s", propertyIndex, p.Name))
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
	panic(fmt.Sprintf("Property %d not owned by player %s", propertyIndex, p.Name))
}
