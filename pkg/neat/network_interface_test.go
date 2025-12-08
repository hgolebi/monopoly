package neatnetwork

import (
	"monopoly/pkg/monopoly"
	"testing"

	"github.com/stretchr/testify/assert"
)

// - { id: 0,  trait_id: 0, type: INPT, activation: NullActivation }  # Property 0 OWNER
// - { id: 1,  trait_id: 0, type: INPT, activation: NullActivation }  # Property 0 IS_MORTGAGED
// - { id: 2,  trait_id: 0, type: INPT, activation: NullActivation }  # Property 0 HOUSES
// - { id: 3,  trait_id: 0, type: INPT, activation: NullActivation }  # Property 1 OWNER
// - { id: 4,  trait_id: 0, type: INPT, activation: NullActivation }  # Property 1 IS_MORTGAGED
// - { id: 5,  trait_id: 0, type: INPT, activation: NullActivation }  # Property 1 HOUSES
// - { id: 6,  trait_id: 0, type: INPT, activation: NullActivation }  # Property 2 OWNER
// - { id: 7,  trait_id: 0, type: INPT, activation: NullActivation }  # Property 2 IS_MORTGAGED
// - { id: 8,  trait_id: 0, type: INPT, activation: NullActivation }  # Property 3 OWNER
// - { id: 9,  trait_id: 0, type: INPT, activation: NullActivation }  # Property 3 IS_MORTGAGED
// - { id: 10, trait_id: 0, type: INPT, activation: NullActivation }  # Property 3 HOUSES
// - { id: 11, trait_id: 0, type: INPT, activation: NullActivation }  # Property 4 OWNER
// - { id: 12, trait_id: 0, type: INPT, activation: NullActivation }  # Property 4 IS_MORTGAGED
// - { id: 13, trait_id: 0, type: INPT, activation: NullActivation }  # Property 4 HOUSES
// - { id: 14, trait_id: 0, type: INPT, activation: NullActivation }  # Property 5 OWNER
// - { id: 15, trait_id: 0, type: INPT, activation: NullActivation }  # Property 5 IS_MORTGAGED
// - { id: 16, trait_id: 0, type: INPT, activation: NullActivation }  # Property 5 HOUSES
// - { id: 17, trait_id: 0, type: INPT, activation: NullActivation }  # Property 6 OWNER
// - { id: 18, trait_id: 0, type: INPT, activation: NullActivation }  # Property 6 IS_MORTGAGED
// - { id: 19, trait_id: 0, type: INPT, activation: NullActivation }  # Property 7 OWNER
// - { id: 20, trait_id: 0, type: INPT, activation: NullActivation }  # Property 7 IS_MORTGAGED
// - { id: 21, trait_id: 0, type: INPT, activation: NullActivation }  # Property 7 HOUSES
// - { id: 22, trait_id: 0, type: INPT, activation: NullActivation }  # Property 8 OWNER
// - { id: 23, trait_id: 0, type: INPT, activation: NullActivation }  # Property 8 IS_MORTGAGED
// - { id: 24, trait_id: 0, type: INPT, activation: NullActivation }  # Property 8 HOUSES
// - { id: 25, trait_id: 0, type: INPT, activation: NullActivation }  # Property 9 OWNER
// - { id: 26, trait_id: 0, type: INPT, activation: NullActivation }  # Property 9 IS_MORTGAGED
// - { id: 27, trait_id: 0, type: INPT, activation: NullActivation }  # Property 9 HOUSES
// - { id: 28, trait_id: 0, type: INPT, activation: NullActivation }  # Property 10 OWNER
// - { id: 29, trait_id: 0, type: INPT, activation: NullActivation }  # Property 10 IS_MORTGAGED
// - { id: 30, trait_id: 0, type: INPT, activation: NullActivation }  # Property 11 OWNER
// - { id: 31, trait_id: 0, type: INPT, activation: NullActivation }  # Property 11 IS_MORTGAGED
// - { id: 32, trait_id: 0, type: INPT, activation: NullActivation }  # Property 11 HOUSES
// - { id: 33, trait_id: 0, type: INPT, activation: NullActivation }  # Property 12 OWNER
// - { id: 34, trait_id: 0, type: INPT, activation: NullActivation }  # Property 12 IS_MORTGAGED
// - { id: 35, trait_id: 0, type: INPT, activation: NullActivation }  # Property 12 HOUSES
// - { id: 36, trait_id: 0, type: INPT, activation: NullActivation }  # Property 13 OWNER
// - { id: 37, trait_id: 0, type: INPT, activation: NullActivation }  # Property 13 IS_MORTGAGED
// - { id: 38, trait_id: 0, type: INPT, activation: NullActivation }  # Property 13 HOUSES
// - { id: 39, trait_id: 0, type: INPT, activation: NullActivation }  # Property 14 OWNER
// - { id: 40, trait_id: 0, type: INPT, activation: NullActivation }  # Property 14 IS_MORTGAGED
// - { id: 41, trait_id: 0, type: INPT, activation: NullActivation }  # Property 15 OWNER
// - { id: 42, trait_id: 0, type: INPT, activation: NullActivation }  # Property 15 IS_MORTGAGED
// - { id: 43, trait_id: 0, type: INPT, activation: NullActivation }  # Property 15 HOUSES
// - { id: 44, trait_id: 0, type: INPT, activation: NullActivation }  # Property 16 OWNER
// - { id: 45, trait_id: 0, type: INPT, activation: NullActivation }  # Property 16 IS_MORTGAGED
// - { id: 46, trait_id: 0, type: INPT, activation: NullActivation }  # Property 16 HOUSES
// - { id: 47, trait_id: 0, type: INPT, activation: NullActivation }  # Property 17 OWNER
// - { id: 48, trait_id: 0, type: INPT, activation: NullActivation }  # Property 17 IS_MORTGAGED
// - { id: 49, trait_id: 0, type: INPT, activation: NullActivation }  # Property 17 HOUSES
// - { id: 50, trait_id: 0, type: INPT, activation: NullActivation }  # Property 18 OWNER
// - { id: 51, trait_id: 0, type: INPT, activation: NullActivation }  # Property 18 IS_MORTGAGED
// - { id: 52, trait_id: 0, type: INPT, activation: NullActivation }  # Property 18 HOUSES
// - { id: 53, trait_id: 0, type: INPT, activation: NullActivation }  # Property 19 OWNER
// - { id: 54, trait_id: 0, type: INPT, activation: NullActivation }  # Property 19 IS_MORTGAGED
// - { id: 55, trait_id: 0, type: INPT, activation: NullActivation }  # Property 20 OWNER
// - { id: 56, trait_id: 0, type: INPT, activation: NullActivation }  # Property 20 IS_MORTGAGED
// - { id: 57, trait_id: 0, type: INPT, activation: NullActivation }  # Property 20 HOUSES
// - { id: 58, trait_id: 0, type: INPT, activation: NullActivation }  # Property 21 OWNER
// - { id: 59, trait_id: 0, type: INPT, activation: NullActivation }  # Property 21 IS_MORTGAGED
// - { id: 60, trait_id: 0, type: INPT, activation: NullActivation }  # Property 21 HOUSES
// - { id: 61, trait_id: 0, type: INPT, activation: NullActivation }  # Property 22 OWNER
// - { id: 62, trait_id: 0, type: INPT, activation: NullActivation }  # Property 22 IS_MORTGAGED
// - { id: 63, trait_id: 0, type: INPT, activation: NullActivation }  # Property 23 OWNER
// - { id: 64, trait_id: 0, type: INPT, activation: NullActivation }  # Property 23 IS_MORTGAGED
// - { id: 65, trait_id: 0, type: INPT, activation: NullActivation }  # Property 23 HOUSES
// - { id: 66, trait_id: 0, type: INPT, activation: NullActivation }  # Property 24 OWNER
// - { id: 67, trait_id: 0, type: INPT, activation: NullActivation }  # Property 24 IS_MORTGAGED
// - { id: 68, trait_id: 0, type: INPT, activation: NullActivation }  # Property 24 HOUSES
// - { id: 69, trait_id: 0, type: INPT, activation: NullActivation }  # Property 25 OWNER
// - { id: 70, trait_id: 0, type: INPT, activation: NullActivation }  # Property 25 IS_MORTGAGED
// - { id: 71, trait_id: 0, type: INPT, activation: NullActivation }  # Property 25 HOUSES
// - { id: 72, trait_id: 0, type: INPT, activation: NullActivation }  # Property 26 OWNER
// - { id: 73, trait_id: 0, type: INPT, activation: NullActivation }  # Property 26 IS_MORTGAGED
// - { id: 74, trait_id: 0, type: INPT, activation: NullActivation }  # Property 26 HOUSES
// - { id: 75, trait_id: 0, type: INPT, activation: NullActivation }  # Property 27 OWNER
// - { id: 76, trait_id: 0, type: INPT, activation: NullActivation }  # Property 27 IS_MORTGAGED
// - { id: 77, trait_id: 0, type: INPT, activation: NullActivation }  # Property 27 HOUSES

// # Player inputs (3 players × 2 attributes = 6 inputs, IDs 78-83)
// - { id: 78, trait_id: 0, type: INPT, activation: NullActivation }  # Player 1 IS_ALIVE
// - { id: 79, trait_id: 0, type: INPT, activation: NullActivation }  # Player 1 MONEY
// - { id: 80, trait_id: 0, type: INPT, activation: NullActivation }  # Player 2 IS_ALIVE
// - { id: 81, trait_id: 0, type: INPT, activation: NullActivation }  # Player 2 MONEY
// - { id: 82, trait_id: 0, type: INPT, activation: NullActivation }  # Player 3 IS_ALIVE
// - { id: 83, trait_id: 0, type: INPT, activation: NullActivation }  # Player 3 MONEY

// # Current player dedicated inputs (IDs 84-88)
// - { id: 84, trait_id: 0, type: INPT, activation: NullActivation }  # CURRENT_PLAYER_IS_ALIVE
// - { id: 85, trait_id: 0, type: INPT, activation: NullActivation }  # CURRENT_PLAYER_IS_JAILED
// - { id: 86, trait_id: 0, type: INPT, activation: NullActivation }  # CURRENT_PLAYER_POSITION
// - { id: 87, trait_id: 0, type: INPT, activation: NullActivation }  # CURRENT_PLAYER_MONEY
// - { id: 88, trait_id: 0, type: INPT, activation: NullActivation }  # CURRENT_PLAYER_JAIL_CARDS

// # Base inputs (IDs 89-94)
// - { id: 89, trait_id: 0, type: INPT, activation: NullActivation }  # DECISION_CONTEXT
// - { id: 90, trait_id: 0, type: INPT, activation: NullActivation }  # PROPERTY_ID
// - { id: 91, trait_id: 0, type: INPT, activation: NullActivation }  # PRIC91E
// - { id: 92, trait_id: 0, type: INPT, activation: NullActivation }  # CURR_BID
// - { id: 93, trait_id: 0, type: INPT, activation: NullActivation }  # CURR_BID_WINNER
// - { id: 94, trait_id: 0, type: INPT, activation: NullActivation }  # CHARGE

// # Available standard action inputs (IDs 95-101)
// - { id: 95, trait_id: 0, type: INPT, activation: NullActivation }  # NOACTION available
// - { id: 96, trait_id: 0, type: INPT, activation: NullActivation }  # MORTGAGE available
// - { id: 97, trait_id: 0, type: INPT, activation: NullActivation }  # BUYOUT available
// - { id: 98, trait_id: 0, type: INPT, activation: NullActivation }  # SELLOFFER available
// - { id: 99, trait_id: 0, type: INPT, activation: NullActivation }  # BUYOFFER available
// - { id: 100, trait_id: 0, type: INPT, activation: NullActivation } # BUYHOUSE available
// - { id: 101, trait_id: 0, type: INPT, activation: NullActivation } # SELLHOUSE available

// # OUTPUT nodes (IDs 102-116)
// - { id: 102, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # BUY_DECISION (0)
// - { id: 103, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # BID_DECISION (1)
// - { id: 104, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # BUY_FROM_PLAYER (2)
// - { id: 105, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # SELL_TO_PLAYER (3)
// - { id: 106, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # NO_ACTION (4)
// - { id: 107, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # MORTGAGE (5)
// - { id: 108, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # BUYOUT (6)
// - { id: 109, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # SELL_OFFER (7)
// - { id: 110, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # BUY_OFFER (8)
// - { id: 111, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # BUY_HOUSE (9)
// - { id: 112, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # SELL_HOUSE (10)
// - { id: 113, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # PLAYER_1 (11)
// - { id: 114, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # PLAYER_2 (12)
// - { id: 115, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # PLAYER_3 (13)
// - { id: 116, trait_id: 0, type: OUTP, activation: SigmoidSteepenedActivation } # PRICE (14)

func TestLoadPlayerState(t *testing.T) {
	var tests = []struct {
		id                   int
		is_alive             bool
		money                int
		expectedIsAliveInput int
		expectedIsAlive      float64
		expectedMoneyInput   int
		expectedMoney        float64
	}{
		{0, true, 1000, 78, -1.23, 79, -1.23},
		{1, false, 0.0, 78, 0.0, 79, 0.0},
		{2, true, 2000, 80, 1.0, 81, 1.0},
		{3, false, 500, 82, 0.0, 83, 0.25},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		for i := range ms {
			ms[i] = -1.23
		}
		player := monopoly.NewPlayer(tt.id, "", tt.money)
		player.IsBankrupt = !tt.is_alive
		ms.loadPlayerState(tt.id, player)
		assert.InDelta(t, tt.expectedIsAlive, ms[tt.expectedIsAliveInput], 0.0001)
		assert.InDelta(t, tt.expectedMoney, ms[tt.expectedMoneyInput], 0.0001)
	}
}

func TestLoadPropertyState(t *testing.T) {
	var tests = []struct {
		id                      int
		ownerId                 int
		isMortgaged             bool
		houses                  int
		expectedOwnerInput      int
		expectedOwnerValue      float64
		expectedIsMortgagedInpt int
		expectedIsMortgaged     float64
		expectedHousesInput     int
		expectedHousesValue     float64
	}{
		{0, -1, false, 0, 0, -1.23, 1, 0.0, 2, 0.0},
		{1, 1, true, 3, 3, 0.5, 4, 1.0, 5, 0.6},
		{2, 2, false, 5, 6, 0.75, 7, 0.0, 8, -1.23},
		{3, 3, true, 2, 8, 1.0, 9, 1.0, 10, 0.4},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		for i := range ms {
			ms[i] = -1.23
		}
		var owner *monopoly.Player
		if tt.ownerId >= 0 {
			owner = monopoly.NewPlayer(tt.ownerId, "", 0)
		} else {
			owner = nil
		}
		property := monopoly.Property{
			Owner:         owner,
			IsMortgaged:   tt.isMortgaged,
			Houses:        tt.houses,
			CanBuildHouse: true,
		}
		ms.loadPropertyState(tt.id, &property, 0)
		assert.InDelta(t, tt.expectedOwnerValue, ms[tt.expectedOwnerInput], 0.0001)
		assert.InDelta(t, tt.expectedIsMortgaged, ms[tt.expectedIsMortgagedInpt], 0.0001)
		assert.InDelta(t, tt.expectedHousesValue, ms[tt.expectedHousesInput], 0.0001)
	}
}

func TestLoadCurrentPlayerState(t *testing.T) {
	var tests = []struct {
		is_alive          bool
		is_jailed         bool
		position          int
		money             int
		jail_cards        int
		expectedIsAlive   float64
		expectedIsJailed  float64
		expectedPosition  float64
		expectedMoney     float64
		expectedJailCards float64
	}{
		{true, false, 0, 0, 0, 1.0, 0.0, 0.0, 0.0, 0.0},
		{false, true, 39, 1000, 1, 0.0, 1.0, 1.0, 0.5, 0.1},
		{true, true, 19, 2000, 2, 1.0, 1.0, 19.0 / 39.0, 1.0, 0.2},
		{false, false, 1, 10, 3, 0.0, 0.0, 1.0 / 39.0, 0.005, 0.3},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		for i := range ms {
			ms[i] = -1.23
		}
		player := monopoly.NewPlayer(0, "", tt.money)
		player.IsBankrupt = !tt.is_alive
		player.IsJailed = tt.is_jailed
		player.CurrentPosition = tt.position
		player.JailCards = tt.jail_cards
		ms.loadCurrentPlayerState(player)
		assert.InDelta(t, tt.expectedIsAlive, ms[84], 0.0001)
		assert.InDelta(t, tt.expectedIsJailed, ms[85], 0.0001)
		assert.InDelta(t, tt.expectedPosition, ms[86], 0.0001)
		assert.InDelta(t, tt.expectedMoney, ms[87], 0.0001)
		assert.InDelta(t, tt.expectedJailCards, ms[88], 0.0001)
	}
}

func TestLoadDecisionContext(t *testing.T) {
	var tests = []struct {
		decisionContext         DecisionContext
		expectedDecisionContext float64
	}{
		{BUY_DECISION, 0.0},
		{BIDDING_DECISION, 0.2},
		{JAIL_DECISION, 0.4},
		{BUY_FROM_PLAYER, 0.6},
		{SELL_TO_PLAYER, 0.8},
		{STD_ACTION, 1.0},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		for i := range ms {
			ms[i] = -1.23
		}
		ms.LoadDecisionContext(tt.decisionContext)
		assert.InDelta(t, tt.expectedDecisionContext, ms[89], 0.0001)
	}
}

func TestLoadPropertyId(t *testing.T) {
	var tests = []struct {
		propertyId         int
		expectedPropertyId float64
	}{
		{0, 1.0 / 28.0},
		{10, 11.0 / 28.0},
		{27, 1.0},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		for i := range ms {
			ms[i] = -1.23
		}
		ms.LoadPropertyId(tt.propertyId)
		assert.InDelta(t, tt.expectedPropertyId, ms[90], 0.0001)
	}
}

func TestLoadPrice(t *testing.T) {
	var tests = []struct {
		price         int
		expectedPrice float64
	}{
		{0, 0.0},
		{500, 0.25},
		{1000, 0.5},
		{2000, 1.0},
		{10, 0.005},
		{3000, 1.0},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		for i := range ms {
			ms[i] = -1.23
		}
		ms.LoadPrice(tt.price)
		assert.InDelta(t, tt.expectedPrice, ms[91], 0.0001)
	}
}

func TestLoadBiddingInputs(t *testing.T) {
	var tests = []struct {
		currBid               int
		currBidWinner         int
		expectedCurrBid       float64
		expectedCurrBidWinner float64
	}{
		{10, 0, 0.005, 0.25},
		{0, 1, 0.0, 0.5},
		{500, 2, 0.25, 0.75},
		{1000, 3, 0.5, 1.0},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		for i := range ms {
			ms[i] = -1.23
		}
		ms.LoadBiddingInputs(tt.currBid, tt.currBidWinner, 0)
		assert.InDelta(t, tt.expectedCurrBid, ms[92], 0.0001)
		assert.InDelta(t, tt.expectedCurrBidWinner, ms[93], 0.0001)
	}
}

func TestLoadCharge(t *testing.T) {
	var tests = []struct {
		charge         int
		expectedCharge float64
	}{
		{0, 0.0},
		{500, 0.25},
		{1000, 0.5},
		{2000, 1.0},
		{10, 0.005},
		{3000, 1.0},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		for i := range ms {
			ms[i] = -1.23
		}
		ms.LoadCharge(tt.charge)
		assert.InDelta(t, tt.expectedCharge, ms[94], 0.0001)
	}
}

func TestLoadAvailableStdActions(t *testing.T) {
	var tests = []struct {
		availableActions  []monopoly.StdAction
		expectedNoAction  float64
		expectedMortgage  float64
		expectedBuyout    float64
		expectedSellOffer float64
		expectedBuyOffer  float64
		expectedBuyHouse  float64
		expectedSellHouse float64
	}{
		{
			[]monopoly.StdAction{},
			0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		},
		{
			[]monopoly.StdAction{
				monopoly.NOACTION,
			},
			1.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		},
		{
			[]monopoly.StdAction{
				monopoly.MORTGAGE,
			},
			0.0, 1.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		},
		{
			[]monopoly.StdAction{
				monopoly.BUYOUT,
			},
			0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 0.0,
		},
		{
			[]monopoly.StdAction{
				monopoly.SELLOFFER,
			},
			0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0,
		},
		{
			[]monopoly.StdAction{
				monopoly.BUYOFFER,
			},
			0.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0,
		},
		{
			[]monopoly.StdAction{
				monopoly.BUYHOUSE,
			},
			0.0, 0.0, 0.0, 0.0, 0.0, 1.0, 0.0,
		},
		{
			[]monopoly.StdAction{
				monopoly.SELLHOUSE,
			},
			0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 1.0,
		},
		{
			[]monopoly.StdAction{
				monopoly.NOACTION,
				monopoly.MORTGAGE,
				monopoly.BUYOUT,
				monopoly.SELLOFFER,
				monopoly.BUYOFFER,
				monopoly.BUYHOUSE,
				monopoly.SELLHOUSE,
			},
			1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0,
		},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		ms.LoadAvailableStdActions(tt.availableActions)
		assert.InDelta(t, tt.expectedNoAction, ms[95], 0.0001)
		assert.InDelta(t, tt.expectedMortgage, ms[96], 0.0001)
		assert.InDelta(t, tt.expectedBuyout, ms[97], 0.0001)
		assert.InDelta(t, tt.expectedSellOffer, ms[98], 0.0001)
		assert.InDelta(t, tt.expectedBuyOffer, ms[99], 0.0001)
		assert.InDelta(t, tt.expectedBuyHouse, ms[100], 0.0001)
		assert.InDelta(t, tt.expectedSellHouse, ms[101], 0.0001)
	}
}

func TestGetOutputs(t *testing.T) {
	output := []float64{
		0.1,  // BUY_DECISION
		0.2,  // BID_DECISION
		0.3,  // BUY_FROM_PLAYER
		0.4,  // SELL_TO_PLAYER
		0.5,  // NO_ACTION
		0.6,  // MORTGAGE
		0.7,  // BUYOUT
		0.8,  // SELL_OFFER
		0.9,  // BUY_OFFER
		0.10, // BUY_HOUSE
		0.11, // SELL_HOUSE
		0.12, // PLAYER_1
		0.13, // PLAYER_2
		0.14, // PLAYER_3
		0.15, // PRICE
	}
	stdActionValues := GetStdActionOutputValues(output)
	assert.InDelta(t, 0.5, stdActionValues[monopoly.NOACTION], 0.0001)
	assert.InDelta(t, 0.44999999999999996, stdActionValues[monopoly.MORTGAGE], 0.0001) //activation function
	assert.InDelta(t, 0.7, stdActionValues[monopoly.BUYOUT], 0.0001)
	assert.InDelta(t, 0.75, stdActionValues[monopoly.SELLOFFER], 0.0001)     //activation function
	assert.InDelta(t, 0.8859375, stdActionValues[monopoly.BUYOFFER], 0.0001) //activation function
	assert.InDelta(t, 0.10, stdActionValues[monopoly.BUYHOUSE], 0.0001)
	assert.InDelta(t, -0.026142187500000014, stdActionValues[monopoly.SELLHOUSE], 0.0001) //activation function
}

func TestGetPlayerOutputs(t *testing.T) {
	output := []float64{
		0.1,  // BUY_DECISION
		0.2,  // BID_DECISION
		0.3,  // BUY_FROM_PLAYER
		0.4,  // SELL_TO_PLAYER
		0.5,  // NO_ACTION
		0.6,  // MORTGAGE
		0.7,  // BUYOUT
		0.8,  // SELL_OFFER
		0.9,  // BUY_OFFER
		0.10, // BUY_HOUSE
		0.11, // SELL_HOUSE
		0.12, // PLAYER_1
		0.13, // PLAYER_2
		0.14, // PLAYER_3
		0.15, // PRICE
	}
	playerMap := GetPlayerOutputValues(output)
	assert.Equal(t, 0.12, playerMap[1])
	assert.Equal(t, 0.13, playerMap[2])
	assert.Equal(t, 0.14, playerMap[3])
}

func TestGetPriceOutput(t *testing.T) {
	output := []float64{
		0.1,  // BUY_DECISION
		0.2,  // BID_DECISION
		0.3,  // BUY_FROM_PLAYER
		0.4,  // SELL_TO_PLAYER
		0.5,  // NO_ACTION
		0.6,  // MORTGAGE
		0.7,  // BUYOUT
		0.8,  // SELL_OFFER
		0.9,  // BUY_OFFER
		0.10, // BUY_HOUSE
		0.11, // SELL_HOUSE
		0.12, // PLAYER_1
		0.13, // PLAYER_2
		0.14, // PLAYER_3
		0.15, // PRICE
	}
	price := GetPriceOutputValue(output)
	assert.InDelta(t, 300, price, 0.0001)
}
