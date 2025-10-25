package neatnetwork

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
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
}

type RoundDetails struct {
	Epoch        int
	PlayersCount int
	Group        int
	Game         int
}

func NewMonopolyEvaluator(outputDir string, groupSize int) *MonopolyEvaluator {
	return &MonopolyEvaluator{
		outputDir: outputDir,
		groupSize: groupSize,
	}
}

func (e *MonopolyEvaluator) TournamentGenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {
	options, ok := neat.FromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to get options from context")
	}

	var players []*NEATMonopolyPlayer
	for i, org := range pop.Organisms {
		org.Fitness = 0
		org, err := NewNEATMonopolyPlayer(org)
		if err != nil {
			return fmt.Errorf("error creating NEATMonopolyPlayer for organism %d: %v", i, err)
		}
		players = append(players, org)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var rd = RoundDetails{
		Epoch: epoch.Id,
	}

	for len(players) > 1 {
		rd.PlayersCount = len(players)
		// neat.InfoLog(fmt.Sprintf("Starting round 1/%d\n", rd.PlayersCount))

		rng.Shuffle(len(players), func(i, j int) {
			players[i], players[j] = players[j], players[i]
		})

		var groups [][]*NEATMonopolyPlayer
		for i := 0; i < len(players); i += e.groupSize {
			end := i + e.groupSize
			if end > len(players) {
				end = len(players)
			}
			groups = append(groups, players[i:end])
		}

		players = []*NEATMonopolyPlayer{}
		for groupID, group := range groups {
			rd.Group = groupID
			winner, err := startTournamentGroup(ctx, rd, group, e.outputDir)
			if err != nil {
				neat.ErrorLog(fmt.Sprintf("Error in round 1/%d group %d: %v\n", rd.PlayersCount, groupID, err))
				return fmt.Errorf("error in round 1/%d group %d: %v", rd.PlayersCount, groupID, err)
			}
			players = append(players, winner)
		}
		// neat.InfoLog(fmt.Sprintf("\nRound 1/%d finished successfully; epoch %d\n", rd.PlayersCount, epoch.Id))
	}
	for _, org := range pop.Organisms {
		neat.DebugLog(fmt.Sprintf("Organism %d finished with fitness %f\n", org.Genotype.Id, org.Fitness))
	}
	epoch.FillPopulationStatistics(pop)

	if (epoch.Id+1)%cfg.PRINT_EVERY == 0 || epoch.Id == options.NumGenerations-1 {
		if _, err := utils.WritePopulationPlain(e.outputDir, pop, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump population, reason: %s\n", err))
			return err
		}
	}
	best := epoch.Champion
	numberOfSpecies := len(pop.Species)
	neat.InfoLog(fmt.Sprintf("Spieces count: %d\n", numberOfSpecies))
	neat.InfoLog(fmt.Sprintf("Champion of epoch %d is organism %d with fitness %f\n", epoch.Id, best.Genotype.Id, best.Fitness))
	neat.InfoLog(fmt.Sprintf("Number of nodes: %d, number of connections: %d\n", len(best.Genotype.Nodes), len(best.Genotype.Genes)))
	return nil
}

func startGame(ctx context.Context, rd RoundDetails, g []*NEATMonopolyPlayer, outputDir string, resultsCh chan struct {
	gameID int
	err    error
}) {
	defer func() {
		if r := recover(); r != nil {
			resultsCh <- struct {
				gameID int
				err    error
			}{gameID: rd.Game, err: fmt.Errorf("panic in group %d: %v", rd.Group, r)}
		}
	}()
	options, ok := neat.FromContext(ctx)
	if !ok {
		resultsCh <- struct {
			gameID int
			err    error
		}{gameID: rd.Game, err: fmt.Errorf("failed to get options from context")}
	}
	playerGroup, err := NewNEATPlayerGroup(rd.Group, g)
	if err != nil {
		resultsCh <- struct {
			gameID int
			err    error
		}{gameID: rd.Game, err: fmt.Errorf("error creating player group for group %d: %v", rd.Group, err)}
		return
	}
	enable_log := false
	enable_log = rd.PlayersCount == 4 && (rd.Epoch == options.NumGenerations-1 || (rd.Epoch+1)%cfg.PRINT_EVERY == 0)
	logger, err := NewTrainerLogger(fmt.Sprintf("%s/games/epoch%d/round1of%d/group%d/game%d.log",
		outputDir, rd.Epoch, rd.PlayersCount, rd.Group, rd.Game), !enable_log)
	if err != nil {
		resultsCh <- struct {
			gameID int
			err    error
		}{gameID: rd.Game, err: fmt.Errorf("error creating logger for group %d: %v", rd.Group, err)}
		return
	}
	game := monopoly.NewGame(ctx, playerGroup, logger, 0)
	game.Start()
	resultsCh <- struct {
		gameID int
		err    error
	}{gameID: rd.Game, err: nil}
}

func startTournamentGroup(ctx context.Context, round RoundDetails, group []*NEATMonopolyPlayer, outputDir string) (champion *NEATMonopolyPlayer, err error) {

	resultsCh := make(chan struct {
		gameID int
		err    error
	}, cfg.GAMES_PER_EPOCH)
	groupCopy := make([]*NEATMonopolyPlayer, len(group))
	copy(groupCopy, group)
	for g := range cfg.GAMES_PER_EPOCH {
		groupCopy = append(groupCopy[1:], groupCopy[0]) // rotate players
		round.Game = g
		go startGame(ctx, round, groupCopy, outputDir, resultsCh)
	}
	for range cfg.GAMES_PER_EPOCH {
		result := <-resultsCh
		if result.err != nil {
			neat.ErrorLog(fmt.Sprintf("Error in game %d of group %d: %v\n", result.gameID, round.Group, result.err))
			err = fmt.Errorf("error in game %d of group %d: %v", result.gameID, round.Group, result.err)
		}
	}
	if err != nil {
		return
	}
	slices.SortFunc(groupCopy, func(a, b *NEATMonopolyPlayer) int {
		return a.score - b.score
	})
	for i, player := range groupCopy {
		player.organism.Fitness += float64(i + 1)
		player.score = 0
	}
	return groupCopy[len(groupCopy)-1], nil
}

type GroupDetails struct {
	Epoch   int
	Round   int
	GroupID int
}

func (e *MonopolyEvaluator) GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {
	options, ok := neat.FromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to get options from context")
	}

	if _, err := utils.WritePopulationPlain(e.outputDir, pop, epoch); err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to dump population, reason: %s\n", err))
		return err
	}
	var players []*NEATMonopolyPlayer
	for i, org := range pop.Organisms {
		org.Fitness = 0
		org, err := NewNEATMonopolyPlayer(org)
		if err != nil {
			return fmt.Errorf("error creating NEATMonopolyPlayer for organism %d: %v", i, err)
		}
		players = append(players, org)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for roundID := range cfg.GAMES_PER_EPOCH {
		// neat.InfoLog(fmt.Sprintf("\nStarting round %d; epoch %d\n", roundID, epoch.Id))
		rng.Shuffle(len(players), func(i, j int) {
			players[i], players[j] = players[j], players[i]
		})

		var groups [][]*NEATMonopolyPlayer

		for i := 0; i < len(players); i += e.groupSize {
			end := i + e.groupSize
			if end > len(players) {
				end = len(players)
			}
			groups = append(groups, players[i:end])
		}
		dumpGroupAssignments(e.outputDir, epoch.Id, roundID, groups)
		resultsCh := make(chan struct {
			groupID int
			err     error
		}, len(groups))
		var groupID int = 0
		for groupID < len(groups) {
			threads := 1
			for groupID < len(groups) && threads%cfg.MAX_THREADS != 0 {
				group := groups[groupID]
				// fmt.Printf("Starting group %d\n", groupID)
				groupDetails := GroupDetails{
					Epoch:   epoch.Id,
					Round:   roundID,
					GroupID: groupID,
				}
				go startGroup(ctx, groupDetails, group, e.outputDir, resultsCh)

				threads++
				groupID++
			}
			var err error
			for i := 0; i+1 < threads; i++ {
				result := <-resultsCh
				if result.err != nil {
					fmt.Printf("[%d] Error in group %d: %v\n", i, result.groupID, result.err)
					err = fmt.Errorf("error in one of the groups: %v", result.err)
				} else {
					// fmt.Printf("[%d] Group %d finished successfully\n", i, result.groupID)
				}
			}
			if err != nil {
				return err
			}
		}

		// neat.InfoLog(fmt.Sprintf("\nRound %d finished successfully; epoch %d\n", roundID, epoch.Id))
	}
	// for _, org := range pop.Organisms {
	// 	neat.InfoLog(fmt.Sprintf("Organism %d finished with fitness %f\n", org.Genotype.Id, org.Fitness))
	// }
	epoch.FillPopulationStatistics(pop)
	if (epoch.Id+1)%cfg.PRINT_EVERY == 0 || epoch.Id == options.NumGenerations-1 {
		if _, err := utils.WritePopulationPlain(e.outputDir, pop, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump population, reason: %s\n", err))
			return err
		}
	}
	best := epoch.Champion
	numberOfSpecies := len(pop.Species)
	neat.InfoLog(fmt.Sprintf("Spieces count: %d\n", numberOfSpecies))
	neat.InfoLog(fmt.Sprintf("Champion of epoch %d is organism %d with fitness %f\n", epoch.Id, best.Genotype.Id, best.Fitness))
	// freeze of 0.5sec
	// time.Sleep(500 * time.Millisecond)
	return nil
}

func startGroup(ctx context.Context, gd GroupDetails, g []*NEATMonopolyPlayer, outputDir string, resultsCh chan struct {
	groupID int
	err     error
}) {
	defer func() {
		if r := recover(); r != nil {
			resultsCh <- struct {
				groupID int
				err     error
			}{groupID: gd.GroupID, err: fmt.Errorf("panic in group %d: %v", gd.GroupID, r)}
		}
	}()
	options, ok := neat.FromContext(ctx)
	if !ok {
		resultsCh <- struct {
			groupID int
			err     error
		}{groupID: gd.GroupID, err: fmt.Errorf("failed to get options from context")}
	}
	playerGroup, err := NewNEATPlayerGroup(gd.GroupID, g)
	if err != nil {
		resultsCh <- struct {
			groupID int
			err     error
		}{groupID: gd.GroupID, err: fmt.Errorf("error creating player group for group %d: %v", gd.GroupID, err)}
		return
	}
	enable_log := false
	if gd.Epoch == options.NumGenerations-1 {
		enable_log = true
	} else if (gd.Epoch+1)%cfg.PRINT_EVERY == 0 {
		enable_log = gd.Round == 0
	}
	logger, err := NewTrainerLogger(fmt.Sprintf("%s/games/epoch%d/round%d/group%d",
		outputDir, gd.Epoch, gd.Round, gd.GroupID), !enable_log)
	if err != nil {
		resultsCh <- struct {
			groupID int
			err     error
		}{groupID: gd.GroupID, err: fmt.Errorf("error creating logger for group %d: %v", gd.GroupID, err)}
		return
	}
	game := monopoly.NewGame(ctx, playerGroup, logger, 0)
	game.Start()
	for _, player := range g {
		// fmt.Printf("Player %d score: %d\n", player.organism.Genotype.Id, player.score)
		player.organism.Fitness += float64(player.score)
		player.score = 0
	}
	resultsCh <- struct {
		groupID int
		err     error
	}{groupID: gd.GroupID, err: nil}
}

func dumpGroupAssignments(outputDir string, epoch int, round int, groups [][]*NEATMonopolyPlayer) {
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
			fmt.Fprintf(file, "\tPlayer %d", player.organism.Genotype.Id)
			players[player.organism.Genotype.Id] = i
		}
		fmt.Fprintln(file)
	}
	for i := 0; i < len(players); i++ {
		fmt.Fprintf(file, "Player %d: Group %d\n", i, players[i])
	}
}
