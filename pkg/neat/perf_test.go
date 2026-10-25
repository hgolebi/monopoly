package neatnetwork

// Deterministic performance workload, reused to measure every fix from .ai_reports/raport_wydajnosc.md.
//
// The workload is a fixed list of games: fixed networks (synthetic genomes with seeded weights and a frozen
// trained genome from testdata), fixed group assignments and fixed game seeds. It does not depend on the
// clock, the global RNG or the scheduling of goroutines, so every run plays exactly the same games.
//
//   go test ./pkg/neat -run TestPerfWorkload -v                    # activation counts per decision path + fingerprint
//   go test ./pkg/neat -run '^$' -bench PerfWorkload -benchmem -count 10   # time and allocations
//
// Activation counts and the fingerprint are exact, so they can be compared directly. Time is noisy: compare
// several runs with benchstat (scripts/perf.ps1 does both).
// A change that only speeds the code up must keep the fingerprint. A change of the bot's decisions
// (e.g. binary price search) changes it, and the activation counts show by how much the work shrank.

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
	"github.com/yaricom/goNEAT/v4/neat/network"
)

const (
	perfSeed          = 20261025
	perfRounds        = 25 // each round plays one game per group of 4, like one round of an epoch
	perfFlatPlayers   = 6  // all-to-all genome (inputs → outputs), depth 1
	perfHiddenPlayers = 6  // one hidden layer of perfHiddenNodes plus direct connections, depth 2
	perfHiddenNodes   = 8
	perfTrainedCopies = 4 // copies of testdata/perf_trained_genome (63 hidden nodes)
	perfWeightRange   = 3.0
	perfTrainedGenome = "testdata/perf_trained_genome"
)

type perfGame struct {
	players []int // indices into the player list
	seed    int64
}

type perfWorkload struct {
	players []MonopolyPlayer
	games   []perfGame
}

// newPerfWorkload builds the players and the list of games. Everything is derived from perfSeed.
func newPerfWorkload(tb testing.TB) *perfWorkload {
	tb.Helper()
	rng := rand.New(rand.NewSource(perfSeed))

	var genomes []*genetics.Genome
	for range perfFlatPlayers {
		genomes = append(genomes, newPerfGenome(len(genomes)+1, 0, rng))
	}
	for range perfHiddenPlayers {
		genomes = append(genomes, newPerfGenome(len(genomes)+1, perfHiddenNodes, rng))
	}
	for range perfTrainedCopies {
		genome := loadOrBuildGenome(perfTrainedGenome, false)
		genome.Id = len(genomes) + 1
		genomes = append(genomes, genome)
	}

	w := &perfWorkload{}
	for _, genome := range genomes {
		organism, err := genetics.NewOrganism(0, genome, 0)
		if err != nil {
			tb.Fatalf("organism %d: %v", genome.Id, err)
		}
		player, err := NewNEATMonopolyPlayer(organism)
		if err != nil {
			tb.Fatalf("player %d: %v", genome.Id, err)
		}
		w.players = append(w.players, player)
	}

	order := make([]int, len(w.players))
	for i := range order {
		order[i] = i
	}
	for range perfRounds {
		rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		for i := 0; i+4 <= len(order); i += 4 {
			w.games = append(w.games, perfGame{
				players: slices.Clone(order[i : i+4]),
				seed:    rng.Int63n(1<<62) + 1, // 0 means "use the clock" in NewGame
			})
		}
	}
	return w
}


// newPerfGenome builds a genome with weights drawn from rng: inputs and bias connected to every output and,
// if hidden > 0, to every hidden node, which is connected to every output.
func newPerfGenome(id int, hidden int, rng *rand.Rand) *genetics.Genome {
	trait := neat.NewTrait()
	trait.Id = 1
	nodes := buildNodes(trait)
	sources := nodes[:int(INPUT_COUNT)+1] // inputs + bias
	outputs := nodes[int(INPUT_COUNT)+1:]

	var hiddenNodes []*network.NNode
	nextId := outputs[len(outputs)-1].Id + 1
	for i := range hidden {
		n := network.NewNNode(nextId+i, network.HiddenNeuron)
		n.ActivationType = outputs[0].ActivationType
		n.Trait = trait
		hiddenNodes = append(hiddenNodes, n)
	}
	nodes = append(nodes, hiddenNodes...)

	var genes []*genetics.Gene
	innovation := int64(1)
	connect := func(in, out *network.NNode) {
		weight := (rng.Float64()*2 - 1) * perfWeightRange
		genes = append(genes, genetics.NewGeneWithTrait(trait, weight, in, out, false, innovation, 0))
		innovation++
	}
	for _, in := range sources {
		for _, out := range outputs {
			connect(in, out)
		}
		for _, out := range hiddenNodes {
			connect(in, out)
		}
	}
	for _, in := range hiddenNodes {
		for _, out := range outputs {
			connect(in, out)
		}
	}
	return genetics.NewGenome(id, []*neat.Trait{trait}, nodes, genes)
}

var perfLoggerOnce sync.Once

func perfContext() context.Context {
	perfLoggerOnce.Do(func() { _ = neat.InitLogger("error") })
	// NumGenerations far from epoch 0 keeps the game logs disabled, as in almost every training epoch
	return neat.NewContext(context.Background(), &neat.Options{NumGenerations: 1000})
}

// gameDetails builds fresh network instances for one game. Like PlayRound, it must run on a single goroutine,
// because ForGame (Genesis) modifies the organism's genome.
func (w *perfWorkload) gameDetails(tb testing.TB, id int) GroupDetails {
	game := w.games[id]
	group := make([]MonopolyPlayer, len(game.players))
	for i, idx := range game.players {
		group[i] = w.players[idx]
	}
	gamePlayers, err := playersForGame(group)
	if err != nil {
		tb.Fatal(err)
	}
	return GroupDetails{Epoch: 0, Round: id, GroupID: 0, Players: gamePlayers, Seed: game.seed}
}

func playPerfGame(ctx context.Context, tb testing.TB, gd GroupDetails, outputDir string) {
	if err := startGroup(ctx, gd, outputDir); err != nil {
		tb.Error(err)
	}
}

func (w *perfWorkload) runSequential(ctx context.Context, tb testing.TB, outputDir string) {
	for id := range w.games {
		playPerfGame(ctx, tb, w.gameDetails(tb, id), outputDir)
	}
}

// runParallel plays the games on runtime.NumCPU workers, the way PlayRound does. Each game has its own seed and
// network instances and the statistics are sums, so the result does not depend on the order the games finish in.
func (w *perfWorkload) runParallel(ctx context.Context, tb testing.TB, outputDir string) {
	jobs := make(chan GroupDetails, runtime.NumCPU())
	var wg sync.WaitGroup
	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for gd := range jobs {
				playPerfGame(ctx, tb, gd, outputDir)
			}
		}()
	}
	for id := range w.games {
		jobs <- w.gameDetails(tb, id)
	}
	close(jobs)
	wg.Wait()
}

// fingerprint summarizes the results of all games. It changes only when some decision or game changes.
func (w *perfWorkload) fingerprint() string {
	var sb strings.Builder
	for _, p := range w.players {
		fmt.Fprintf(&sb, "%d:%d/%d/%d;", p.GetId(), p.GetScore(), p.GetWins(), p.GetDraws())
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(sb.String())))[:16]
}

type activationCounts struct {
	total   int64
	byPath  map[string]int64 // "entry ← direct caller of GetDecision"
	byEntry map[string]int64 // decision requested by the game (interface method)
}

// countActivations plays the workload sequentially with decisionHook set and attributes every activation
// to the decision path it came from.
func countActivations(tb testing.TB, outputDir string) (activationCounts, string) {
	w := newPerfWorkload(tb)
	counts := activationCounts{byPath: map[string]int64{}, byEntry: map[string]int64{}}
	pcs := make([]uintptr, 64)
	decisionHook = func() {
		entry, direct := decisionPath(pcs)
		counts.total++
		counts.byEntry[entry]++
		counts.byPath[entry+" ← "+direct]++
	}
	defer func() { decisionHook = nil }()
	w.runSequential(perfContext(), tb, outputDir)
	return counts, w.fingerprint()
}

// decisionPath returns the outermost NEATMonopolyPlayer method on the stack and the caller of GetDecision.
func decisionPath(pcs []uintptr) (entry string, direct string) {
	const playerPrefix = ".(*NEATMonopolyPlayer)."
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	afterGetDecision := false
	for {
		frame, more := frames.Next()
		name := frame.Function[strings.LastIndex(frame.Function, "/")+1:]
		if _, method, found := strings.Cut(name, playerPrefix); found {
			if afterGetDecision && direct == "" {
				direct = method
			}
			if method == "GetDecision" {
				afterGetDecision = true
			}
			entry = method
		}
		if !more {
			break
		}
	}
	return entry, direct
}

func TestPerfWorkloadDeterministic(t *testing.T) {
	if testing.Short() {
		t.Skip("plays the whole performance workload")
	}
	runs := []struct {
		name string
		run  func(w *perfWorkload, ctx context.Context, tb testing.TB, outputDir string)
	}{
		{"sequential", (*perfWorkload).runSequential},
		{"sequential again", (*perfWorkload).runSequential},
		{"parallel", (*perfWorkload).runParallel},
	}
	var want string
	for _, r := range runs {
		w := newPerfWorkload(t)
		r.run(w, perfContext(), t, t.TempDir())
		got := w.fingerprint()
		if want == "" {
			want = got
		} else if got != want {
			t.Fatalf("workload is not deterministic: %s run gave %s, first run %s", r.name, got, want)
		}
	}
}

// TestPerfWorkload prints the exact amount of work done by the workload. Compare its output before and after a fix.
func TestPerfWorkload(t *testing.T) {
	if testing.Short() {
		t.Skip("plays the whole performance workload")
	}
	counts, fingerprint := countActivations(t, t.TempDir())
	w := newPerfWorkload(t)
	games := int64(len(w.games))

	var sb strings.Builder
	fmt.Fprintf(&sb, "games:          %d\n", games)
	fmt.Fprintf(&sb, "fingerprint:    %s\n", fingerprint)
	fmt.Fprintf(&sb, "activations:    %d (%.1f per game)\n", counts.total, float64(counts.total)/float64(games))
	sb.WriteString("\nactivations by decision requested by the game:\n")
	writeCounts(&sb, counts.byEntry, counts.total)
	sb.WriteString("\nactivations by path (entry ← caller of GetDecision):\n")
	writeCounts(&sb, counts.byPath, counts.total)
	t.Log("\n" + sb.String())
}

func writeCounts(sb *strings.Builder, counts map[string]int64, total int64) {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if counts[keys[i]] != counts[keys[j]] {
			return counts[keys[i]] > counts[keys[j]]
		}
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		fmt.Fprintf(sb, "  %-62s %10d  %5.1f%%\n", k, counts[k], 100*float64(counts[k])/float64(total))
	}
}

// counted once per process: the counts are exact, so repeating them for every -count would only waste time
var (
	perfCountsOnce sync.Once
	perfCounts     activationCounts
)

// BenchmarkPerfWorkload measures time and allocations of the whole workload (one op = all games).
// "sequential" shows the cost of a game on one core, "parallel" the throughput of a training-like worker pool.
func BenchmarkPerfWorkload(b *testing.B) {
	perfCountsOnce.Do(func() { perfCounts, _ = countActivations(b, b.TempDir()) })
	counts := perfCounts
	for _, mode := range []string{"sequential", "parallel"} {
		b.Run(mode, func(b *testing.B) {
			ctx := perfContext()
			dir := b.TempDir()
			var games int
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				b.StopTimer()
				w := newPerfWorkload(b) // fresh statistics, network building is not measured
				games = len(w.games)
				b.StartTimer()
				if mode == "sequential" {
					w.runSequential(ctx, b, dir)
				} else {
					w.runParallel(ctx, b, dir)
				}
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*games), "ns/game")
			b.ReportMetric(float64(counts.total), "activations/op")
		})
	}
}
