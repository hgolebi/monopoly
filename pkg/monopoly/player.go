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

func (p *Player) AddMoney(amount int) {
	if amount < 0 {
		panic("Cannot add negative amount")
	}
	p.Money += amount
}

func (p *Player) AddProperty(property *Property) {
	for _, elem := range p.Properties {
		if elem == property {
			panic("Adding property which is already owned")
		}
	}
	p.Properties = append(p.Properties, property)
	property.Owner = p
}

func (p *Player) SetPosition(pos int) {
	p.CurrentPosition = pos
}

func (p *Player) Charge(amount int, target *Player) {
	if p.Money < amount {
		p.GoBankrupt(target)
	}
	p.Money -= amount
	if target != nil {
		target.AddMoney(amount)
	}
}

func (p *Player) GoBankrupt(target *Player) {
	if target != nil {
		target.AddMoney(max(0, p.Money))
		for _, property := range p.Properties {
			p.TransferProperty(target, property)
		}
	}
	p.IsBankrupt = true
	p.Properties = nil
	p.CurrentPosition = -1
	p.Money = -1
}

func (p *Player) TransferProperty(target *Player, property *Property) {
	var new_list []*Property
	found := false
	for _, elem := range p.Properties {
		if elem == property {
			found = true
			target.AddProperty(elem)
		} else {
			new_list = append(new_list, elem)
		}
	}
	p.Properties = new_list
	if !found {
		panic("Property not found during transfer")
	}
}
