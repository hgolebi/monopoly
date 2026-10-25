package neatnetwork

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
	neatmath "github.com/yaricom/goNEAT/v4/neat/math"
	"github.com/yaricom/goNEAT/v4/neat/network"
)

// randomDAGGenome builds a feed-forward genome with the structures evolution produces: hidden neurons in an order
// different from the evaluation order, disabled genes, neurons without inputs, neurons that lead nowhere,
// possibly outputs without inputs, and several activation functions.
func randomDAGGenome(id int, rng *rand.Rand) *genetics.Genome {
	trait := neat.NewTrait()
	trait.Id = 1
	nodes := buildNodes(trait)
	sources := nodes[:int(INPUT_COUNT)+1]
	outputs := nodes[int(INPUT_COUNT)+1:]
	activations := []neatmath.NodeActivationType{
		neatmath.SigmoidSteepenedActivation, neatmath.SigmoidSteepenedActivation,
		neatmath.TanhActivation, neatmath.LinearActivation, neatmath.SigmoidPlainActivation,
	}

	hiddenCount := rng.Intn(25)
	hidden := make([]*network.NNode, hiddenCount)
	nextId := outputs[len(outputs)-1].Id + 1
	for i := range hidden {
		n := network.NewNNode(nextId+i, network.HiddenNeuron)
		n.ActivationType = activations[rng.Intn(len(activations))]
		n.Trait = trait
		hidden[i] = n
	}
	nodes = append(nodes, hidden...)
	rank := rng.Perm(hiddenCount) // links between hidden neurons go from lower to higher rank

	var genes []*genetics.Gene
	innovation := int64(1)
	connect := func(in, out *network.NNode) {
		gene := genetics.NewGeneWithTrait(trait, rng.Float64()*6-3, in, out, false, innovation, 0)
		gene.IsEnabled = rng.Float64() > 0.15
		genes = append(genes, gene)
		innovation++
	}
	for _, out := range append(append([]*network.NNode{}, hidden...), outputs...) {
		for range rng.Intn(4) {
			connect(sources[rng.Intn(len(sources))], out)
		}
	}
	for i := range hidden {
		for j := range hidden {
			if rank[i] < rank[j] && rng.Float64() < 0.2 {
				connect(hidden[i], hidden[j])
			}
		}
		for _, out := range outputs {
			if rng.Float64() < 0.3 {
				connect(hidden[i], out)
			}
		}
	}
	rng.Shuffle(len(genes), func(i, j int) { genes[i], genes[j] = genes[j], genes[i] })
	if len(genes) == 0 {
		connect(sources[0], outputs[0])
	}
	return genetics.NewGenome(id, []*neat.Trait{trait}, nodes, genes)
}

func randomInputs(rng *rand.Rand) []float64 {
	in := make([]float64, int(INPUT_COUNT))
	for i := range in {
		in[i] = rng.Float64()
	}
	return in
}

// assertSameAsGoNEAT activates the compiled network and goNEAT's Network on the same inputs, one after another
// (the state carries over between decisions in both), and requires bit-for-bit equal outputs.
func assertSameAsGoNEAT(t *testing.T, genome *genetics.Genome, rng *rand.Rand, decisions int) bool {
	t.Helper()
	net, err := genome.Genesis(genome.Id)
	require.NoError(t, err)
	depth, err := net.MaxActivationDepthWithCap(0)
	if err != nil || depth <= 0 {
		return false // such networks are rejected by NewNEATMonopolyPlayer
	}
	fast, err := compileFastNetwork(genome)
	require.NoError(t, err)

	for range decisions {
		in := randomInputs(rng)
		require.NoError(t, net.LoadSensors(in))
		_, _ = net.ForwardSteps(depth) // errors for networks with an inactive output, like in GetDecision
		got, err := fast.activate(in)
		require.NoError(t, err)
		for i, out := range net.Outputs {
			if math.Float64bits(out.Activation) != math.Float64bits(got[i]) {
				t.Fatalf("genome %d, output %d: goNEAT %v, compiled %v", genome.Id, i, out.Activation, got[i])
			}
		}
	}
	return true
}

func TestFastNetworkMatchesGoNEAT(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	t.Run("workload networks", func(t *testing.T) {
		genomes := []*genetics.Genome{
			newPerfGenome(1, 0, rng),
			newPerfGenome(2, perfHiddenNodes, rng),
			loadOrBuildGenome(perfTrainedGenome, false),
		}
		for _, genome := range genomes {
			assert.True(t, assertSameAsGoNEAT(t, genome, rng, 2000))
		}
	})
	t.Run("random feed-forward networks", func(t *testing.T) {
		compared := 0
		for id := range 500 {
			if assertSameAsGoNEAT(t, randomDAGGenome(id+1, rng), rng, 30) {
				compared++
			}
		}
		assert.Greater(t, compared, 400)
	})
}

func TestFastNetworkSteepenedSigmoidMatchesGoNEAT(t *testing.T) {
	for x := -10.0; x <= 10.0; x += 0.001 {
		want, err := neatmath.NodeActivators.ActivateByType(x, nil, neatmath.SigmoidSteepenedActivation)
		require.NoError(t, err)
		assert.Equal(t, math.Float64bits(want), math.Float64bits(1.0/(1.0+math.Exp(-4.924273*x))), "x=%v", x)
	}
}

func TestFastNetworkInstancesHaveOwnValues(t *testing.T) {
	genome := loadOrBuildGenome(perfTrainedGenome, false)
	first, err := compileFastNetwork(genome)
	require.NoError(t, err)
	second := first.newInstance()
	reference, err := compileFastNetwork(genome)
	require.NoError(t, err)

	rng := rand.New(rand.NewSource(3))
	for range 100 {
		a, b := randomInputs(rng), randomInputs(rng)
		gotA, _ := first.activate(a)
		gotB, _ := second.activate(b)
		wantA, _ := reference.activate(a)
		wantB, _ := reference.activate(b)
		assert.Equal(t, wantA, gotA)
		assert.Equal(t, wantB, gotB)
	}
}

func TestFastNetworkRejectsRecurrentNetworks(t *testing.T) {
	rng := rand.New(rand.NewSource(4))

	cycle := newPerfGenome(1, 2, rng)
	hidden := cycle.Nodes[len(cycle.Nodes)-2:]
	cycle.Genes = append(cycle.Genes,
		genetics.NewGeneWithTrait(cycle.Traits[0], 1, hidden[0], hidden[1], false, 1000, 0),
		genetics.NewGeneWithTrait(cycle.Traits[0], 1, hidden[1], hidden[0], false, 1001, 0))
	_, err := compileFastNetwork(cycle)
	assert.ErrorIs(t, err, errRecurrentNetwork, "cycle between hidden neurons")

	flagged := newPerfGenome(2, 0, rng)
	flagged.Genes[0].Link.IsRecurrent = true
	_, err = compileFastNetwork(flagged)
	assert.ErrorIs(t, err, errRecurrentNetwork, "link marked as recurrent")
}

func TestPlayerFallsBackToGoNEATForRecurrentNetworks(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	genome := newPerfGenome(1, 0, rng)
	genome.Genes[0].Link.IsRecurrent = true
	organism, err := genetics.NewOrganism(0, genome, 0)
	require.NoError(t, err)
	player, err := NewNEATMonopolyPlayer(organism)
	require.NoError(t, err)
	assert.Nil(t, player.fast)
	assert.Len(t, player.GetDecision(randomInputs(rng)), int(OUTPUT_COUNT))

	instance, err := player.ForGame()
	require.NoError(t, err)
	assert.Nil(t, instance.(*NEATMonopolyPlayer).fast)
}

func TestPlayerUsesCompiledNetwork(t *testing.T) {
	player := newTestNetworkPlayer(t)
	require.NotNil(t, player.fast)
	instance, err := player.ForGame()
	require.NoError(t, err)
	gamePlayer := instance.(*NEATMonopolyPlayer)
	assert.Same(t, player.fast.fastStructure, gamePlayer.fast.fastStructure, "the structure is shared")
	assert.NotSame(t, &player.fast.values[0], &gamePlayer.fast.values[0], "the values are not")
}

// TestPlayerWithDisconnectedOutputs — evolution can disable every link to the outputs. goNEAT reports depth 0 for
// such a network, which used to make NewNEATMonopolyPlayer fail and abort the training; now the player is created,
// always answers "no" and plays a whole game.
func TestPlayerWithDisconnectedOutputs(t *testing.T) {
	rng := rand.New(rand.NewSource(6))
	disconnected := newPerfGenome(1, perfHiddenNodes, rng)
	for _, gene := range disconnected.Genes {
		if gene.Link.OutNode.NeuronType == network.OutputNeuron {
			gene.IsEnabled = false
		}
	}
	net, err := disconnected.Genesis(1)
	require.NoError(t, err)
	depth, err := net.MaxActivationDepthWithCap(0)
	require.NoError(t, err)
	require.Zero(t, depth, "goNEAT reports depth 0 for this network")

	noGenes := newPerfGenome(2, 0, rng)
	noGenes.Genes = nil

	for _, genome := range []*genetics.Genome{disconnected, noGenes} {
		organism, err := genetics.NewOrganism(0, genome, 0)
		require.NoError(t, err)
		player, err := NewNEATMonopolyPlayer(organism)
		require.NoError(t, err, "genome %d", genome.Id)
		require.NotNil(t, player.fast)
		assert.Equal(t, make([]float64, int(OUTPUT_COUNT)), player.GetDecision(randomInputs(rng)))

		group := []MonopolyPlayer{player}
		for id := range 3 {
			other, err := genetics.NewOrganism(0, newPerfGenome(10+id, 0, rng), 0)
			require.NoError(t, err)
			otherPlayer, err := NewNEATMonopolyPlayer(other)
			require.NoError(t, err)
			group = append(group, otherPlayer)
		}
		gamePlayers, err := playersForGame(group)
		require.NoError(t, err)
		gd := GroupDetails{Players: gamePlayers, Seed: 1}
		require.NoError(t, startGroup(perfContext(), gd, t.TempDir()))
	}
}
