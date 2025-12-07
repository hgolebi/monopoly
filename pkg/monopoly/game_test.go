package monopoly

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var playerNames = []string{"player1", "player2", "player3", "player4"}

type MockMonopolyIO struct {
	mock.Mock
}

func (m *MockMonopolyIO) Init() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockMonopolyIO) GetStdAction(player int, state GameState, availableActions FullActionList) ActionDetails {
	args := m.Called(player, state, availableActions)
	return args.Get(0).(ActionDetails)
}

func (m *MockMonopolyIO) GetJailAction(player int, state GameState, available []JailAction) JailAction {
	args := m.Called(player, state, available)
	return args.Get(0).(JailAction)
}

func (m *MockMonopolyIO) BuyDecision(player int, state GameState, propertyId int) bool {
	args := m.Called(player, state, propertyId)
	return args.Bool(0)
}

func (m *MockMonopolyIO) BuyFromPlayerDecision(player int, state GameState, propertyId int, price int) bool {
	args := m.Called(player, state, propertyId, price)
	return args.Bool(0)
}

func (m *MockMonopolyIO) SellToPlayerDecision(player int, state GameState, propertyId int, price int) bool {
	args := m.Called(player, state, propertyId, price)
	return args.Bool(0)
}

// BiddingDecision(player int, state GameState, propertyId int, currentPrice int, currentWinner int) int

func (m *MockMonopolyIO) BiddingDecision(player int, state GameState, propertyId int, currentPrice int, currentWinner int) int {
	args := m.Called(player, state, propertyId, currentPrice, currentWinner)
	return args.Int(0)
}

func (m *MockMonopolyIO) Finish(f FinishOption, winner int, state GameState) {
	m.Called(f, winner, state)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Init() {
	m.Called()
}

func (m *MockLogger) Log(message string) {
	m.Called(message)
}

func (m *MockLogger) LogState(state GameState) {
	m.Called(state)
}

func (m *MockLogger) Error(msg string, state GameState) {
	m.Called(msg, state)
}

func (m *MockLogger) LogWithState(msg string, state GameState) {
	m.Called(msg, state)
}

func TestGetState(t *testing.T) {
	io := &MockMonopolyIO{}
	io.On("Init").Return(playerNames[:2])
	logger := &MockLogger{}
	logger.On("Init").Return()
	logger.On("Log", mock.Anything).Return()
	logger.On("LogState", mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	logger.On("LogWithState", mock.Anything, mock.Anything).Return()
	game := NewGame(context.Background(), io, logger, 0)
	state := game.getState()

	assert.NotNil(t, state, "State should not be nil")
	assert.Equal(t, state.CurrentPlayerIdx, 0, "CurrentPlayerIdx should be 0 at the start of the game")
	assert.Equal(t, len(state.Players), 2, "Players list should contain 2 players at the start of the game")
	assert.Equal(t, state.Round, 1, "Round should be 1 at the start of the game")
	assert.NotNil(t, state.Properties, "Properties should not be nil")
	assert.Equal(t, len(state.Properties), 28, "There should be 28 properties at the start of the game")
	assert.Equal(t, state.Charge, 0, "Charge should be 0 at the start of the game")

}

func TestGetActivePlayers(t *testing.T) {
	tests := []struct {
		playersCount    int
		bankruptPlayers []int
		expectedOutput  []int
	}{
		{2, []int{}, []int{0, 1}},
		{2, []int{0}, []int{1}},
		{2, []int{1}, []int{0}},
		{2, []int{0, 1}, []int{}},
		{4, []int{}, []int{0, 1, 2, 3}},
		{4, []int{0, 1}, []int{2, 3}},
		{4, []int{2, 3}, []int{0, 1}},
		{4, []int{0, 2}, []int{1, 3}},
		{4, []int{1, 3}, []int{0, 2}},
		{4, []int{0, 1, 2}, []int{3}},
		{4, []int{1, 2, 3}, []int{0}},
		{4, []int{0, 1, 3}, []int{2}},
		{4, []int{0, 1, 2, 3}, []int{}},
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:test.playersCount])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 0)
		for _, playerId := range test.bankruptPlayers {
			game.players[playerId].IsBankrupt = true
		}

		activePlayers := game.getActivePlayers()
		assert.Equal(t, len(activePlayers), len(test.expectedOutput), "Number of active players should match expected output")
		for i, playerId := range activePlayers {
			assert.Equal(t, playerId, test.expectedOutput[i], "Active player ID should match expected output")
		}
	}
}

func TestGetCurrPlayer(t *testing.T) {
	tests := []struct {
		playersCount       int
		currentPlayer      int
		expectedPlayerId   int
		expectedPlayerName string
	}{
		{2, 0, 0, "player1"},
		{2, 1, 1, "player2"},
		{4, 0, 0, "player1"},
		{4, 1, 1, "player2"},
		{4, 2, 2, "player3"},
		{4, 3, 3, "player4"},
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:test.playersCount])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 0)
		game.currentPlayerIdx = test.currentPlayer

		currPlayer := game.getCurrPlayer()
		assert.Equal(t, currPlayer, game.players[test.currentPlayer], "Current player should match the expected player")
		assert.Equal(t, currPlayer.ID, test.expectedPlayerId, "Current player ID should match expected")
		assert.Equal(t, currPlayer.Name, test.expectedPlayerName, "Current player name should match expected")
	}
}

// func TestCheckForWinner(t *testing.T) {
// 	tests := []struct {
// 		playersCount         int
// 		bankrupts            []int
// 		roundCount           int
// 		expectedFinish       bool
// 		expectedFinishOption FinishOption
// 		expectedWinnerID     int
// 	}{
// 		{2, []int{}, 10, false, 0, -1},
// 		{2, []int{0}, 10, true, WIN, 1},
// 		{2, []int{1}, 10, true, WIN, 0},
// 		{2, []int{0, 1}, 10, true, DRAW, -1},
// 		{2, []int{}, 50, false, 0, -1},
// 		{2, []int{}, 51, true, ROUND_LIMIT, 0},
// 		{4, []int{}, 10, false, 0, -1},
// 		{4, []int{0}, 10, false, 0, -1},
// 		{4, []int{0, 1, 2}, 10, true, WIN, 3},
// 		{4, []int{0, 1, 2, 3}, 51, true, DRAW, -1},
// 	}

// 	for _, test := range tests {
// 		io := &MockMonopolyIO{}
// 		io.On("Init", test.playersCount).Return()
// 		io.On("Finish", mock.Anything, mock.Anything, mock.Anything).Return()
// 		logger := &MockLogger{}
// 		logger.On("Init").Return()
// 		logger.On("Log", mock.Anything).Return()
// 		logger.On("LogState", mock.Anything).Return()

// 		game := NewGame(context.Background(), io, logger, 0)
// 		for _, playerId := range test.bankrupts {
// 			game.players[playerId].IsBankrupt = true
// 		}
// 		game.round = test.roundCount

// 		gameFinished := game.checkForWinner()
// 		assert.Equal(t, gameFinished, test.expectedFinish, "Game finish status should match expected")
// 		if !test.expectedFinish {
// 			io.AssertNotCalled(t, "Finish", mock.Anything, mock.Anything, mock.Anything)
// 		} else {
// 			io.AssertCalled(t, "Finish", test.expectedFinishOption, test.expectedWinnerID, mock.Anything)
// 		}
// 	}
// }

func TestJailPlayer(t *testing.T) {
	tests := []struct {
		playerId int
	}{
		{0},
		{1},
		{2},
		{3},
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 0)
		game.currentPlayerIdx = test.playerId
		game.jailPlayer()

		assert.True(t, game.players[test.playerId].IsJailed, "Player should be jailed")
		assert.Equal(t, game.players[test.playerId].RoundsInJail, 0, "Rounds in jail should be 1")
		assert.Equal(t, game.players[test.playerId].CurrentPosition, 10, "Player's position should be set to jail")
	}
}

func TestMovePlayer(t *testing.T) {
	tests := []struct {
		playerId         int
		startingPosition int
		numberOfFields   int
		expectedPosition int
		expectedCash     int
	}{
		{0, 0, 2, 2, 1500},
		{1, 0, 2, 2, 1500},
		{2, 0, 2, 2, 1500},
		{3, 0, 2, 2, 1500},
		{0, 2, 4, 6, 1500},
		{0, 39, 1, 0, 1700},
		{0, 38, 3, 1, 1700},
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 0)
		game.currentPlayerIdx = test.playerId
		game.players[test.playerId].CurrentPosition = test.startingPosition

		game.movePlayer(test.numberOfFields)

		assert.Equal(t, game.players[test.playerId].CurrentPosition, test.expectedPosition, "Player's position should match expected")
		assert.Equal(t, game.players[test.playerId].Money, test.expectedCash, "Player's cash should match expected")
	}
}

func TestRollDiceBetween1And6(t *testing.T) {
	io := &MockMonopolyIO{}
	io.On("Init").Return(playerNames[:4])
	logger := &MockLogger{}
	logger.On("Init").Return()
	logger.On("Log", mock.Anything).Return()
	logger.On("LogState", mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	logger.On("LogWithState", mock.Anything, mock.Anything).Return()

	g := NewGame(context.Background(), io, logger, 0)

	for range 1000 {
		d1, d2 := g.rollDice()
		if d1 < 1 || d1 > 6 {
			t.Errorf("dice1 out of range: got %d", d1)
		}
		if d2 < 1 || d2 > 6 {
			t.Errorf("dice2 out of range: got %d", d2)
		}
	}
}

func TestSeededRollDice(t *testing.T) {
	tests := []struct {
		seed int64
		d1   int
		d2   int
	}{
		{1, 6, 4},
		{2, 5, 1},
		{22, 4, 4},
		{42, 6, 6},
		{56, 1, 1},
	}
	for _, test := range tests {
		for i := 0; i < 100; i++ {
			io := &MockMonopolyIO{}
			io.On("Init").Return(playerNames[:4])
			logger := &MockLogger{}
			logger.On("Init").Return()
			logger.On("Log", mock.Anything).Return()
			logger.On("LogState", mock.Anything).Return()
			logger.On("Error", mock.Anything, mock.Anything).Return()
			logger.On("LogWithState", mock.Anything, mock.Anything).Return()

			game := NewGame(context.Background(), io, logger, test.seed)
			d1, d2 := game.rollDice()
			assert.Equal(t, test.d1, d1, "First die should match expected")
			assert.Equal(t, test.d2, d2, "Second die should match expected")
		}
	}
}

// func TestFindSeed(t *testing.T) {
// 	for i := 0; i < 100; i++ {
// 		io := &MockMonopolyIO{}
// 		io.On("Init", 4).Return()
// 		logger := &MockLogger{}
// 		logger.On("Init").Return()
// 		logger.On("Log", mock.Anything).Return()
// 		logger.On("LogState", mock.Anything).Return()

// 		game := NewGame(4, io, logger, int64(i))
// 		d1, d2 := game.rollDice()
// 		assert.NotEqual(t, d1, d2, fmt.Sprintf("Seed: %d", i))
// 	}
// }

func TestRollDiceAllValuesPossible(t *testing.T) {
	io := &MockMonopolyIO{}
	io.On("Init").Return(playerNames[:4])
	logger := &MockLogger{}
	logger.On("Init").Return()
	logger.On("Log", mock.Anything).Return()
	logger.On("LogState", mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	logger.On("LogWithState", mock.Anything, mock.Anything).Return()

	g := NewGame(context.Background(), io, logger, 0)

	countsD1 := make([]int, 6)
	countsD2 := make([]int, 6)

	for range 10000 {
		d1, d2 := g.rollDice()
		countsD1[d1-1]++
		countsD2[d2-1]++
	}
	for i, c := range countsD1 {
		if c == 0 {
			t.Errorf("Value %d never rolled on dice1", i+1)
		}
	}
	for i, c := range countsD2 {
		if c == 0 {
			t.Errorf("Value %d never rolled on dice2", i+1)
		}
	}
}

func TestHandleJail(t *testing.T) {
	tests := []struct {
		playerId         int
		jailCards        int
		roundsInJail     int
		availableActions []JailAction
	}{
		{0, 0, 0, []JailAction{BAIL, ROLL_DICE}},
		{0, 0, 1, []JailAction{BAIL, ROLL_DICE}},
		{0, 0, 2, []JailAction{BAIL, ROLL_DICE}},
		{0, 0, 3, []JailAction{BAIL}},
		{0, 1, 1, []JailAction{BAIL, CARD, ROLL_DICE}},
		{0, 1, 3, []JailAction{BAIL, CARD}},
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		io.On("BuyDecision", mock.Anything, mock.Anything, mock.Anything).Return(true)

		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		io.On("GetStdAction", test.playerId, mock.Anything, mock.Anything).Return(ActionDetails{
			Action: NOACTION,
		})
		io.On("GetJailAction", test.playerId, mock.Anything, mock.Anything).Return(ROLL_DICE)
		game := NewGame(context.Background(), io, logger, 0)
		game.currentPlayerIdx = test.playerId
		player := game.players[test.playerId]
		player.IsJailed = true
		player.JailCards = test.jailCards
		player.RoundsInJail = test.roundsInJail

		game.handleJail()
		io.AssertCalled(t, "GetJailAction", test.playerId, mock.Anything, test.availableActions)

	}
}

func TestJailRollDice(t *testing.T) {
	tests := []struct {
		playerId             int
		roundsInJail         int
		seed                 int64
		expectedPosition     int
		expectedRoundsInJail int
		expectedJailed       bool
	}{
		{1, 0, 1, 10, 1, true},
		{1, 1, 1, 10, 2, true},
		{1, 2, 1, 10, 3, true},
		{0, 0, 56, 20, 0, false},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		io.On("BuyDecision", mock.Anything, mock.Anything, mock.Anything).Return(true)
		io.On("GetStdAction", test.playerId, mock.Anything, mock.Anything).Return(ActionDetails{
			Action: NOACTION,
		})
		game := NewGame(context.Background(), io, logger, test.seed)
		game.currentPlayerIdx = test.playerId
		player := game.players[test.playerId]
		player.IsJailed = true
		player.CurrentPosition = 10
		player.RoundsInJail = test.roundsInJail

		game.jailRollDice()
		assert.Equal(t, test.expectedPosition, player.CurrentPosition, "Player's position should match expected after rolling dice in jail")
		assert.Equal(t, test.expectedRoundsInJail, player.RoundsInJail, "Player's rounds in jail should match expected after rolling dice in jail")
		assert.Equal(t, test.expectedJailed, player.IsJailed, "Player's jailed status should match expected after rolling dice in jail")
	}
}

func TestJailBail(t *testing.T) {
	tests := []struct {
		playerId             int
		roundsInJail         int
		cash                 int
		expectedCash         int
		expectedJailed       bool
		expectedBankrupt     bool
		expectedRoundsInJail int
	}{
		{1, 0, 1500, 1450, false, false, 0},
		{1, 1, 1500, 1450, false, false, 0},
		{1, 3, 1500, 1450, false, false, 0},
		{0, 0, 50, 0, false, false, 0},
		{0, 0, 49, -1, true, true, 0},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		io.On("BuyDecision", mock.Anything, mock.Anything, mock.Anything).Return(true)
		io.On("GetStdAction", test.playerId, mock.Anything, mock.Anything).Return(ActionDetails{
			Action: NOACTION,
		})
		game := NewGame(context.Background(), io, logger, 1)
		game.currentPlayerIdx = test.playerId
		player := game.players[test.playerId]
		player.IsJailed = true
		player.CurrentPosition = 10
		player.Money = test.cash
		player.RoundsInJail = test.roundsInJail

		game.jailBail()
		assert.Equal(t, test.expectedCash, player.Money, "Player's cash should match expected after bailing out of jail")
		assert.Equal(t, test.expectedJailed, player.IsJailed, "Player's jailed status should match expected after bailing out of jail")
		assert.Equal(t, test.expectedBankrupt, player.IsBankrupt, "Player's bankrupt status should match expected after bailing out of jail")
		assert.Equal(t, test.expectedRoundsInJail, player.RoundsInJail, "Player's rounds in jail should match expected after rolling dice in jail")
	}
}

func TestJailCard(t *testing.T) {
	tests := []struct {
		playerId             int
		roundsInJail         int
		cards                int
		expectedCards        int
		expectedJailed       bool
		expectedRoundsInJail int
	}{
		{1, 0, 1, 0, false, 0},
		{1, 1, 2, 1, false, 0},
		{1, 3, 3, 2, false, 0},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		io.On("BuyDecision", mock.Anything, mock.Anything, mock.Anything).Return(true)
		io.On("GetStdAction", test.playerId, mock.Anything, mock.Anything).Return(ActionDetails{
			Action: NOACTION,
		})
		game := NewGame(context.Background(), io, logger, 1)
		game.currentPlayerIdx = test.playerId
		player := game.players[test.playerId]
		player.IsJailed = true
		player.CurrentPosition = 10
		player.JailCards = test.cards
		player.RoundsInJail = test.roundsInJail

		game.jailCard()
		assert.Equal(t, test.expectedCards, player.JailCards, "Player's jail cards should match expected after using a jail card")
		assert.Equal(t, test.expectedJailed, player.IsJailed, "Player's jailed status should match expected after using a jail card")
		assert.Equal(t, test.expectedRoundsInJail, player.RoundsInJail, "Player's rounds in jail should match expected after using a jail card")
	}
}

func TestCheckHouses(t *testing.T) {
	tests := []struct {
		propertyId     int
		canBuildHouses bool
		Houses         int
		expectedResult bool
	}{
		{1, false, 0, false},
		{2, true, 0, false},
		{3, true, 1, true},
		{4, true, 2, true},
		{5, true, 3, true},
		{6, true, 4, true},
		{7, true, 5, true},
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 1)
		property := game.properties[test.propertyId]
		property.CanBuildHouse = test.canBuildHouses
		property.Houses = test.Houses

		result := game.checkHouses(game.properties[test.propertyId])
		assert.Equal(t, test.expectedResult, result, "Check houses result should match expected")
	}
}

// "Brown":      {0, 1},
// "Light Blue": {3, 4, 5},
// "Pink":       {6, 8, 9},
// "Orange":     {11, 12, 13},
// "Red":        {14, 15, 16},
// "Yellow":     {18, 19, 21},
// "Green":      {22, 23, 24},
// "Dark Blue":  {26, 27},
func TestCheckHousesSets(t *testing.T) {
	tests := []struct {
		propertyIds       []int
		propertyWithHouse int
		expectedResult    bool
	}{
		{[]int{0, 1}, 0, true},        // Brown
		{[]int{0, 1}, 1, true},        // Brown
		{[]int{0, 1}, 2, false},       // Brown
		{[]int{3, 4, 5}, 3, true},     // Light Blue
		{[]int{3, 4, 5}, 4, true},     // Light Blue
		{[]int{3, 4, 5}, 5, true},     // Light Blue
		{[]int{6, 8, 9}, 6, true},     // Pink
		{[]int{6, 8, 9}, 8, true},     // Pink
		{[]int{6, 8, 9}, 7, false},    // Pink
		{[]int{6, 8, 9}, 9, true},     // Pink
		{[]int{11, 12, 13}, 11, true}, // Orange
		{[]int{11, 12, 13}, 12, true}, // Orange
		{[]int{11, 12, 13}, 13, true}, // Orange
		{[]int{14, 15, 16}, 14, true}, // Red
		{[]int{14, 15, 16}, 15, true}, // Red
		{[]int{14, 15, 16}, 16, true}, // Red
		{[]int{18, 19, 21}, 18, true}, // Yellow
		{[]int{18, 19, 21}, 19, true}, // Yellow
		{[]int{18, 19, 21}, 21, true}, // Yellow
		{[]int{22, 23, 24}, 22, true}, // Green
		{[]int{22, 23, 24}, 23, true}, // Green
		{[]int{22, 23, 24}, 24, true}, // Green
		{[]int{26, 27}, 26, true},     // Dark Blue
		{[]int{26, 27}, 27, true},     // Dark Blue
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 1)
		game.properties[test.propertyWithHouse].Houses = 1
		for _, propId := range test.propertyIds {
			property := game.properties[propId]
			result := game.checkHouses(property)
			assert.Equal(t, test.expectedResult, result, "Check houses sets result should match expected")
		}

	}
}

func TestGetMortgageList(t *testing.T) {
	tests := []struct {
		playerId             int
		playerProperties     []int
		mortgagedProperties  []int
		propertiesWithHouses []int
		expectedMortgageList []int
	}{
		{
			playerId:             1,
			playerProperties:     []int{},
			mortgagedProperties:  []int{},
			propertiesWithHouses: []int{},
			expectedMortgageList: []int{},
		},
		{
			playerId:             1,
			playerProperties:     []int{4, 5},
			mortgagedProperties:  []int{},
			propertiesWithHouses: []int{},
			expectedMortgageList: []int{4, 5},
		},
		{
			playerId:             0,
			playerProperties:     []int{1, 2, 3},
			mortgagedProperties:  []int{2},
			propertiesWithHouses: []int{},
			expectedMortgageList: []int{1, 3},
		},
		{
			playerId:             0,
			playerProperties:     []int{1, 2, 3},
			mortgagedProperties:  []int{},
			propertiesWithHouses: []int{1},
			expectedMortgageList: []int{2, 3},
		},
		{
			playerId:             0,
			playerProperties:     []int{1, 2, 3},
			mortgagedProperties:  []int{1},
			propertiesWithHouses: []int{3},
			expectedMortgageList: []int{2},
		},
		{
			playerId:             0,
			playerProperties:     []int{0, 1},
			mortgagedProperties:  []int{},
			propertiesWithHouses: []int{1},
			expectedMortgageList: []int{},
		},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 1)
		game.currentPlayerIdx = test.playerId
		player := game.players[test.playerId]
		player.Properties = test.playerProperties
		for _, propId := range test.mortgagedProperties {
			property := game.properties[propId]
			property.IsMortgaged = true
		}

		for _, propId := range test.propertiesWithHouses {
			property := game.properties[propId]
			property.Houses = rand.Intn(5) + 1
		}

		mortgageList := game.getMortgageList(test.playerId)
		assert.ElementsMatch(t, test.expectedMortgageList, mortgageList, "Mortgage list should match expected")
	}
}

func TestGetBuyOutList(t *testing.T) {
	tests := []struct {
		playerId            int
		playerProperties    []int
		mortgagedProperties []int
		expectedBuyOutList  []int
	}{
		{
			playerId:            1,
			playerProperties:    []int{},
			mortgagedProperties: []int{},
			expectedBuyOutList:  []int{},
		},
		{
			playerId:            3,
			playerProperties:    []int{4, 5},
			mortgagedProperties: []int{},
			expectedBuyOutList:  []int{},
		},
		{
			playerId:            0,
			playerProperties:    []int{1, 2, 3},
			mortgagedProperties: []int{2},
			expectedBuyOutList:  []int{2},
		},
		{
			playerId:            2,
			playerProperties:    []int{1, 2, 3},
			mortgagedProperties: []int{1, 2, 3},
			expectedBuyOutList:  []int{1, 2, 3},
		},
		{
			playerId:            3,
			playerProperties:    []int{4, 5},
			mortgagedProperties: []int{6, 7},
			expectedBuyOutList:  []int{},
		},
		{
			playerId:            3,
			playerProperties:    []int{4, 5},
			mortgagedProperties: []int{4, 6, 7},
			expectedBuyOutList:  []int{4},
		},
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 1)
		game.currentPlayerIdx = test.playerId
		player := game.players[test.playerId]
		player.Properties = test.playerProperties
		for _, propId := range test.mortgagedProperties {
			property := game.properties[propId]
			property.IsMortgaged = true
		}
		buyOutList := game.getBuyOutList(test.playerId)
		assert.ElementsMatch(t, test.expectedBuyOutList, buyOutList, "Buy out list should match expected")
	}
}

func TestGetSellPropertyList(t *testing.T) {
	tests := []struct {
		playerId             int
		playerProperties     []int
		mortgagedProperties  []int
		propertiesWithHouses []int
		expectedSellList     []int
	}{
		{
			playerId:             1,
			playerProperties:     []int{},
			mortgagedProperties:  []int{},
			propertiesWithHouses: []int{},
			expectedSellList:     []int{},
		},
		{
			playerId:             3,
			playerProperties:     []int{4, 5},
			mortgagedProperties:  []int{},
			propertiesWithHouses: []int{},
			expectedSellList:     []int{4, 5},
		},
		{
			playerId:             0,
			playerProperties:     []int{1, 2, 3},
			mortgagedProperties:  []int{2},
			propertiesWithHouses: []int{},
			expectedSellList:     []int{1, 2, 3},
		},
		{
			playerId:             2,
			playerProperties:     []int{1, 2, 3},
			mortgagedProperties:  []int{1, 2, 3},
			propertiesWithHouses: []int{},
			expectedSellList:     []int{1, 2, 3},
		},
		{
			playerId:             3,
			playerProperties:     []int{4, 5},
			mortgagedProperties:  []int{6, 7},
			propertiesWithHouses: []int{},
			expectedSellList:     []int{4, 5},
		},
		{
			playerId:             3,
			playerProperties:     []int{1, 20},
			mortgagedProperties:  []int{},
			propertiesWithHouses: []int{1},
			expectedSellList:     []int{20},
		},
		{
			playerId:             3,
			playerProperties:     []int{4, 5},
			mortgagedProperties:  []int{},
			propertiesWithHouses: []int{4},
			expectedSellList:     []int{},
		},
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 1)
		game.currentPlayerIdx = test.playerId
		player := game.players[test.playerId]
		player.Properties = test.playerProperties
		for _, propId := range test.mortgagedProperties {
			property := game.properties[propId]
			property.IsMortgaged = true
		}
		for _, propId := range test.propertiesWithHouses {
			property := game.properties[propId]
			property.Houses = rand.Intn(5) + 1
		}
		sellList := game.getSellPropertyList(test.playerId)
		assert.ElementsMatch(t, test.expectedSellList, sellList, "Sell list should match expected")
	}
}

func TestBuyPropertyList(t *testing.T) {
	tests := []struct {
		playerId             int
		player0Properties    []int
		player1Properties    []int
		player2Properties    []int
		player3Properties    []int
		propertiesWithHouses []int
		expectedBuyList      []int
	}{
		{
			playerId:             0,
			player0Properties:    []int{},
			player1Properties:    []int{},
			player2Properties:    []int{},
			player3Properties:    []int{},
			propertiesWithHouses: []int{},
			expectedBuyList:      []int{},
		},
		{
			playerId:             0,
			player0Properties:    []int{1, 2, 3},
			player1Properties:    []int{},
			player2Properties:    []int{},
			player3Properties:    []int{},
			propertiesWithHouses: []int{},
			expectedBuyList:      []int{},
		},
		{
			playerId:             0,
			player0Properties:    []int{1, 2, 3},
			player1Properties:    []int{},
			player2Properties:    []int{},
			player3Properties:    []int{6, 7},
			propertiesWithHouses: []int{},
			expectedBuyList:      []int{6, 7},
		},
		{
			playerId:             0,
			player0Properties:    []int{1, 2, 3},
			player1Properties:    []int{4, 5},
			player2Properties:    []int{6, 7},
			player3Properties:    []int{8, 9},
			propertiesWithHouses: []int{},
			expectedBuyList:      []int{4, 5, 6, 7, 8, 9},
		},
		{
			playerId:             0,
			player0Properties:    []int{1, 2},
			player1Properties:    []int{3, 4, 5},
			player2Properties:    []int{6, 7},
			player3Properties:    []int{8, 9},
			propertiesWithHouses: []int{5},
			expectedBuyList:      []int{6, 7, 8, 9},
		},
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 1)
		game.currentPlayerIdx = test.playerId

		game.players[0].Properties = test.player0Properties
		game.players[1].Properties = test.player1Properties
		game.players[2].Properties = test.player2Properties
		game.players[3].Properties = test.player3Properties
		for _, propId := range test.player0Properties {
			game.properties[propId].Owner = game.players[0]
		}
		for _, propId := range test.player1Properties {
			game.properties[propId].Owner = game.players[1]
		}
		for _, propId := range test.player2Properties {
			game.properties[propId].Owner = game.players[2]
		}
		for _, propId := range test.player3Properties {
			game.properties[propId].Owner = game.players[3]
		}
		for _, propId := range test.propertiesWithHouses {
			property := game.properties[propId]
			property.Houses = rand.Intn(5) + 1
		}
		buyList := game.getBuyPropertyList(test.playerId)
		assert.ElementsMatch(t, test.expectedBuyList, buyList, "Buy property list should match expected")
	}
}

func TestBankruptTransferingProperties(t *testing.T) {
	tests := []struct {
		players                  []string
		bankruptPlayer           int
		playerProperties         []int
		targetPlayer             int
		targetPlayerProperties   []int
		expectedTargetProperties []int
	}{
		{
			players:                  []string{"player1", "player2"},
			bankruptPlayer:           0,
			playerProperties:         []int{1, 2, 3},
			targetPlayer:             1,
			targetPlayerProperties:   []int{4, 5},
			expectedTargetProperties: []int{1, 2, 3, 4, 5},
		},
		// {
		// 	players:                 []string{"player1", "player2", "player3", "player4"},
		// 	bankruptPlayer:          3,
		// 	playerProperties:        []int{1, 2, 3},
		// 	targetPlayer:            1,
		// 	targetPlayerProperties:  []int{4, 5},
		// 	expectedTargetProperties: []int{1, 2, 3, 4, 5},
		// },
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(test.players)
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 0)
		state := game.getState()

		bankruptPlayer := state.Players[test.bankruptPlayer]
		targetPlayer := state.Players[test.targetPlayer]
		game.players[test.bankruptPlayer].Properties = test.playerProperties
		for _, propId := range test.playerProperties {
			game.properties[propId].Owner = bankruptPlayer
		}
		game.players[test.targetPlayer].Properties = test.targetPlayerProperties
		for _, propId := range test.targetPlayerProperties {
			game.properties[propId].Owner = targetPlayer
		}
		game.bankrupt(bankruptPlayer, targetPlayer)
		assert.Equal(t, true, bankruptPlayer.IsBankrupt, "Bankrupt player should be marked as bankrupt")
		assert.ElementsMatch(t, test.expectedTargetProperties, game.players[test.targetPlayer].Properties, "Target player's properties should match expected after transfer")
	}
}

func TestChargePlayer(t *testing.T) {
	tests := []struct {
		playerId         int
		initialCash      int
		chargeAmount     int
		expectedCash     int
		expectedBankrupt bool
	}{
		{0, 500, 200, 300, false},
		{1, 100, 150, -1, true},
		{2, 300, 300, 0, false},
		{3, 0, 50, -1, true},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])

		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 0)
		player := game.players[test.playerId]
		player.Money = test.initialCash
		game.chargePlayer(test.playerId, test.chargeAmount, nil)
		assert.Equal(t, test.expectedCash, player.Money, "Player's cash should match expected after charge")
		assert.Equal(t, test.expectedBankrupt, player.IsBankrupt, "Player's bankrupt status should match expected after charge")
	}
}

func TestMortgage(t *testing.T) {
	tests := []struct {
		playerId     int
		initialCash  int
		propertyId   int
		expectedCash int
	}{
		{0, 500, 1, 530},
		{1, 100, 3, 150},
		{2, 300, 5, 360},
		{3, 0, 8, 70},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()

		game := NewGame(context.Background(), io, logger, 0)
		player := game.players[test.playerId]
		player.Money = test.initialCash
		player.Properties = []int{test.propertyId}
		property := game.properties[test.propertyId]
		property.Owner = player
		game.mortgage(test.playerId, test.propertyId)
		assert.Equal(t, test.expectedCash, player.Money, "Player's cash should match expected after mortgage")
		assert.Equal(t, true, property.IsMortgaged, "Property should be marked as mortgaged after mortgage")
	}
}

func TestSellHouse(t *testing.T) {
	tests := []struct {
		playerId       int
		initialCash    int
		initialHouses  int
		propertyId     int
		expectedCash   int
		expectedHouses int
	}{
		{0, 500, 2, 1, 525, 1},
		{1, 100, 1, 3, 125, 0},
		{2, 300, 3, 5, 325, 2},
		{3, 0, 4, 8, 50, 3},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		player := game.players[test.playerId]
		player.Money = test.initialCash
		player.Properties = []int{test.propertyId}
		property := game.properties[test.propertyId]
		property.Owner = player
		property.Houses = test.initialHouses
		game.sellHouse(test.playerId, test.propertyId)
		assert.Equal(t, test.expectedCash, player.Money, "Player's cash should match expected after selling house")
		assert.Equal(t, test.expectedHouses, property.Houses, "Property's houses should match expected after selling house")
	}
}

func TestBuyHouse(t *testing.T) {
	tests := []struct {
		playerId       int
		initialCash    int
		initialHouses  int
		propertyId     int
		expectedCash   int
		expectedHouses int
	}{
		{0, 500, 1, 1, 450, 2},
		{1, 200, 0, 3, 150, 1},
		{2, 300, 2, 5, 250, 3},
		{3, 100, 3, 8, 0, 4},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		player := game.players[test.playerId]
		player.Money = test.initialCash
		player.Properties = []int{test.propertyId}
		property := game.properties[test.propertyId]
		property.Owner = player
		property.CanBuildHouse = true
		property.Houses = test.initialHouses
		game.buyHouse(test.playerId, test.propertyId)
		assert.Equal(t, test.expectedCash, player.Money, "Player's cash should match expected after buying house")
		assert.Equal(t, test.expectedHouses, property.Houses, "Property's houses should match expected after buying house")
	}
}

func TestSendSellOffer(t *testing.T) {
	tests := []struct {
		playerId      int
		targetPlayers []int
		propertyId    int
		price         int
	}{
		{0, []int{1, 2}, 3, 200},
		{1, []int{0, 3}, 5, 300},
		{2, []int{1}, 8, 150},
		{3, []int{0, 2}, 11, 400},
		{1, []int{}, 8, 150},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		//BuyFromPlayerDecision(player int, state GameState, propertyId int, price int) bool
		io.On("BuyFromPlayerDecision", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)

		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		player := game.players[test.playerId]
		player.Properties = []int{test.propertyId}
		property := game.properties[test.propertyId]
		property.Owner = player
		game.sendSellOffer(test.playerId, test.targetPlayers, test.propertyId, test.price)
		for _, targetId := range test.targetPlayers {
			io.AssertCalled(t, "BuyFromPlayerDecision", targetId, mock.Anything, test.propertyId, test.price)
		}
	}
}

func TestSendBuyOffer(t *testing.T) {
	tests := []struct {
		playerId     int
		targetPlayer int
		propertyId   int
		price        int
	}{
		{0, 1, 3, 200},
		{1, 2, 5, 300},
		{2, 3, 8, 150},
		{3, 1, 11, 400},
		{1, 0, 8, 150},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		//SellToPlayerDecision(player int, state GameState, propertyId int, price int) bool
		io.On("SellToPlayerDecision", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)

		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)

		target := game.players[test.targetPlayer]
		target.Properties = []int{test.propertyId}
		property := game.properties[test.propertyId]
		property.Owner = target
		game.sendBuyOffer(test.playerId, test.propertyId, test.price)

		io.AssertCalled(t, "SellToPlayerDecision", test.targetPlayer, mock.Anything, test.propertyId, test.price)

	}
}

func TestBuyOut(t *testing.T) {
	tests := []struct {
		playerId     int
		propertyId   int
		cash         int
		expectedCash int
	}{
		{0, 1, 500, 434},
		{1, 3, 300, 190},
		{2, 5, 400, 268},
		{3, 8, 600, 446},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])

		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		player := game.players[test.playerId]
		player.Money = test.cash
		property := game.properties[test.propertyId]
		property.IsMortgaged = true
		property.Owner = player
		game.buyOut(test.playerId, test.propertyId)
		assert.Equal(t, test.expectedCash, player.Money, "Player's cash should match expected after buying out property")
		assert.Equal(t, false, property.IsMortgaged, "Property should not be mortgaged after buy out")
	}
}
func TestDoForProperty(t *testing.T) {
	tests := []struct {
		playerId           int
		cash               int
		propertyId         int
		houses             int
		propertyOwner      int
		ownerCash          int
		expectedBuyOption  bool
		expectedPlayerCash int
		expectedOwnerCash  int
	}{
		{0, 500, 1, 0, -1, 0, true, 500, 0},
		{1, 300, 3, 0, 2, 400, false, 294, 406},
		{2, 200, 5, 2, 3, 100, false, 192, 108},
		{2, 200, 5, 1, 2, 200, false, 200, 200},
		{3, 600, 8, 0, -1, 0, true, 600, 0},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		io.On("BuyDecision", test.playerId, mock.Anything, mock.Anything).Return(false)
		io.On("BiddingDecision", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(0)

		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		game.currentPlayerIdx = test.playerId
		player := game.players[test.playerId]
		player.Money = test.cash
		property := game.properties[test.propertyId]
		property.Houses = test.houses
		if test.propertyOwner >= 0 {
			owner := game.players[test.propertyOwner]
			property.Owner = owner
			owner.Money = test.ownerCash
		}
		game.doForProperty(property)
		if test.propertyOwner < 0 {
			io.AssertCalled(t, "BuyDecision", test.playerId, mock.Anything, test.propertyId)
		} else {
			assert.Equal(t, test.expectedPlayerCash, player.Money, "Player's cash should match expected after doForProperty")
			assert.Equal(t, test.expectedOwnerCash, property.Owner.Money, "Owner's cash should match expected after doForProperty")

		}
	}
}

func TestDoForTaxField(t *testing.T) {
	tests := []struct {
		playerId     int
		cash         int
		tax          int
		expectedCash int
	}{
		{0, 500, 200, 300},
		{1, 150, 100, 50},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		game.currentPlayerIdx = test.playerId
		player := game.players[test.playerId]
		player.Money = test.cash
		taxField := &TaxField{
			Tax: test.tax,
		}
		game.doForTaxField(taxField)
		assert.Equal(t, test.expectedCash, player.Money, "Player's cash should match expected after doForTaxField")
	}
}

func TestDoForGoToJailField(t *testing.T) {
	tests := []struct {
		playerId         int
		playerCash       int
		expectedPosition int
		expectedIsJailed bool
		expectedCash     int
	}{
		{0, 500, 10, true, 500},
		{1, 0, 10, true, 0},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		game.currentPlayerIdx = test.playerId
		player := game.players[test.playerId]
		player.Money = test.playerCash
		player.CurrentPosition = 30

		game.doForGoToJailField()
		assert.Equal(t, test.expectedPosition, player.CurrentPosition, "Player's position should match expected after doForGoToJailField")
		assert.Equal(t, test.expectedIsJailed, player.IsJailed, "Player's jailed status should match expected after doForGoToJailField")
		assert.Equal(t, test.expectedCash, player.Money, "Player's cash should match expected after doForGoToJailField")
	}
}

func TestGameAddProperty(t *testing.T) {
	tests := []struct {
		player     int
		propertyId int
	}{
		{0, 1},
		{1, 3},
		{2, 5},
		{3, 8},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		player := game.players[test.player]
		game.addProperty(player, test.propertyId)
		assert.Contains(t, player.Properties, test.propertyId, "Player's properties should contain the added property")
	}
}

func TestCharge(t *testing.T) {
	tests := []struct {
		player             int
		playerCash         int
		target             int
		targetCash         int
		amount             int
		expectedCash       int
		expectedTargetCash int
	}{
		{0, 500, -1, 0, 200, 300, 0},
		{1, 150, 2, 400, 100, 50, 500},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		player := game.players[test.player]
		player.Money = test.playerCash
		var targetPlayer *Player
		if test.target >= 0 {
			targetPlayer = game.players[test.target]
			targetPlayer.Money = test.targetCash
		}
		game.charge(player, test.amount, targetPlayer)
		assert.Equal(t, test.expectedCash, player.Money, "Player's cash should match expected after charge")
		if test.target >= 0 {
			assert.Equal(t, test.expectedTargetCash, targetPlayer.Money, "Target player's cash should match expected after charge")
		}
	}
}

func TestTransferProperty(t *testing.T) {
	tests := []struct {
		fromPlayer    int
		toPlayer      int
		propertyId    int
		expectedOwner int
	}{
		{0, 1, 1, 1},
		{1, 2, 3, 2},
		{2, 3, 5, 3},
		{3, 0, 8, 0},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		fromPlayer := game.players[test.fromPlayer]
		toPlayer := game.players[test.toPlayer]
		fromPlayer.Properties = []int{test.propertyId}
		property := game.properties[test.propertyId]
		property.Owner = fromPlayer
		game.transferProperty(fromPlayer, toPlayer, test.propertyId)
		assert.Equal(t, test.expectedOwner, property.Owner.ID, "Property's owner should match expected after transfer")
	}
}

func TestAddMoneyGame(t *testing.T) {
	tests := []struct {
		playerId     int
		initialCash  int
		addAmount    int
		expectedCash int
	}{
		{0, 500, 200, 700},
		{1, 100, 150, 250},
		{2, 300, 300, 600},
		{3, 0, 50, 50},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		player := game.players[test.playerId]
		player.Money = test.initialCash
		game.addMoney(player, test.addAmount)
		assert.Equal(t, test.expectedCash, player.Money, "Player's cash should match expected after adding money")
	}
}

func TestSetPosition(t *testing.T) {
	tests := []struct {
		playerId         int
		initialPosition  int
		newPosition      int
		expectedPosition int
	}{
		{0, 5, 10, 10},
		{1, 15, 20, 20},
		{2, 25, 30, 30},
		{3, 35, 0, 0},
	}
	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init").Return(playerNames[:4])
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()
		logger.On("Error", mock.Anything, mock.Anything).Return()
		logger.On("LogWithState", mock.Anything, mock.Anything).Return()
		game := NewGame(context.Background(), io, logger, 0)
		player := game.players[test.playerId]
		player.CurrentPosition = test.initialPosition
		game.setPosition(player, test.newPosition)
		assert.Equal(t, test.expectedPosition, player.CurrentPosition, "Player's position should match expected after setting position")
	}
}
