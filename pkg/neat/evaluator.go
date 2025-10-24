package neatnetwork

import (
	"context"
	"fmt"
	"math/rand"
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

func (e *MonopolyEvaluator) GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {
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
			winner, err := startGroup(ctx, rd, group, e.outputDir)
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

func startGroup(ctx context.Context, round RoundDetails, group []*NEATMonopolyPlayer, outputDir string) (champion *NEATMonopolyPlayer, err error) {
	// neat.InfoLog(fmt.Sprintf("Group %d\n", round.Group))
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
