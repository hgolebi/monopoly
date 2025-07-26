package monopoly

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMonopolyIO struct {
	mock.Mock
}

func (m *MockMonopolyIO) Init(players int) {
	m.Called(players)
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

func (m *MockMonopolyIO) BiddingDecision(player int, state GameState, propertyId int, currentPrice int) int {
	args := m.Called(player, state, propertyId, currentPrice)
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

func TestGetState(t *testing.T) {
	io := &MockMonopolyIO{}
	io.On("Init", 2).Return()
	logger := &MockLogger{}
	logger.On("Init").Return()
	logger.On("Log", mock.Anything).Return()
	logger.On("LogState", mock.Anything).Return()
	game := NewGame(2, io, logger)
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
		io.On("Init", test.playersCount).Return()
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()

		game := NewGame(test.playersCount, io, logger)
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
		io.On("Init", test.playersCount).Return()
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()

		game := NewGame(test.playersCount, io, logger)
		game.currentPlayerIdx = test.currentPlayer

		currPlayer := game.getCurrPlayer()
		assert.Equal(t, currPlayer, game.players[test.currentPlayer], "Current player should match the expected player")
		assert.Equal(t, currPlayer.ID, test.expectedPlayerId, "Current player ID should match expected")
		assert.Equal(t, currPlayer.Name, test.expectedPlayerName, "Current player name should match expected")
	}
}

func TestCheckForWinner(t *testing.T) {
	tests := []struct {
		playersCount         int
		bankrupts            []int
		roundCount           int
		expectedFinish       bool
		expectedFinishOption FinishOption
		expectedWinnerID     int
	}{
		{2, []int{}, 10, false, 0, -1},
		{2, []int{0}, 10, true, WIN, 1},
		{2, []int{1}, 10, true, WIN, 0},
		{2, []int{0, 1}, 10, true, DRAW, -1},
		{2, []int{}, 50, false, 0, -1},
		{2, []int{}, 51, true, ROUND_LIMIT, 0},
		{4, []int{}, 10, false, 0, -1},
		{4, []int{0}, 10, false, 0, -1},
		{4, []int{0, 1, 2}, 10, true, WIN, 3},
		{4, []int{0, 1, 2, 3}, 51, true, DRAW, -1},
	}

	for _, test := range tests {
		io := &MockMonopolyIO{}
		io.On("Init", test.playersCount).Return()
		io.On("Finish", mock.Anything, mock.Anything, mock.Anything).Return()
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()

		game := NewGame(test.playersCount, io, logger)
		for _, playerId := range test.bankrupts {
			game.players[playerId].IsBankrupt = true
		}
		game.round = test.roundCount

		gameFinished := game.checkForWinner()
		assert.Equal(t, gameFinished, test.expectedFinish, "Game finish status should match expected")
		if !test.expectedFinish {
			io.AssertNotCalled(t, "Finish", mock.Anything, mock.Anything, mock.Anything)
		} else {
			io.AssertCalled(t, "Finish", test.expectedFinishOption, test.expectedWinnerID, mock.Anything)
		}
	}
}

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
		io.On("Init", 4).Return()
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()

		game := NewGame(4, io, logger)
		game.currentPlayerIdx = test.playerId
		game.jailPlayer()

		assert.True(t, game.players[test.playerId].IsJailed, "Player should be jailed")
		assert.Equal(t, game.players[test.playerId].roundsInJail, 0, "Rounds in jail should be 1")
		assert.Equal(t, game.players[test.playerId].CurrentPosition, game.settings.JAIL_POSITION, "Player's position should be set to jail")
	}
}
