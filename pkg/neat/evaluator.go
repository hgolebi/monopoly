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
	rng              *rand.Rand
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

	// start workers
	jobsCh := make(chan GroupDetails, 100)
	var wg sync.WaitGroup
	for i := range cfg.MAX_THREADS {
		wg.Add(1)
		go startWorker(ctx, i, jobsCh, &wg, e.outputDir)
	}

	// starting rounds
	for roundID := range cfg.GAMES_PER_EPOCH {
		// prepare groups
		var groups [][]MonopolyPlayer
		if cfg.INCLUDE_HEURISTIC_BOT {
			groups = e.prepareGroupsWithHeuristicBot(players, e.groupSize)
		} else {
			groups = e.prepareGroups(players, e.groupSize)
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

	bestPlayer := e.calculateFitness(players)
	bestOrg := bestPlayer.GetOrganism()
	epoch.FillPopulationStatistics(pop)
	numberOfSpecies := len(pop.Species)
	// e.lastChampion = best.GetOrganism()
	// e.lastChampFitness = e.lastChampion.Fitness

	// log info
	neat.InfoLog(fmt.Sprintf("Species count: %d\n", numberOfSpecies))
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
	if err := appendChampionInfo(e.outputDir, bestPlayer, epoch.Id); err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to append champion info, reason: %s\n", err))
		return err
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

func (e *MonopolyEvaluator) prepareGroups(players []MonopolyPlayer, groupSize int) [][]MonopolyPlayer {
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

func (e *MonopolyEvaluator) prepareGroupsWithHeuristicBot(players []MonopolyPlayer, groupSize int) [][]MonopolyPlayer {
	e.rng.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
	var groups [][]MonopolyPlayer
	for i := 0; i < len(players); i += (e.groupSize - 1) {
		end := min(i+e.groupSize-1, len(players))
		group := make([]MonopolyPlayer, 0, e.groupSize)
		group = append(group, players[i:end]...)
		group = append(group, new(SimplePlayerBot))
		e.rng.Shuffle(len(group), func(i, j int) {
			group[i], group[j] = group[j], group[i]
		})
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

func (e *MonopolyEvaluator) calculateFitness(players []MonopolyPlayer) (best MonopolyPlayer) {
	highestFitness := 0.0
	for _, player := range players {
		org := player.GetOrganism()
		if org == nil {
			continue
		}
		org.Fitness += float64(player.GetScore()) / cfg.GAMES_PER_EPOCH
		if org.Fitness > highestFitness {
			highestFitness = org.Fitness
			best = player
		}
	}
	return best
}

func appendChampionInfo(outputDir string, champion MonopolyPlayer, epoch int) error {
	filePath := fmt.Sprintf("%s/champions.txt", outputDir)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open champions file: %v", err)
	}
	defer file.Close()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	org := champion.GetOrganism()
	line := fmt.Sprintf("%s EPOCH %d, Organism ID: %d, Fitness: %f\n (wins: %d, draws: %d)", timestamp, epoch, org.Genotype.Id, org.Fitness, champion.GetWins(), champion.GetDraws())
	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("failed to write to champions file: %v", err)
	}
	return nil
}
