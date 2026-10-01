package neatnetwork

import (
	"monopoly/pkg/monopoly"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetChargeAction(t *testing.T) {
	tests := []struct {
		name          string
		owned         []int
		mortgageList  []int
		sellHouseList []int
		expected      monopoly.ActionDetails
	}{
		{
			name:         "property outside full set before full set",
			owned:        []int{0, 1, 4},
			mortgageList: []int{0, 1, 4},
			expected:     monopoly.ActionDetails{Action: monopoly.MORTGAGE, PropertyId: 4},
		},
		{
			name:         "lowest id among properties outside full sets",
			owned:        []int{4, 12},
			mortgageList: []int{12, 4},
			expected:     monopoly.ActionDetails{Action: monopoly.MORTGAGE, PropertyId: 4},
		},
		{
			name:         "incomplete railroads before full set",
			owned:        []int{0, 1, 2, 10},
			mortgageList: []int{0, 1, 2, 10},
			expected:     monopoly.ActionDetails{Action: monopoly.MORTGAGE, PropertyId: 2},
		},
		{
			name:         "lowest id among full sets",
			owned:        []int{0, 1, 26, 27},
			mortgageList: []int{27, 26, 1, 0},
			expected:     monopoly.ActionDetails{Action: monopoly.MORTGAGE, PropertyId: 0},
		},
		{
			name:          "mortgage before selling houses",
			owned:         []int{0, 1, 26, 27},
			mortgageList:  []int{26, 27},
			sellHouseList: []int{0, 1},
			expected:      monopoly.ActionDetails{Action: monopoly.MORTGAGE, PropertyId: 26},
		},
		{
			name:          "sell house from lowest id",
			owned:         []int{0, 1, 22, 23, 24},
			sellHouseList: []int{24, 22, 23, 1, 0},
			expected:      monopoly.ActionDetails{Action: monopoly.SELLHOUSE, PropertyId: 0},
		},
		{
			name:     "nothing available",
			expected: monopoly.ActionDetails{Action: monopoly.NOACTION},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := makeTestState(4)
			for _, id := range tt.owned {
				state.Properties[id].Owner = state.Players[0]
			}
			available := monopoly.FullActionList{MortgageList: tt.mortgageList, SellHouseList: tt.sellHouseList}
			assert.Equal(t, tt.expected, getChargeAction(0, state, available))
		})
	}
}

// TestGetChargeAction_FullLiquidation repeats the engine's liquidation loop until nothing is left
// and checks that houses are sold set by set, each set being mortgaged before the next one is touched.
func TestGetChargeAction_FullLiquidation(t *testing.T) {
	state := makeTestState(4)
	owned := []int{0, 1, 4, 11, 12, 13}
	for _, id := range owned {
		state.Properties[id].Owner = state.Players[0]
	}
	state.Properties[0].Houses = 1
	state.Properties[1].Houses = 2
	state.Properties[11].Houses = 1
	state.Properties[12].Houses = 1
	state.Properties[13].Houses = 1

	setHasHouses := func(setIdx int) bool {
		for _, id := range monopoly.Sets[setIdx] {
			if state.Properties[id].Houses > 0 {
				return true
			}
		}
		return false
	}

	var sequence []monopoly.ActionDetails
	for range 50 {
		available := monopoly.FullActionList{}
		for _, id := range owned {
			property := state.Properties[id]
			if !property.IsMortgaged && !setHasHouses(property.SetIndex) {
				available.MortgageList = append(available.MortgageList, id)
			}
			if property.Houses > 0 {
				available.SellHouseList = append(available.SellHouseList, id)
			}
		}
		action := getChargeAction(0, state, available)
		if action.Action == monopoly.NOACTION {
			break
		}
		sequence = append(sequence, action)
		switch action.Action {
		case monopoly.MORTGAGE:
			state.Properties[action.PropertyId].IsMortgaged = true
		case monopoly.SELLHOUSE:
			state.Properties[action.PropertyId].Houses--
		}
	}

	mortgage := func(id int) monopoly.ActionDetails {
		return monopoly.ActionDetails{Action: monopoly.MORTGAGE, PropertyId: id}
	}
	sellHouse := func(id int) monopoly.ActionDetails {
		return monopoly.ActionDetails{Action: monopoly.SELLHOUSE, PropertyId: id}
	}
	expected := []monopoly.ActionDetails{
		mortgage(4),                              // Light Blue: not a full set
		sellHouse(0), sellHouse(1), sellHouse(1), // Brown houses
		mortgage(0), mortgage(1), // Brown mortgaged before touching Orange
		sellHouse(11), sellHouse(12), sellHouse(13),
		mortgage(11), mortgage(12), mortgage(13),
	}
	assert.Equal(t, expected, sequence)
}
