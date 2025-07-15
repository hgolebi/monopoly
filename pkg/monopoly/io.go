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
}

type IMonopoly_IO interface {
	GetStdAction(state GameState, available []StdAction) StdAction
	GetProperty(state GameState, available []int, action StdAction) int
	GetPlayer(state GameState, available []int) int
	GetMoney(state GameState, min int, max int) int
	GetJailAction(state GameState, available []JailAction) JailAction
	BuyDecision(state GameState, propertyId int) bool
}
