package neatnetwork

import (
	cfg "monopoly/pkg/config"
	"monopoly/pkg/monopoly"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newScoringGroup(t *testing.T) (*NEATPlayerGroup, []*NEATMonopolyPlayer) {
	t.Helper()
	bots := make([]*NEATMonopolyPlayer, 4)
	players := make([]MonopolyPlayer, 4)
	for i := range bots {
		bots[i] = &NEATMonopolyPlayer{stats: &playerStats{}}
		players[i] = bots[i]
	}
	group, err := NewNEATPlayerGroup(0, players)
	require.NoError(t, err)
	return group, bots
}

// finishState builds a final game state; houses and properties are set to check that they are not scored.
func finishState(bankrupt ...int) monopoly.GameState {
	state := makeTestState(4)
	for _, p := range state.Players {
		p.Properties = []int{0, 1}
	}
	state.Properties[0].Houses = 5
	for _, id := range bankrupt {
		state.Players[id].IsBankrupt = true
	}
	return state
}

func TestFinish_WinScoresOnlyWinner(t *testing.T) {
	group, bots := newScoringGroup(t)

	group.Finish(monopoly.WIN, 2, finishState(0, 1, 3))

	for i, b := range bots {
		expected := 0
		if i == 2 {
			expected = cfg.FIRST_PLACE_SCORE
		}
		assert.Equal(t, expected, b.GetScore(), "score of player %d", i)
		assert.Equal(t, 0, b.GetDraws(), "draws of player %d", i)
	}
	assert.Equal(t, 1, bots[2].GetWins())
	assert.Equal(t, 0, bots[0].GetWins())
}

func TestFinish_RoundLimitScoresRichestPlayer(t *testing.T) {
	group, bots := newScoringGroup(t)

	// winner by net worth is chosen by the engine (endRoundLimit)
	group.Finish(monopoly.ROUND_LIMIT, 1, finishState(3))

	for i, b := range bots {
		expected := 0
		if i == 1 {
			expected = cfg.DRAW_WINNER_SCORE
		}
		assert.Equal(t, expected, b.GetScore(), "score of player %d", i)
		assert.Equal(t, 0, b.GetWins(), "round limit is not counted as a win, player %d", i)
	}
	assert.Equal(t, []int{1, 1, 1, 0}, []int{bots[0].GetDraws(), bots[1].GetDraws(), bots[2].GetDraws(), bots[3].GetDraws()},
		"players alive at the round limit get a draw")
}

func TestFinish_DrawGivesNoPoints(t *testing.T) {
	group, bots := newScoringGroup(t)

	group.Finish(monopoly.DRAW, -1, finishState(0, 1, 2, 3))

	for i, b := range bots {
		assert.Equal(t, 0, b.GetScore(), "score of player %d", i)
		assert.Equal(t, 0, b.GetWins(), "wins of player %d", i)
		assert.Equal(t, 0, b.GetDraws(), "draws of player %d", i)
	}
}

func TestFinish_ScoredOnlyOnce(t *testing.T) {
	group, bots := newScoringGroup(t)

	group.Finish(monopoly.WIN, 0, finishState(1, 2, 3))
	group.Finish(monopoly.WIN, 0, finishState(1, 2, 3))

	assert.Equal(t, cfg.FIRST_PLACE_SCORE, bots[0].GetScore())
	assert.Equal(t, 1, bots[0].GetWins())
}

// TestFinish_ScoresAccumulateAcrossGames — fitness is the average over games, so scores of the same
// organism from many games must add up (instances from ForGame share the stats).
func TestFinish_ScoresAccumulateAcrossGames(t *testing.T) {
	bot := &NEATMonopolyPlayer{stats: &playerStats{}}
	others := []MonopolyPlayer{&NEATMonopolyPlayer{stats: &playerStats{}}, &NEATMonopolyPlayer{stats: &playerStats{}}, &NEATMonopolyPlayer{stats: &playerStats{}}}

	for game, winner := range []int{0, 1, 0} {
		group, err := NewNEATPlayerGroup(game, append([]MonopolyPlayer{bot}, others...))
		require.NoError(t, err)
		group.Finish(monopoly.WIN, winner, finishState())
	}

	assert.Equal(t, 2*cfg.FIRST_PLACE_SCORE, bot.GetScore())
	assert.Equal(t, 2, bot.GetWins())
}
