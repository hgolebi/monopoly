package neatnetwork

import (
	"monopoly/pkg/monopoly"
	"testing"

	"github.com/stretchr/testify/assert"
)

// makeTestState creates a GameState with numPlayers players (each starting with 1000 money)
// and all 28 properties unowned, matching the property setup from game.go.
func makeTestState(numPlayers int) monopoly.GameState {
	names := []string{"Alice", "Bob", "Charlie", "Dave"}
	players := make([]*monopoly.Player, numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = monopoly.NewPlayer(i, names[i], 1000)
	}

	properties := []*monopoly.Property{
		monopoly.NewProperty(1, 0, "Brown1", 60, 50, true, 0),
		monopoly.NewProperty(3, 1, "Brown2", 60, 50, true, 0),
		monopoly.NewProperty(5, 2, "Railroad1", 200, 0, false, 1),
		monopoly.NewProperty(6, 3, "LightBlue1", 100, 50, true, 2),
		monopoly.NewProperty(8, 4, "LightBlue2", 100, 50, true, 2),
		monopoly.NewProperty(9, 5, "LightBlue3", 120, 50, true, 2),
		monopoly.NewProperty(11, 6, "Pink1", 140, 100, true, 3),
		monopoly.NewProperty(12, 7, "Utility1", 150, 0, false, 4),
		monopoly.NewProperty(13, 8, "Pink2", 140, 100, true, 3),
		monopoly.NewProperty(14, 9, "Pink3", 160, 100, true, 3),
		monopoly.NewProperty(15, 10, "Railroad2", 200, 0, false, 1),
		monopoly.NewProperty(16, 11, "Orange1", 180, 100, true, 5),
		monopoly.NewProperty(18, 12, "Orange2", 180, 100, true, 5),
		monopoly.NewProperty(19, 13, "Orange3", 200, 100, true, 5),
		monopoly.NewProperty(21, 14, "Red1", 220, 150, true, 6),
		monopoly.NewProperty(23, 15, "Red2", 220, 150, true, 6),
		monopoly.NewProperty(24, 16, "Red3", 240, 150, true, 6),
		monopoly.NewProperty(25, 17, "Railroad3", 200, 0, false, 1),
		monopoly.NewProperty(26, 18, "Yellow1", 260, 150, true, 7),
		monopoly.NewProperty(27, 19, "Yellow2", 260, 150, true, 7),
		monopoly.NewProperty(28, 20, "Utility2", 150, 0, false, 4),
		monopoly.NewProperty(29, 21, "Yellow3", 280, 150, true, 7),
		monopoly.NewProperty(31, 22, "Green1", 300, 200, true, 8),
		monopoly.NewProperty(32, 23, "Green2", 300, 200, true, 8),
		monopoly.NewProperty(34, 24, "Green3", 320, 200, true, 8),
		monopoly.NewProperty(35, 25, "Railroad4", 200, 0, false, 1),
		monopoly.NewProperty(37, 26, "DarkBlue1", 350, 200, true, 9),
		monopoly.NewProperty(39, 27, "DarkBlue2", 400, 200, true, 9),
	}

	return monopoly.GameState{
		Players:    players,
		Properties: properties,
	}
}

func TestNewMonopolySensors(t *testing.T) {
	ms := NewMonopolySensors()
	assert.Equal(t, int(INPUT_COUNT), len(ms))
}

func TestLoadState_PlayerInputs(t *testing.T) {
	tests := []struct {
		isJailed         bool
		position         int
		money            int
		expectedIsJailed float64
		expectedPosition float64
		expectedMoney    float64
	}{
		{false, 0, 0, 0.0, 0.0, 0.0},
		{true, 39, 2000, 1.0, 1.0, 1.0},
		{false, 19, 1000, 0.0, 19.0 / 39.0, 0.5},
		{true, 10, 500, 1.0, 10.0 / 39.0, 0.25},
	}
	for _, tt := range tests {
		state := makeTestState(1)
		state.Players[0].IsJailed = tt.isJailed
		state.Players[0].CurrentPosition = tt.position
		state.Players[0].Money = tt.money

		ms := NewMonopolySensors()
		ms.LoadState(state, 0, 0)

		assert.InDelta(t, tt.expectedIsJailed, ms[IN_PLAYER_IS_JAILED], 0.0001)
		assert.InDelta(t, tt.expectedPosition, ms[IN_PLAYER_POSITION], 0.0001)
		assert.InDelta(t, tt.expectedMoney, ms[IN_PLAYER_MONEY], 0.0001)
	}
}

func TestLoadState_PropertyType(t *testing.T) {
	state := makeTestState(1)
	tests := []struct {
		propertyID   int
		expectedType float64
	}{
		{0, 1.0},  // Brown — street
		{2, 0.0},  // Railroad1
		{7, 0.5},  // Utility1
		{20, 0.5}, // Utility2
		{27, 1.0}, // DarkBlue — street
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		ms.LoadState(state, 0, tt.propertyID)
		assert.InDelta(t, tt.expectedType, ms[IN_PROPERTY_TYPE], 0.0001, "propertyID=%d", tt.propertyID)
	}
}

func TestLoadState_PropertyID(t *testing.T) {
	state := makeTestState(1)
	tests := []struct {
		propertyID int
		expected   float64
	}{
		{0, 0.0},
		{27, 1.0},
		{13, 13.0 / 27.0},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		ms.LoadState(state, 0, tt.propertyID)
		assert.InDelta(t, tt.expected, ms[IN_PROPERTY_ID], 0.0001, "propertyID=%d", tt.propertyID)
	}
}

func TestLoadState_PropertyPrice_Unmortgaged(t *testing.T) {
	state := makeTestState(1)
	ms := NewMonopolySensors()
	ms.LoadState(state, 0, 0) // Brown1, price=60
	assert.InDelta(t, 60.0/2000.0, ms[IN_PROPERTY_PRICE], 0.0001)
}

func TestLoadState_PropertyPrice_Mortgaged(t *testing.T) {
	state := makeTestState(1)
	state.Properties[0].IsMortgaged = true
	// Effective price = 60 + GetBuyoutPrice() = 60 + int(60*0.55) = 60 + 33 = 93
	ms := NewMonopolySensors()
	ms.LoadState(state, 0, 0)
	assert.InDelta(t, 93.0/2000.0, ms[IN_PROPERTY_PRICE], 0.0001)
}

func TestLoadState_Round(t *testing.T) {
	tests := []struct {
		round    int
		expected float64
	}{
		{0, 0.0},
		{25, 0.5},
		{50, 1.0},
	}
	for _, tt := range tests {
		state := makeTestState(1)
		state.Round = tt.round
		ms := NewMonopolySensors()
		ms.LoadState(state, 0, 0)
		assert.InDelta(t, tt.expected, ms[IN_ROUND], 0.0001, "round=%d", tt.round)
	}
}

func TestLoadState_AvailableProperties(t *testing.T) {
	state := makeTestState(2)
	state.Properties[0].Owner = state.Players[0]
	state.Properties[1].Owner = state.Players[1]

	ms := NewMonopolySensors()
	ms.LoadState(state, 0, 0)
	// 28 total, 2 owned → 26 available; normalized by 28
	assert.InDelta(t, 26.0/28.0, ms[IN_AVAILABLE_PROPERTIES], 0.0001)
}

func TestLoadState_EnemyNotLoadedByDefault(t *testing.T) {
	state := makeTestState(1)
	ms := NewMonopolySensors()
	ms.LoadState(state, 0, 0)
	// LoadState does not call LoadEnemyInputs — enemy inputs stay 0
	assert.InDelta(t, 0.0, ms[IN_ENEMY_INVOLVED], 0.0001)
	assert.InDelta(t, 0.0, ms[IN_ENEMY_MONEY], 0.0001)
}

func TestLoadSetInputs_Brown(t *testing.T) {
	state := makeTestState(2)

	// No one owns anything — player needs all 2 Brown properties
	ms := NewMonopolySensors()
	ms.loadSetInputs(state, 0, 0) // property 0 = Brown, setId=0
	assert.InDelta(t, 0.0, ms[IN_SET_ID], 0.0001)
	assert.InDelta(t, 2.0/9.0, ms[IN_SET_PROPERTIES_NEEDED], 0.0001)
	assert.InDelta(t, 0.0, ms[IN_SET_PROPERTIES_OCCUPIED], 0.0001)

	// Player 0 owns Brown1 — needs 1 more
	state.Properties[0].Owner = state.Players[0]
	ms2 := NewMonopolySensors()
	ms2.loadSetInputs(state, 0, 0)
	assert.InDelta(t, 1.0/9.0, ms2[IN_SET_PROPERTIES_NEEDED], 0.0001)
	assert.InDelta(t, 0.0, ms2[IN_SET_PROPERTIES_OCCUPIED], 0.0001)

	// Player 1 also owns Brown2 — enemy blocks the remaining property
	state.Properties[1].Owner = state.Players[1]
	ms3 := NewMonopolySensors()
	ms3.loadSetInputs(state, 0, 0)
	assert.InDelta(t, 1.0/9.0, ms3[IN_SET_PROPERTIES_NEEDED], 0.0001)
	assert.InDelta(t, 1.0/9.0, ms3[IN_SET_PROPERTIES_OCCUPIED], 0.0001)
}

func TestLoadSetInputs_Railroad(t *testing.T) {
	state := makeTestState(1)
	ms := NewMonopolySensors()
	ms.loadSetInputs(state, 2, 0) // Railroad1, setId=1
	assert.InDelta(t, 1.0/9.0, ms[IN_SET_ID], 0.0001)
	assert.InDelta(t, 4.0/9.0, ms[IN_SET_PROPERTIES_NEEDED], 0.0001) // needs all 4
	assert.InDelta(t, 0.0, ms[IN_SET_PROPERTIES_OCCUPIED], 0.0001)
}

func TestLoadEnemyInputs(t *testing.T) {
	state := makeTestState(2)
	state.Players[1].Money = 500
	state.Properties[1].Owner = state.Players[1] // enemy owns Brown2

	ms := NewMonopolySensors()
	ms.LoadEnemyInputs(state, 1, 0) // enemy=player1, property 0 (Brown set)

	assert.InDelta(t, 1.0, ms[IN_ENEMY_INVOLVED], 0.0001)
	assert.InDelta(t, 500.0/2000.0, ms[IN_ENEMY_MONEY], 0.0001)
	// Enemy owns 1 of 2 Brown — needs 1 more
	assert.InDelta(t, 1.0/9.0, ms[IN_ENEMY_SET_PROPERTIES_NEEDED], 0.0001)
	// No other player owns Brown properties
	assert.InDelta(t, 0.0, ms[IN_ENEMY_SET_PROPERTIES_OCCUPIED], 0.0001)
}

func TestLoadEnemyInputs_OtherPlayerBlocks(t *testing.T) {
	state := makeTestState(2)
	state.Properties[0].Owner = state.Players[0] // player 0 owns Brown1

	ms := NewMonopolySensors()
	ms.LoadEnemyInputs(state, 1, 0) // query from enemy=player1's perspective

	// Enemy (player1) owns 0 Brown — needs 2
	assert.InDelta(t, 2.0/9.0, ms[IN_ENEMY_SET_PROPERTIES_NEEDED], 0.0001)
	// Player 0 counts as "others" from enemy's perspective
	assert.InDelta(t, 1.0/9.0, ms[IN_ENEMY_SET_PROPERTIES_OCCUPIED], 0.0001)
}

func TestLoadBuyPriceInput(t *testing.T) {
	tests := []struct {
		price    int
		expected float64
	}{
		{0, 0.0},
		{500, 0.25},
		{1000, 0.5},
		{2000, 1.0},
		{3000, 1.0}, // clipped to max
		{10, 0.005},
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		ms.LoadBuyPriceInput(tt.price)
		assert.InDelta(t, tt.expected, ms[IN_BUY_PRICE], 0.0001, "price=%d", tt.price)
		assert.InDelta(t, 0.0, ms[IN_SELL_PRICE], 0.0001, "sell price must be unaffected")
	}
}

func TestLoadSellPriceInput(t *testing.T) {
	tests := []struct {
		price    int
		expected float64
	}{
		{0, 0.0},
		{500, 0.25},
		{1000, 0.5},
		{2000, 1.0},
		{3000, 1.0}, // clipped to max
	}
	for _, tt := range tests {
		ms := NewMonopolySensors()
		ms.LoadSellPriceInput(tt.price)
		assert.InDelta(t, tt.expected, ms[IN_SELL_PRICE], 0.0001, "price=%d", tt.price)
		assert.InDelta(t, 0.0, ms[IN_BUY_PRICE], 0.0001, "buy price must be unaffected")
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		value    int
		min      int
		max      int
		shift    bool
		expected float64
	}{
		{0, 0, 100, false, 0.0},
		{50, 0, 100, false, 0.5},
		{100, 0, 100, false, 1.0},
		{150, 0, 100, false, 1.0},  // clipped to 1
		{-10, 0, 100, false, 0.0},  // clipped to 0
		{5, 5, 5, false, 0.0},      // max == min
		{1, 0, 3, true, 2.0 / 4.0}, // (1-0+1)/(3-0+1) = 2/4
		{0, 0, 39, false, 0.0},
		{39, 0, 39, false, 1.0},
		{2000, 0, 2000, false, 1.0},
	}
	for _, tt := range tests {
		result := normalize(tt.value, tt.min, tt.max, tt.shift)
		assert.InDelta(t, tt.expected, result, 0.0001,
			"normalize(%d, %d, %d, %v)", tt.value, tt.min, tt.max, tt.shift)
	}
}

func TestActivation(t *testing.T) {
	// f(x) = x * (1 - ((1-x)/0.8)^2)
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},       // 0 * anything = 0
		{1.0, 1.0},       // 1 * (1 - 0) = 1
		{0.8, 0.75},      // 0.8 * (1 - 0.0625) = 0.75
		{0.5, 0.3046875}, // 0.5 * (1 - 0.390625)
	}
	for _, tt := range tests {
		result := activation(tt.input)
		assert.InDelta(t, tt.expected, result, 0.0001, "activation(%f)", tt.input)
	}
}
