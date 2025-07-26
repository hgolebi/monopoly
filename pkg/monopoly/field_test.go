package monopoly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProperty(t *testing.T) {
	property := NewProperty(1, 1, "Test Property", 100, 50, true, "Test Set")
	assert.Equal(t, property.FieldIndex, 1, "FieldIndex should be 1")
	assert.Equal(t, property.PropertyIndex, 1, "PropertyIndex should be 1")
	assert.Equal(t, property.Name, "Test Property", "Name should be 'Test Property'")
	assert.Equal(t, property.Price, 100, "Price should be 100")
	assert.Equal(t, property.HousePrice, 50, "HousePrice should be 50")
	assert.Equal(t, property.CanBuildHouse, true, "CanBuildHouse should be true")
	assert.Equal(t, property.Set, "Test Set", "Set should be 'Test Set'")
}

func TestNewPropertyFail(t *testing.T) {
	var tests = []struct {
		fieldID    int
		propertyID int
		name       string
		price      int
		housePrice int
		canBuild   bool
	}{
		{1, 1, "Test Property", -100, 50, true},
		{1, 2, "Test Property 2", 200, -100, false},
		{2, 1, "Test Property 3", -300, -150, true},
	}
	for _, test := range tests {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for property with price %d and house price %d, but got none", test.price, test.housePrice)
			}
		}()
		NewProperty(test.fieldID, test.propertyID, test.name, test.price, test.housePrice, test.canBuild, "Test Set")
	}

}
