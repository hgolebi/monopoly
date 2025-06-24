package monopoly

type Field interface {
	Action(*Game)
}

type Property struct {
	FieldIndex    int
	PropertyIndex int
	Name          string
	Price         int
	Owner         *Player
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
	Name       int
	Tax        int
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
