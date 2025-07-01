package monopoly

type Action int

const (
	QUIT Action = iota
	JAIL_ROLL_DICE
	JAIL_BAIL
	JAIL_CARD
	NOACTION
	MORTGAGE
	BUYOUT
	SELLOFFER
	BUYOFFER
	BUYHOUSE
	SELLHOUSE
	BUY
)

var actionNames = map[Action]string{
	JAIL_ROLL_DICE: "JAIL_ROLL_DICE",
	JAIL_BAIL:      "JAIL_BAIL",
	JAIL_CARD:      "JAIL_CARD",
	NOACTION:       "NOACTION",
	MORTGAGE:       "MORTGAGE",
	BUYOUT:         "BUYOUT",
	SELLOFFER:      "SELLOFFER",
	BUYOFFER:       "BUYOFFER",
	BUYHOUSE:       "BUYHOUSE",
	SELLHOUSE:      "SELLHOUSE",
	BUY:            "BUY",
}

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

type GameState struct {
	Players          []*Player
	Fields           []Field
	Round            int
	CurrentPlayerIdx int
}

type IMonopoly_IO interface {
	GetAction(availableActions FullActionList, state GameState) ActionDetails
}
