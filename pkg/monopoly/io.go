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
	SellOfferTries   int
	BuyOfferTries    int
	StdActionsUsed   int
	// Negotiation context — set by sendBuyOffer, zero outside active negotiation
	NegotiationRound       int // Number of offer/counteroffer rounds in the current negotiation, starting at 1 for the initial offer
	NegotiationBuyerOffer  int // Current buyer's offer
	NegotiationSellerOffer int // Current seller's counteroffer
	SellerImpasse          bool
	BuyerImpasse           bool
}

func formatStr(str string, length int) string {
	return str + fmt.Sprintf("%*s", length-len(str), " ")
}

func (s GameState) String() string {
	result := "=============================================================================\n"
	result += fmt.Sprintf("  ROUND %d | PLAYER %d\n", s.Round, s.CurrentPlayerIdx)
	result += "\nPLAYERS:\n"
	for i, p := range s.Players {
		status := ""
		if p.IsBankrupt {
			status = "DEAD"
		} else if p.IsJailed {
			status = "JAIL"
		}
		position := p.CurrentPosition
		result +=
			fmt.Sprintf("%d %s %d$ position=%d jail_cards=%d %s\n", i, formatStr(p.Name, 15), p.Money, position, p.JailCards, status)

		for _, propId := range p.Properties {
			prop := s.Properties[propId]
			mortgaged := "---------"
			if prop.IsMortgaged {
				mortgaged = "MORTGAGED"
			}
			result += fmt.Sprintf("    %2d %s houses=%d %s\n",
				prop.PropertyIndex, formatStr(prop.Name, 10), prop.Houses, mortgaged)
		}
	}
	result += "\nPROPERTIES:\n"
	for _, prop := range s.Properties {
		ownerName := "BANK"
		if prop.Owner != nil {
			ownerName = prop.Owner.Name
		}
		mortgaged := "---------"
		if prop.IsMortgaged {
			mortgaged = "MORTGAGED"
		}
		result += fmt.Sprintf("%2d %s %s %s houses=%d price=%d$ housePrice=%d$\n",
			prop.PropertyIndex, formatStr(prop.Name, 12), formatStr(ownerName, 15), mortgaged, prop.Houses, prop.Price, prop.HousePrice)
	}
	result += "=============================================================================\n"
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
	Players    []int
}

type FinishOption int

const (
	WIN FinishOption = iota
	DRAW
	ROUND_LIMIT
)

type IMonopoly_IO interface {
	Init() []string
	GetStdAction(player int, state GameState, availableActions FullActionList) ActionDetails
	GetJailAction(player int, state GameState, available []JailAction) JailAction
	BuyDecision(player int, state GameState, propertyId int) bool
	// BuyFromPlayerDecision is called during BUYOFFER negotiation when the seller has made a counteroffer.
	// Returns (false, _) to withdraw immediately.
	// Returns (true, price) to continue: price >= sellerOffer = accept, price <= buyerPrice = impasse signal, in between = raise offer.
	BuyFromPlayerDecision(player int, state GameState, propertyId int, sellerOffer int, tradingPartnerId int) (bool, int)
	// SellToPlayerDecision is called during BUYOFFER negotiation when the buyer has made an offer.
	// Returns (false, _) to hard-reject immediately.
	// Returns (true, price) to continue: price <= buyerOffer = accept, price >= lastSellerPrice = impasse signal, in between = lower counteroffer.
	SellToPlayerDecision(player int, state GameState, propertyId int, buyerOffer int, tradingPartnerId int) (bool, int)
	BiddingDecision(player int, state GameState, propertyId int, currentPrice int, currentWinner int) int
	Finish(f FinishOption, winner int, state GameState)
}
