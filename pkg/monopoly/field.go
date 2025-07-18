package monopoly

type Field interface {
	Action(*Game)
	GetName() string
}

type Property struct {
	FieldIndex    int
	PropertyIndex int
	Name          string
	Price         int
	HousePrice    int
	Owner         int
	IsMortgaged   bool
	CanBuildHouse bool
	Houses        int
	Set           string
}

type NoActionField struct {
	FieldIndex int
	Name       string
}

type GoToJailField struct {
	FieldIndex int
}

type Chest struct {
	FieldIndex int
}

type Chance struct {
	FieldIndex int
}

type TaxField struct {
	FieldIndex int
	Name       string
	Tax        int
}

func NewProperty(field_id int, property_id int, name string, price int, house_price int, can_build bool, set string) *Property {
	if price < 0 || house_price < 0 {
		panic("Price and house price cannot be negative")
	}
	return &Property{
		FieldIndex:    field_id,
		PropertyIndex: property_id,
		Name:          name,
		Price:         price,
		HousePrice:    house_price,
		CanBuildHouse: can_build,
		Set:           set,
		IsMortgaged:   false,
		Houses:        0,
	}
}

func (f *NoActionField) Action(game *Game) {
	game.doForNoActionField()
}

func (f *Property) Action(game *Game) {
	game.doForProperty(f)
}

func (f *GoToJailField) Action(game *Game) {
	game.doForGoToJailField()
}

func (f *Chest) Action(game *Game) {
	game.doForChest()
}

func (f *Chance) Action(game *Game) {
	game.doForChance()
}

func (f *TaxField) Action(game *Game) {
	game.doForTaxField(f)
}

func (f *NoActionField) GetName() string {
	return f.Name
}

func (f *Property) GetName() string {
	return f.Name
}

func (f *GoToJailField) GetName() string {
	return "Go to Jail Field"
}

func (f *Chest) GetName() string {
	return "Chest Field"
}

func (f *Chance) GetName() string {
	return "Chance Field"
}

func (f *TaxField) GetName() string {
	return f.Name
}
