package monopoly

type Action int

const (
	JAIL_ROLL_DICE Action = iota
	JAIL_BAIL
	JAIL_CARD
	MORTGAGE
)

type FullActionList struct {
	Actions          []Action
	MortgageList     []int
	BuyOutList       []int
	SellPropertyList []int
	BuyPropertyList  []int
	BuyHouseList     []int
	SellHouseList    []int
}

type ActionDetails struct {
	Action     Action
	PropertyId int
	Price      int
	PlayerId   int
}

type IMonopoly_IO interface {
	SendState(state GameState)
	GetAction(availableActions FullActionList) ActionDetails
}
