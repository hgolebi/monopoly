package monopoly

import (
	"testing"

	cfg "monopoly/pkg/config"

	"github.com/stretchr/testify/assert"
)

func largestSetSize(sets [][]int) int {
	m := 0
	for _, set := range sets {
		m = max(m, len(set))
	}
	return m
}

func TestMaxSetSizeMatchesSets(t *testing.T) {
	assert.Equal(t, largestSetSize(Sets), cfg.MAX_SET_SIZE, "cfg.MAX_SET_SIZE must equal the size of the largest set in Sets")
}
