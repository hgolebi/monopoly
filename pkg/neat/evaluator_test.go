package neatnetwork

import (
	"context"
	"fmt"
	"math/rand"
	"monopoly/pkg/monopoly"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaricom/goNEAT/v4/neat"
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

// TestNextGameSeed — game seeds come from the evaluator's rng: the same evaluator seed gives the same
// sequence, and no seed is 0 (NewGame would replace it with the clock).
func TestNextGameSeed(t *testing.T) {
	first := NewMonopolyEvaluator(t.TempDir(), 4, rand.New(rand.NewSource(7)))
	second := NewMonopolyEvaluator(t.TempDir(), 4, rand.New(rand.NewSource(7)))
	seen := make(map[int64]bool)
	for range 1000 {
		seed := first.nextGameSeed()
		assert.Equal(t, seed, second.nextGameSeed())
		assert.Positive(t, seed)
		seen[seed] = true
	}
	assert.Len(t, seen, 1000)
}

// TestPlayRound_Reproducible — two evaluations with the same evaluator seed give the same scores,
// although the games run concurrently.
func TestPlayRound_Reproducible(t *testing.T) {
	genomes := []string{
		"../../genomes/bestForNow",
		"../../genomes/bestForNow copy",
		"../../genomes/bestForNow copy 2",
		"../../genomes/bestForNow copy 3",
	}
	ctx := neat.NewContext(context.Background(), &neat.Options{NumGenerations: 1000})
	evaluate := func() []int {
		players := make([]MonopolyPlayer, len(genomes))
		for i, genome := range genomes {
			players[i] = loadTestPlayer(t, genome)
		}
		scored := slices.Clone(players)
		e := NewMonopolyEvaluator(t.TempDir(), 4, rand.New(rand.NewSource(42)))
		require.NoError(t, e.PlayRound(ctx, players, 0))
		scores := make([]int, len(scored))
		for i, player := range scored {
			scores[i] = player.GetScore()
		}
		return scores
	}

	assert.Equal(t, evaluate(), evaluate())
}

// TestTrainNetworkSeed_GlobalRand — goNEAT uses the global math/rand source, so rand.Seed must still
// take effect (it becomes a no-op when go.mod targets go 1.24+ without GODEBUG=randseednop=0).
func TestTrainNetworkSeed_GlobalRand(t *testing.T) {
	rand.Seed(5) //nolint:staticcheck
	first := rand.Int63()
	rand.Seed(5) //nolint:staticcheck
	assert.Equal(t, first, rand.Int63())
}
