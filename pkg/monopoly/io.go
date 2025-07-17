package monopoly

type StdAction int

const (
	NOACTION StdAction = iota
	MORTGAGE
	BUYOUT
	SELLOFFER
	BUYOFFER
	BUYHOUSE
	SELLHOUSE
)

var stdActionNames = map[StdAction]string{
	NOACTION:  "NOACTION",
	MORTGAGE:  "MORTGAGE",
	BUYOUT:    "BUYOUT",
	SELLOFFER: "SELLOFFER",
	BUYOFFER:  "BUYOFFER",
	BUYHOUSE:  "BUYHOUSE",
	SELLHOUSE: "SELLHOUSE",
}

type JailAction int

const (
	ROLL_DICE JailAction = iota
	BAIL
	CARD
)

var jailActionNames = map[JailAction]string{
	ROLL_DICE: "ROLL DICE",
	BAIL:      "BAIL",
	CARD:      "USE CARD",
}

type GameState struct {
	Players          []*Player
	Fields           []Field
	Round            int
	CurrentPlayerIdx int
	Charge           int // In case of a charge that would result in a player going bankrupt
}

type FullActionList struct {
	Actions          []StdAction
	MortgageList     []int
	BuyOutList       []int
	SellPropertyList []int
	BuyPropertyList  []int
	BuyHouseList     []int
	SellHouseList    []int
}

type ActionDetails struct {
	Action     StdAction
	PropertyId int
	Price      int
	PlayerId   int
}

type IMonopoly_IO interface {
	GetStdAction(player int, state GameState, availableActions FullActionList) ActionDetails
	GetJailAction(player int, state GameState, available []JailAction) JailAction
	BuyDecision(player int, state GameState, propertyId int) bool
	BuyFromPlayerDecision(player int, state GameState, propertyId int, price int) bool
	SellToPlayerDecision(player int, state GameState, propertyId int, price int) bool
	BiddingDecision(player int, state GameState, propertyId int, currentPrice int) int
}
