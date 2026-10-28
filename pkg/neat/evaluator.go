package neatnetwork

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	cfg "monopoly/pkg/config"
	"monopoly/pkg/monopoly"

	"github.com/yaricom/goNEAT/v4/experiment"
	"github.com/yaricom/goNEAT/v4/experiment/utils"
	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

type MonopolyEvaluator struct {
	outputDir string
	groupSize int
	rng       *rand.Rand
}

func NewMonopolyEvaluator(outputDir string, groupSize int, rng *rand.Rand) *MonopolyEvaluator {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return &MonopolyEvaluator{
		outputDir: outputDir,
		groupSize: groupSize,
		rng:       rng,
	}
}

type GroupDetails struct {
	Epoch   int
	Round   int
	GroupID int
	Players []MonopolyPlayer
	Seed    int64 // seed of the game's random source, drawn from the evaluator's rng
}

func (e *MonopolyEvaluator) GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {

	options, ok := neat.FromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to get options from context")
	}

	// create players from population
	players, err := e.createPlayersFromPopulation(pop)
	if err != nil {
		return fmt.Errorf("failed to create players from population: %v", err)
	}

	if err := e.PlayRound(ctx, players, epoch.Id); err != nil {
		return fmt.Errorf("failed to play round: %v", err)
	}
	for _, player := range players {
		player.GetOrganism().Fitness = float64(player.GetScore()) / cfg.GAMES_PER_EPOCH
	}
	epoch.FillPopulationStatistics(pop)
	numberOfSpecies := len(pop.Species)
	bestPlayer := findBestPlayer(players)
	bestOrg := bestPlayer.GetOrganism()

	// Disable goNEAT's delta coding: when the population's fitness record has not been beaten for dropoff_age+5 epochs,
	// it gives the whole next generation to the two best species. Fitness is measured against the evolving population,
	// so an early record (e.g. a random network among random ones in epoch 0) is never beaten and delta coding would
	// wipe out the diversity every dropoff_age+5 epochs. dropoff_age still removes stagnating species.
	pop.EpochsHighestLastChanged = 0

	// log info
	compatThreshold := options.CompatThreshold
	adjustCompatThreshold(options, numberOfSpecies)
	neat.InfoLog(fmt.Sprintf("Species count: %d, compat threshold: %.2f -> %.2f\n", numberOfSpecies, compatThreshold, options.CompatThreshold))
	neat.InfoLog(fmt.Sprintf("Champion of epoch %d is organism %d\n with fitness: %f (wins: %d, draws: %d)", epoch.Id, bestOrg.Genotype.Id, bestOrg.Fitness, bestPlayer.GetWins(), bestPlayer.GetDraws()))
	neat.InfoLog(fmt.Sprintf("Number of nodes: %d, number of connections: %d\n", len(bestOrg.Genotype.Nodes), len(bestOrg.Genotype.Genes)))

	// dump population
	if (epoch.Id+1)%cfg.PRINT_EVERY == 0 || epoch.Id == options.NumGenerations-1 {
		if _, err := utils.WritePopulationPlain(e.outputDir, pop, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump population, reason: %s\n", err))
			return err
		}
	} else {
		// dump only champion
		genomeFile := fmt.Sprintf("gen_%d_champion", epoch.Id)
		if _, err := utils.WriteGenomePlain(genomeFile, e.outputDir, bestOrg, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump champion organism, reason: %s\n", err))
			return err
		}
	}

	// add line to champions.txt with champion info
	if err := appendChampionInfo(e.outputDir, bestPlayer, epoch.Id, numberOfSpecies, compatThreshold); err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to append champion info, reason: %s\n", err))
		return err
	}
	return nil
}

func (e *MonopolyEvaluator) PlayRound(ctx context.Context, players []MonopolyPlayer, epoch int) error {
	options, ok := neat.FromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to get options from context")
	}
	if err := checkGroupSize(len(players), e.groupSize); err != nil {
		return err
	}
	// start workers
	jobsCh := make(chan GroupDetails, 100)
	var wg sync.WaitGroup
	for i := range cfg.MAX_THREADS {
		wg.Add(1)
		go startWorker(ctx, i, jobsCh, &wg, e.outputDir)
	}

	// starting rounds
	var err error
rounds:
	for roundID := range cfg.GAMES_PER_EPOCH {
		// prepare groups
		groups := e.prepareGroups(players)

		if roundID == 0 && (epoch == options.NumGenerations-1 || (epoch+1)%cfg.PRINT_EVERY == 0) {
			dumpGroupAssignments(e.outputDir, epoch, roundID, groups)
		}

		// create job for every group
		for groupID, group := range groups {
			gamePlayers, gameErr := playersForGame(group)
			if gameErr != nil {
				err = fmt.Errorf("group %d (round %d): %w", groupID, roundID, gameErr)
				break rounds
			}
			gd := GroupDetails{
				Epoch:   epoch,
				Round:   roundID,
				GroupID: groupID,
				Players: gamePlayers,
				Seed:    e.nextGameSeed(),
			}
			jobsCh <- gd
		}
	}
	close(jobsCh)
	wg.Wait()
	return err
}

// checkGroupSize requires the players to split into full groups. Otherwise the last group would play with fewer
// players (an unfair fitness) or alone, which the game does not allow (panic in a worker, ending the training).
func checkGroupSize(players int, groupSize int) error {
	if players == 0 || players%groupSize != 0 {
		return fmt.Errorf("the number of players (%d) must be a positive multiple of the group size (%d); "+
			"set pop_size in the NEAT options accordingly", players, groupSize)
	}
	return nil
}

// nextGameSeed draws a game seed from the evaluator's rng, so every game gets its own seed and the whole
// evaluation is reproducible from the evaluator's seed. Zero is skipped, because NewGame treats it as "use the clock".
func (e *MonopolyEvaluator) nextGameSeed() int64 {
	return e.rng.Int63n(math.MaxInt64) + 1
}

func playersForGame(group []MonopolyPlayer) ([]MonopolyPlayer, error) {
	gamePlayers := make([]MonopolyPlayer, len(group))
	for i, player := range group {
		gamePlayer, err := player.ForGame()
		if err != nil {
			return nil, err
		}
		gamePlayers[i] = gamePlayer
	}
	return gamePlayers, nil
}

func startWorker(ctx context.Context, id int, jobsCh <-chan GroupDetails, wg *sync.WaitGroup, outputDir string) {
	defer wg.Done()
	// neat.InfoLog(fmt.Sprintf("Worker %d started\n", id))
	for gd := range jobsCh {
		// neat.InfoLog(fmt.Sprintf("Worker %d processing group %d (round %d)\n", id, gd.GroupID, gd.Round))
		if err := startGroup(ctx, gd, outputDir); err != nil {
			neat.ErrorLog(err.Error())
			continue
		}
	}
}

func (e *MonopolyEvaluator) prepareGroups(players []MonopolyPlayer) [][]MonopolyPlayer {
	e.rng.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
	var groups [][]MonopolyPlayer
	for i := 0; i < len(players); i += (e.groupSize) {
		end := min(i+e.groupSize, len(players))
		group := make([]MonopolyPlayer, 0, e.groupSize)
		group = append(group, players[i:end]...)
		groups = append(groups, group)
	}
	return groups
}

func startGroup(ctx context.Context, gd GroupDetails, outputDir string) error {
	neat.DebugLog(fmt.Sprintf("Starting group %d (round %d)\n", gd.GroupID, gd.Round))
	options, ok := neat.FromContext(ctx)
	if !ok {
		return fmt.Errorf("Error in group %d (round %d): %s", gd.GroupID, gd.Round, "failed to get options from context")
	}
	playerGroup, err := NewNEATPlayerGroup(gd.GroupID, gd.Players)
	if err != nil {
		return fmt.Errorf("Error in group %d (round %d): %v", gd.GroupID, gd.Round, err)
	}
	enable_log := false
	if gd.Epoch == options.NumGenerations-1 {
		enable_log = gd.Round < 10
	} else if (gd.Epoch+1)%cfg.PRINT_EVERY == 0 {
		enable_log = gd.Round == 0
	}
	logger, err := NewTrainerLogger(fmt.Sprintf("%s/games/epoch%d/round%d/group%d",
		outputDir, gd.Epoch, gd.Round, gd.GroupID), !enable_log)
	if err != nil {
		return fmt.Errorf("Error in group %d (round %d): %v", gd.GroupID, gd.Round, err)
	}

	game := monopoly.NewGame(ctx, playerGroup, logger, gd.Seed)
	game.Start()
	return nil
}

func dumpGroupAssignments(outputDir string, epoch int, round int, groups [][]MonopolyPlayer) {
	filePath := fmt.Sprintf("%s/games/epoch%d/round%d/group_assignments.txt", outputDir, epoch, round)
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to create directory for group assignments, reason: %s\n", err))
		return
	}
	file, err := os.Create(filePath)
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to create group assignments file, reason: %s\n", err))
		return
	}
	defer file.Close()
	players := map[int]int{}
	for i, group := range groups {
		fmt.Fprintf(file, "Group %d: ", i)
		for _, player := range group {
			fmt.Fprintf(file, "\tPlayer %d", player.GetId())
			players[player.GetId()] = i
		}
		fmt.Fprintln(file)
	}
	for i := 0; i < len(players); i++ {
		fmt.Fprintf(file, "Player %d: Group %d\n", i, players[i])
	}

}

func (e *MonopolyEvaluator) createPlayersFromPopulation(pop *genetics.Population) ([]MonopolyPlayer, error) {
	var players []MonopolyPlayer

	for i, org := range pop.Organisms {
		org.Fitness = 0
		org, err := NewNEATMonopolyPlayer(org)
		if err != nil {
			return nil, fmt.Errorf("error creating NEATMonopolyPlayer for organism %d: %v", i, err)
		}
		players = append(players, org)
	}
	return players, nil
}

// adjustCompatThreshold moves the compatibility threshold towards TARGET_SPECIES species for the next speciation:
// up when there are too many species (a new species then needs a bigger difference), down when there are too few.
// The options are shared through the context, so goNEAT's speciation of the next generation uses the new value.
func adjustCompatThreshold(options *neat.Options, species int) {
	if species > cfg.TARGET_SPECIES {
		options.CompatThreshold = min(cfg.COMPAT_THRESHOLD_MAX, options.CompatThreshold+cfg.COMPAT_THRESHOLD_STEP_UP)
	} else if species < cfg.TARGET_SPECIES {
		options.CompatThreshold = max(cfg.COMPAT_THRESHOLD_MIN, options.CompatThreshold-cfg.COMPAT_THRESHOLD_STEP_DOWN)
	}
}

func appendChampionInfo(outputDir string, champion MonopolyPlayer, epoch int, species int, compatThreshold float64) error {
	filePath := fmt.Sprintf("%s/champions.txt", outputDir)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open champions file: %v", err)
	}
	defer file.Close()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	org := champion.GetOrganism()
	line := fmt.Sprintf("%s EPOCH %d, Organism ID: %d, Fitness: %f (wins: %d, draws: %d), species: %d, compat threshold: %.2f\n",
		timestamp, epoch, org.Genotype.Id, org.Fitness, champion.GetWins(), champion.GetDraws(), species, compatThreshold)
	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("failed to write to champions file: %v", err)
	}
	return nil
}

// RunDebugGame runs a single game between the given players.
// Decision loggers should be set on individual players before calling this function via SetDecisionLogger.
func RunDebugGame(ctx context.Context, players []MonopolyPlayer, outputDir string) error {
	playerGroup, err := NewNEATPlayerGroup(0, players)
	if err != nil {
		return fmt.Errorf("failed to create player group: %w", err)
	}
	gameLogPath := filepath.Join(outputDir, fmt.Sprintf("game_%s", time.Now().Format("2006-01-02_15-04-05")))
	logger, err := NewTrainerLogger(gameLogPath, false)
	if err != nil {
		return fmt.Errorf("failed to create game logger: %w", err)
	}
	game := monopoly.NewGame(ctx, playerGroup, logger, 0)
	game.Start()
	return nil
}

func findBestPlayer(players []MonopolyPlayer) MonopolyPlayer {
	var bestPlayer MonopolyPlayer
	var bestFitness *float64
	for _, player := range players {
		org := player.GetOrganism()
		if bestFitness == nil || org.Fitness > *bestFitness {
			bestFitness = &org.Fitness
			bestPlayer = player
		}
	}
	return bestPlayer
}
