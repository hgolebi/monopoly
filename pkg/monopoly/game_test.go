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
