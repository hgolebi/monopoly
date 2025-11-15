package neatnetwork

import (
	"context"
	"fmt"
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

type RoundDetails struct {
	Epoch        int
	PlayersCount int
	Group        int
	Game         int
}

type MonopolyEvaluator struct {
	outputDir        string
	groupSize        int
	lastChampion     *genetics.Organism
	lastChampFitness float64
}

func NewMonopolyEvaluator(outputDir string, groupSize int) *MonopolyEvaluator {
	return &MonopolyEvaluator{
		outputDir: outputDir,
		groupSize: groupSize,
	}
}

type GroupDetails struct {
	Epoch   int
	Round   int
	GroupID int
	Players []MonopolyPlayer
}

func (e *MonopolyEvaluator) GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	options, ok := neat.FromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to get options from context")
	}

	// create players from population
	players, err := e.createPlayersFromPopulation(pop)
	if err != nil {
		return fmt.Errorf("failed to create players from population: %v", err)
	}

	// start workers
	jobsCh := make(chan GroupDetails, 100)
	var wg sync.WaitGroup
	for i := 0; i < cfg.MAX_THREADS; i++ {
		wg.Add(1)
		go startWorker(ctx, i, jobsCh, &wg, e.outputDir)
	}

	// starting rounds
	for roundID := range cfg.GAMES_PER_EPOCH {
		// prepare groups
		rng.Shuffle(len(players), func(i, j int) {
			players[i], players[j] = players[j], players[i]
		})
		var groups [][]MonopolyPlayer
		for i := 0; i < len(players); i += (e.groupSize - 1) {
			end := min(i+e.groupSize-1, len(players))
			// Create a new slice for the group to avoid modifying the underlying 'players' slice
			group := make([]MonopolyPlayer, 0, e.groupSize)
			group = append(group, players[i:end]...)
			group = append(group, new(SimplePlayerBot))
			rng.Shuffle(len(group), func(i, j int) {
				group[i], group[j] = group[j], group[i]
			})
			groups = append(groups, group)
		}
		if roundID == 0 && (epoch.Id == options.NumGenerations-1 || (epoch.Id+1)%cfg.PRINT_EVERY == 0) {
			dumpGroupAssignments(e.outputDir, epoch.Id, roundID, groups)
		}

		// create job for every group
		for groupID, group := range groups {

			gd := GroupDetails{
				Epoch:   epoch.Id,
				Round:   roundID,
				GroupID: groupID,
				Players: group,
			}
			jobsCh <- gd
		}
	}
	close(jobsCh)
	wg.Wait()

	e.calculateFitness(players)
	epoch.FillPopulationStatistics(pop)
	best := epoch.Champion
	numberOfSpecies := len(pop.Species)
	e.lastChampion = best
	e.lastChampFitness = best.Fitness

	// log info
	neat.InfoLog(fmt.Sprintf("Species count: %d\n", numberOfSpecies))
	neat.InfoLog(fmt.Sprintf("Champion of epoch %d is organism %d\n with fitness: %f", epoch.Id, best.Genotype.Id, best.Fitness))
	neat.InfoLog(fmt.Sprintf("Number of nodes: %d, number of connections: %d\n", len(best.Genotype.Nodes), len(best.Genotype.Genes)))

	// dump population
	if (epoch.Id+1)%cfg.PRINT_EVERY == 0 || epoch.Id == options.NumGenerations-1 {
		if _, err := utils.WritePopulationPlain(e.outputDir, pop, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump population, reason: %s\n", err))
			return err
		}
	} else {
		// // dump only champion
		// genomeFile := fmt.Sprintf("gen_%d_champion", epoch.Id)
		// if _, err := utils.WriteGenomePlain(genomeFile, e.outputDir, best, epoch); err != nil {
		// 	neat.ErrorLog(fmt.Sprintf("Failed to dump champion organism, reason: %s\n", err))
		// 	return err
		// }
	}
	return nil
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
	playerGroup, err = NewNEATPlayerGroup(gd.GroupID, gd.Players)
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
	game := monopoly.NewGame(ctx, playerGroup, logger, 0)
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
	// if e.lastChampion == nil {
	// 	org, err := NewNEATMonopolyPlayer(pop.Organisms[0])
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error creating NEATMonopolyPlayer for last champion (duplication): %v", err)
	// 	}
	// 	players = append(players, org)
	// } else {
	// 	org, err := NewNEATMonopolyPlayer(e.lastChampion)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error creating NEATMonopolyPlayer for last champion: %v", err)
	// 	}
	// 	players = append(players, org)
	// }

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

func (e *MonopolyEvaluator) calculateFitness(players []MonopolyPlayer) {
	// var lastChampFitness float64
	// var lastChampScore float64
	// if e.lastChampion == nil {
	// 	lastChampFitness = 1.0
	// 	lastChampScore = 1.0
	// } else {
	// 	lastChampFitness = e.lastChampFitness
	// 	lastChampScore = float64(players[0].GetScore()) / cfg.GAMES_PER_EPOCH
	// }

	// highestScore := 0.0
	// for _, player := range players[1:] {
	// 	org := player.GetOrganism()
	// 	if org == nil {
	// 		continue
	// 	}
	// 	score := float64(player.GetScore()) / cfg.GAMES_PER_EPOCH
	// 	distance := score - lastChampScore
	// 	org.Fitness += lastChampFitness + distance/100.0
	// 	highestScore = math.Max(highestScore, score)
	// }

	for _, player := range players {
		org := player.GetOrganism()
		if org == nil {
			continue
		}
		org.Fitness += float64(player.GetScore()) / cfg.GAMES_PER_EPOCH
	}
}
