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
	game := NewGame(2, io, logger, 0)
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

		game := NewGame(test.playersCount, io, logger, 0)
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

		game := NewGame(test.playersCount, io, logger, 0)
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

		game := NewGame(test.playersCount, io, logger, 0)
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

func TestMakeMove(t *testing.T) {
	panic("not implemented")
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

		game := NewGame(4, io, logger, 0)
		game.currentPlayerIdx = test.playerId
		game.jailPlayer()

		assert.True(t, game.players[test.playerId].IsJailed, "Player should be jailed")
		assert.Equal(t, game.players[test.playerId].RoundsInJail, 0, "Rounds in jail should be 1")
		assert.Equal(t, game.players[test.playerId].CurrentPosition, game.settings.JAIL_POSITION, "Player's position should be set to jail")
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
		io.On("Init", 4).Return()
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()

		game := NewGame(4, io, logger, 0)
		game.currentPlayerIdx = test.playerId
		game.players[test.playerId].CurrentPosition = test.startingPosition

		game.movePlayer(test.numberOfFields)

		assert.Equal(t, game.players[test.playerId].CurrentPosition, test.expectedPosition, "Player's position should match expected")
		assert.Equal(t, game.players[test.playerId].Money, test.expectedCash, "Player's cash should match expected")
	}
}

func TestRollDiceBetween1And6(t *testing.T) {
	io := &MockMonopolyIO{}
	io.On("Init", 4).Return()
	logger := &MockLogger{}
	logger.On("Init").Return()
	logger.On("Log", mock.Anything).Return()
	logger.On("LogState", mock.Anything).Return()

	g := NewGame(4, io, logger, 0)

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
			io.On("Init", 4).Return()
			logger := &MockLogger{}
			logger.On("Init").Return()
			logger.On("Log", mock.Anything).Return()
			logger.On("LogState", mock.Anything).Return()

			game := NewGame(4, io, logger, test.seed)
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
	io.On("Init", 4).Return()
	logger := &MockLogger{}
	logger.On("Init").Return()
	logger.On("Log", mock.Anything).Return()
	logger.On("LogState", mock.Anything).Return()

	g := NewGame(4, io, logger, 0)

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
		io.On("Init", 4).Return()
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()

		io.On("GetStdAction", test.playerId, mock.Anything, mock.Anything).Return(ActionDetails{
			Action: NOACTION,
		})
		io.On("GetJailAction", test.playerId, mock.Anything, mock.Anything).Return(ROLL_DICE)
		game := NewGame(4, io, logger, 0)
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
		io.On("Init", 4).Return()
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()

		io.On("BuyDecision", mock.Anything, mock.Anything, mock.Anything).Return(true)
		io.On("GetStdAction", test.playerId, mock.Anything, mock.Anything).Return(ActionDetails{
			Action: NOACTION,
		})
		game := NewGame(4, io, logger, test.seed)
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
		io.On("Init", 4).Return()
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()

		io.On("BuyDecision", mock.Anything, mock.Anything, mock.Anything).Return(true)
		io.On("GetStdAction", test.playerId, mock.Anything, mock.Anything).Return(ActionDetails{
			Action: NOACTION,
		})
		game := NewGame(4, io, logger, 1)
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
		io.On("Init", 4).Return()
		logger := &MockLogger{}
		logger.On("Init").Return()
		logger.On("Log", mock.Anything).Return()
		logger.On("LogState", mock.Anything).Return()

		io.On("BuyDecision", mock.Anything, mock.Anything, mock.Anything).Return(true)
		io.On("GetStdAction", test.playerId, mock.Anything, mock.Anything).Return(ActionDetails{
			Action: NOACTION,
		})
		game := NewGame(4, io, logger, 1)
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
