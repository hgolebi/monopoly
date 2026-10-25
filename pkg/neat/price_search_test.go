package neatnetwork

import (
	"math/bits"
	"math/rand"
	"testing"

	"monopoly/pkg/monopoly"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

// linearHighest and linearLowest are the old step-by-step searches, the reference for the binary ones.
func linearHighest(lo, hi int, accept func(int) bool) (int, bool) {
	best, found := 0, false
	if hi < lo {
		return best, found
	}
	for k := 0; k <= priceGridSize(lo, hi); k++ {
		if !accept(priceAt(lo, hi, k)) {
			break
		}
		best, found = priceAt(lo, hi, k), true
	}
	return best, found
}

func linearLowest(lo, hi int, accept func(int) bool) (int, bool) {
	if hi < lo {
		return 0, false
	}
	for k := 0; k <= priceGridSize(lo, hi); k++ {
		if accept(priceAt(lo, hi, k)) {
			return priceAt(lo, hi, k), true
		}
	}
	return 0, false
}

func TestPriceSearchMatchesLinearForMonotonicAnswers(t *testing.T) {
	for _, lo := range []int{0, 10, 30, 55} {
		for _, hi := range []int{lo - 10, lo, lo + 5, lo + 10, lo + 95, lo + 100, 2000} {
			for threshold := lo - 20; threshold <= hi+20; threshold += 5 {
				buyer := func(price int) bool { return price <= threshold }
				seller := func(price int) bool { return price >= threshold }

				calls := 0
				counted := func(accept func(int) bool) func(int) bool {
					return func(price int) bool {
						calls++
						assert.GreaterOrEqual(t, price, lo)
						assert.LessOrEqual(t, price, hi)
						return accept(price)
					}
				}
				maxCalls := 0
				if hi >= lo {
					maxCalls = 2 + 2*bits.Len(uint(priceGridSize(lo, hi)))
				}

				wantPrice, wantFound := linearHighest(lo, hi, buyer)
				price, found := highestAcceptedPrice(lo, hi, counted(buyer))
				assert.Equal(t, wantFound, found, "highest lo=%d hi=%d threshold=%d", lo, hi, threshold)
				assert.Equal(t, wantPrice, price, "highest lo=%d hi=%d threshold=%d", lo, hi, threshold)
				assert.LessOrEqual(t, calls, maxCalls)

				calls = 0
				wantPrice, wantFound = linearLowest(lo, hi, seller)
				price, found = lowestAcceptedPrice(lo, hi, counted(seller))
				assert.Equal(t, wantFound, found, "lowest lo=%d hi=%d threshold=%d", lo, hi, threshold)
				assert.Equal(t, wantPrice, price, "lowest lo=%d hi=%d threshold=%d", lo, hi, threshold)
				assert.LessOrEqual(t, calls, maxCalls)
			}
		}
	}
}

func TestPriceSearchIncludesUpperBound(t *testing.T) {
	// 37 is not on the 10$ grid from 0, but it is searched as the last price
	price, found := highestAcceptedPrice(0, 37, func(int) bool { return true })
	assert.True(t, found)
	assert.Equal(t, 37, price)

	price, found = lowestAcceptedPrice(0, 37, func(p int) bool { return p > 30 })
	assert.True(t, found)
	assert.Equal(t, 37, price)
}

func TestGetSellCandidates(t *testing.T) {
	// property 3: Light Blue (set 3,4,5), price 100; property 0: Brown (set 0,1), price 60
	tests := []struct {
		name      string
		owned     []int
		mortgaged []int
		setup     func(state monopoly.GameState)
		sellList  []int
		expected  []sellCandidate
	}{
		{
			name:     "property in full set is skipped",
			owned:    []int{0, 1, 3},
			sellList: []int{0, 1, 3},
			expected: []sellCandidate{{propertyId: 3, minPrice: 50, maxPrice: 100, buyerId: 1}},
		},
		{
			name:      "mortgaged property is worth half its price",
			owned:     []int{3},
			mortgaged: []int{3},
			sellList:  []int{3},
			expected:  []sellCandidate{{propertyId: 3, minPrice: 10, maxPrice: 50, buyerId: 1}},
		},
		{
			name:     "buyer is the opponent closest to the full set",
			owned:    []int{3},
			setup:    func(s monopoly.GameState) { s.Properties[4].Owner = s.Players[2] },
			sellList: []int{3},
			expected: []sellCandidate{{propertyId: 3, minPrice: 50, maxPrice: 100, buyerId: 2}},
		},
		{
			name:     "tie goes to the richer opponent",
			owned:    []int{3},
			setup:    func(s monopoly.GameState) { s.Players[3].Money = 1200 },
			sellList: []int{3},
			expected: []sellCandidate{{propertyId: 3, minPrice: 50, maxPrice: 100, buyerId: 3}},
		},
		{
			name:  "maximum price is capped by the richest opponent",
			owned: []int{3},
			setup: func(s monopoly.GameState) {
				s.Players[1].Money, s.Players[2].Money, s.Players[3].Money = 70, 60, 20
			},
			sellList: []int{3},
			expected: []sellCandidate{{propertyId: 3, minPrice: 50, maxPrice: 70, buyerId: 1}},
		},
		{
			name:  "skipped when no opponent can pay the minimum price",
			owned: []int{3},
			setup: func(s monopoly.GameState) {
				s.Players[1].Money, s.Players[2].Money, s.Players[3].Money = 40, 40, 40
			},
			sellList: []int{3},
			expected: nil,
		},
		{
			name:  "bankrupt opponent is not a buyer",
			owned: []int{3},
			setup: func(s monopoly.GameState) {
				s.Properties[4].Owner = s.Players[2]
				s.Players[2].IsBankrupt = true
			},
			sellList: []int{3},
			expected: []sellCandidate{{propertyId: 3, minPrice: 50, maxPrice: 100, buyerId: 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := makeTestState(4)
			for _, id := range tt.owned {
				state.Properties[id].Owner = state.Players[0]
			}
			for _, id := range tt.mortgaged {
				state.Properties[id].IsMortgaged = true
			}
			if tt.setup != nil {
				tt.setup(state)
			}
			assert.Equal(t, tt.expected, getSellCandidates(0, state, tt.sellList))
		})
	}
}

func TestTrySellPropertyRequiresSellOffer(t *testing.T) {
	player := newTestNetworkPlayer(t)
	state := makeTestState(2)
	state.Properties[3].Owner = state.Players[0]
	available := monopoly.FullActionList{SellPropertyList: []int{3}} // SELLOFFER already used this turn
	decision, _, _ := player.trySellProperty(0, state, available)
	assert.False(t, decision)
}

func newTestNetworkPlayer(t *testing.T) *NEATMonopolyPlayer {
	t.Helper()
	organism, err := genetics.NewOrganism(0, newPerfGenome(1, perfHiddenNodes, rand.New(rand.NewSource(1))), 0)
	require.NoError(t, err)
	player, err := NewNEATMonopolyPlayer(organism)
	require.NoError(t, err)
	return player
}

func TestAuctionBidsMatchAskingAtEveryBid(t *testing.T) {
	for threshold := -10; threshold <= 2010; threshold += 5 {
		a := auctionBids{acceptedUpTo: -1, rejectedFrom: 2001, lookahead: 1}
		calls := 0
		accept := func(price int) bool {
			calls++
			return price <= threshold
		}
		var firstBidCalls int
		passed := false
		for bid := 10; bid <= 2000 && !passed; bid += 10 {
			got := a.accepts(bid, 2000, accept)
			if bid == 10 {
				firstBidCalls = calls
			}
			assert.Equal(t, bid <= threshold, got, "threshold=%d bid=%d", threshold, bid)
			passed = !got
		}
		assert.Equal(t, 1, firstBidCalls, "the first bid costs one activation")
		// a war of d steps: ~log2(d) probes ahead plus a few to settle the overshoot
		assert.LessOrEqual(t, calls, 2*bits.Len(uint(max(threshold, 10)/10))+2, "threshold=%d", threshold)
	}
}

func TestAuctionBidsRespectMaxPrice(t *testing.T) {
	a := auctionBids{acceptedUpTo: -1, rejectedFrom: 2001, lookahead: 8}
	var asked []int
	assert.True(t, a.accepts(100, 130, func(price int) bool {
		asked = append(asked, price)
		return true
	}))
	assert.Equal(t, []int{130}, asked, "the probe ahead is capped by the bidder's money")
}

func TestBiddingDecisionReusesAnswersWithinAuction(t *testing.T) {
	player := newTestNetworkPlayer(t)
	state := makeTestState(2)
	activations := 0
	decisionHook = func(*NEATMonopolyPlayer) { activations++ }
	defer func() { decisionHook = nil }()

	bid := func(auctionId int, price int, winner int) int {
		activations = 0
		state.AuctionId = auctionId
		player.BiddingDecision(0, state, 3, price, winner)
		return activations
	}

	assert.Equal(t, 1, bid(1, 10, -1), "first bid of an auction")
	total := 0
	for price := 10; price < 1000; price += 10 {
		total += bid(1, price, 1)
	}
	assert.LessOrEqual(t, total, 2*bits.Len(100)+2, "a whole bidding war in one auction")
	assert.Equal(t, 1, bid(2, 10, -1), "new auction")
	assert.Equal(t, 1, bid(0, 10, -1), "no auction id: every call activates")
	assert.Equal(t, 1, bid(0, 10, -1), "no auction id: every call activates")
}

func TestBiddingDecisionFirstBidIsStartingPrice(t *testing.T) {
	player := newTestNetworkPlayer(t)
	state := makeTestState(2)
	state.AuctionId = 1
	got := player.BiddingDecision(0, state, 3, 10, -1)
	assert.Contains(t, []int{0, 10}, got, "nobody has bid yet: the bot bids the starting price or passes")
	state.Players[0].Money = 15
	state.AuctionId = 2
	assert.Zero(t, player.BiddingDecision(0, state, 3, 10, 1), "cannot afford 20$")
}
