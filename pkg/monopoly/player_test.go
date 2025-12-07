package monopoly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPlayer(t *testing.T) {
	var tests = []struct {
		id    int
		name  string
		money int
	}{
		{1, "Alice", 1500},
		{2, "Bob", 1500},
		{3, "Charlie", 1500},
		{4, "Diana", 1500},
	}

	for _, tt := range tests {
		player := NewPlayer(tt.id, tt.name, tt.money)
		assert.Equal(t, tt.id, player.ID)
		assert.Equal(t, tt.name, player.Name)
		assert.Equal(t, tt.money, player.Money)
	}
}

func TestNewPlayerFail(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for negative initial money, but got none")
		}
	}()
	NewPlayer(1, "Eve", -500)
}

func TestAddMoney(t *testing.T) {
	var tests = []struct {
		initialMoney  int
		amountToAdd   int
		expectedMoney int
	}{
		{1000, 500, 1500},
		{2000, 300, 2300},
		{0, 100, 100},
	}
	for _, tt := range tests {
		player := NewPlayer(1, "TestPlayer", tt.initialMoney)
		player.AddMoney(tt.amountToAdd)
		assert.Equal(t, tt.expectedMoney, player.Money)
	}
}

func TestAddMoneyFail(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for adding negative amount, but got none")
		}
	}()
	player := NewPlayer(1, "TestPlayer", 1000)
	player.AddMoney(-500)
}

func TestRemoveMoney(t *testing.T) {
	var tests = []struct {
		initialMoney   int
		amountToRemove int
		expectedMoney  int
	}{
		{1500, 500, 1000},
		{2000, 300, 1700},
		{100, 100, 0},
	}
	for _, tt := range tests {
		player := NewPlayer(1, "TestPlayer", tt.initialMoney)
		player.RemoveMoney(tt.amountToRemove)
		assert.Equal(t, tt.expectedMoney, player.Money)
	}
}

func TestRemoveMoneyFail(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for removing negative amount, but got none")
		}
	}()
	player := NewPlayer(1, "TestPlayer", 1000)
	player.RemoveMoney(-500)
}
func TestAddProperty(t *testing.T) {
	var tests = []struct {
		initialProperties  []int
		propertyToAdd      int
		expectedProperties []int
	}{
		{[]int{1, 2}, 3, []int{1, 2, 3}},
		{[]int{}, 1, []int{1}},
		{[]int{4, 5, 6}, 7, []int{4, 5, 6, 7}},
	}
	for _, tt := range tests {
		player := NewPlayer(1, "TestPlayer", 1000)
		player.Properties = append(player.Properties, tt.initialProperties...)
		player.AddProperty(tt.propertyToAdd)
		assert.Equal(t, tt.expectedProperties, player.Properties)
	}
}

func TestAddPropertyFail(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for adding already owned property, but got none")
		}
	}()
	player := NewPlayer(1, "TestPlayer", 1000)
	player.AddProperty(1)
	player.AddProperty(1) // Attempt to add the same property again
}

func TestRemoveProperty(t *testing.T) {
	var tests = []struct {
		initialProperties  []int
		propertyToRemove   int
		expectedProperties []int
	}{
		{[]int{1, 2, 3}, 2, []int{1, 3}},
		{[]int{1}, 1, []int{}},
		{[]int{4, 5, 6, 7}, 5, []int{4, 6, 7}},
	}
	for _, tt := range tests {
		player := NewPlayer(1, "TestPlayer", 1000)
		player.Properties = append(player.Properties, tt.initialProperties...)
		player.RemoveProperty(tt.propertyToRemove)
		assert.Equal(t, tt.expectedProperties, player.Properties)
	}
}
func TestRemovePropertyFail(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for removing unowned property, but got none")
		}
	}()
	player := NewPlayer(1, "TestPlayer", 1000)
	player.AddProperty(1)
	player.RemoveProperty(2) // Attempt to remove a property not owned
}
