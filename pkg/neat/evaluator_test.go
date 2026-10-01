package neatnetwork

import (
	"context"
	"fmt"
	"monopoly/pkg/monopoly"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

type nopLogger struct{}

func (nopLogger) Log(string)                              {}
func (nopLogger) LogWithState(string, monopoly.GameState) {}
func (nopLogger) LogState(monopoly.GameState)             {}
func (nopLogger) Error(string, monopoly.GameState)        {}

// resultRecorder records the final state of a game as a comparable string.
type resultRecorder struct {
	*NEATPlayerGroup
	result string
}

func (r *resultRecorder) Finish(f monopoly.FinishOption, winner int, state monopoly.GameState) {
	r.result = fmt.Sprintf("finish=%v winner=%d round=%d", f, winner, state.Round)
	for _, p := range state.Players {
		r.result += fmt.Sprintf(" [money=%d properties=%v]", p.Money, p.Properties)
	}
	r.NEATPlayerGroup.Finish(f, winner, state)
}

func loadTestPlayer(t *testing.T, genomeFile string) *NEATMonopolyPlayer {
	t.Helper()
	reader, err := genetics.NewGenomeReaderFromFile(genomeFile)
	if err != nil {
		t.Skipf("genome %s not available: %v", genomeFile, err)
	}
	genome, err := reader.Read()
	require.NoError(t, err)
	organism, err := genetics.NewOrganism(0.0, genome, 0)
	require.NoError(t, err)
	player, err := NewNEATMonopolyPlayer(organism)
	require.NoError(t, err)
	return player
}

func playTestGame(t *testing.T, players []MonopolyPlayer, seed int64) string {
	t.Helper()
	group, err := NewNEATPlayerGroup(0, players)
	require.NoError(t, err)
	recorder := &resultRecorder{NEATPlayerGroup: group}
	monopoly.NewGame(context.Background(), recorder, nopLogger{}, seed).Start()
	return recorder.result
}

// TestForGame_ConcurrentGamesMatchSequential — games played concurrently by instances of the same organisms
// must end exactly like the same games played one after another. A network shared between concurrent games
// mixes activations of different games and changes decisions.
func TestForGame_ConcurrentGamesMatchSequential(t *testing.T) {
	const games = 40
	players := []MonopolyPlayer{
		loadTestPlayer(t, "../../genomes/bestForNow"),
		loadTestPlayer(t, "../../genomes/bestForNow copy"),
		loadTestPlayer(t, "../../genomes/bestForNow copy 2"),
		loadTestPlayer(t, "../../genomes/bestForNow copy 3"),
	}

	sequential := make([]string, games)
	for i := range games {
		gamePlayers, err := playersForGame(players)
		require.NoError(t, err)
		sequential[i] = playTestGame(t, gamePlayers, int64(i+1))
	}

	instances := make([][]MonopolyPlayer, games)
	for i := range games {
		gamePlayers, err := playersForGame(players)
		require.NoError(t, err)
		instances[i] = gamePlayers
	}
	concurrent := make([]string, games)
	var wg sync.WaitGroup
	for i := range games {
		wg.Add(1)
		go func() {
			defer wg.Done()
			concurrent[i] = playTestGame(t, instances[i], int64(i+1))
		}()
	}
	wg.Wait()

	for i := range games {
		assert.Equal(t, sequential[i], concurrent[i], "game with seed %d", i+1)
	}
}

// TestForGame_SharesStats — scores of all game instances are added to the same organism.
func TestForGame_SharesStats(t *testing.T) {
	player := loadTestPlayer(t, "../../genomes/bestForNow")
	first, err := player.ForGame()
	require.NoError(t, err)
	second, err := player.ForGame()
	require.NoError(t, err)

	first.AddScore(10)
	second.AddScore(5)
	first.AddWin()
	second.AddDraw()

	assert.Equal(t, 15, player.GetScore())
	assert.Equal(t, 1, player.GetWins())
	assert.Equal(t, 1, player.GetDraws())
	assert.NotSame(t, player.network, first.(*NEATMonopolyPlayer).network)
	assert.NotSame(t, first.(*NEATMonopolyPlayer).network, second.(*NEATMonopolyPlayer).network)
}
