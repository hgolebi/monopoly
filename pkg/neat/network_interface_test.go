package neatnetwork

import (
	"monopoly/pkg/monopoly"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
