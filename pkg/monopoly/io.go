package monopoly

import "fmt"

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

var StdActionNames = map[StdAction]string{
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

var JailActionNames = map[JailAction]string{
	ROLL_DICE: "ROLL DICE",
	BAIL:      "BAIL",
	CARD:      "USE CARD",
}

type GameState struct {
	Players          []*Player
	Properties       []*Property
	Round            int
	CurrentPlayerIdx int
	Charge           int // In case of a charge that would result in a player going bankrupt
}

func (s GameState) String() string {
	result := "========================="
	result = fmt.Sprintf("ROUND %d || PLAYER %d", s.Round, s.CurrentPlayerIdx)
	result = "========================="
	result = "PLAYERS:\n"
	for i, p := range s.Players {
		status := "----"
		if !p.IsBankrupt {
			status = "DEAD"
		} else if p.IsJailed {
			status = "JAIL"
		}
		result +=
			fmt.Sprintf("%d %s %s %d$ %dcard\n", i, p.Name, status, p.Money, p.JailCards)

		for _, propId := range p.Properties {
			prop := s.Properties[propId]
			owner := ""
			if prop.Owner != p {

				owner = "BANK"
				if prop.Owner != nil {
					owner = prop.Owner.Name
				}
			}
			mortgaged := "---------"
			if prop.IsMortgaged {
				mortgaged = "MORTGAGED"
			}
			result += fmt.Sprintf("    %s %d %d %s %s %dHouse\n",
				owner, prop.FieldIndex, prop.PropertyIndex, prop.Name, mortgaged, prop.Houses)
		}
	}
	result += "\nPROPERTIES:\n"
	for _, prop := range s.Properties {
		ownerName := "-------"
		if prop.Owner != nil {
			ownerName = prop.Owner.Name
		}
		mortgaged := "---------"
		if prop.IsMortgaged {
			mortgaged = "MORTGAGED"
		}
		result += fmt.Sprintf("field%d property%d %s %s %s %dHouse %v %s %d$ %d$\n",
			prop.FieldIndex, prop.PropertyIndex, prop.Name, ownerName, mortgaged, prop.Houses, prop.CanBuildHouse, prop.Set, prop.Price, prop.HousePrice)
	}
	return result
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
	Init(int)
	GetStdAction(player int, state GameState, availableActions FullActionList) ActionDetails
	GetJailAction(player int, state GameState, available []JailAction) JailAction
	BuyDecision(player int, state GameState, propertyId int) bool
	BuyFromPlayerDecision(player int, state GameState, propertyId int, price int) bool
	SellToPlayerDecision(player int, state GameState, propertyId int, price int) bool
	BiddingDecision(player int, state GameState, propertyId int, currentPrice int) int
}
