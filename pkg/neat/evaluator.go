package neatnetwork

import (
	"context"
	"fmt"
	"math/rand"
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

func NewMonopolyEvaluator(outputDir string, groupSize int) *MonopolyEvaluator {
	return &MonopolyEvaluator{
		outputDir: outputDir,
		groupSize: groupSize,
	}
}

func (e *MonopolyEvaluator) GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {
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
	// epochDir := fmt.Sprintf("%s/epoch%d", e.outputDir, epoch.Id)
	// if err := os.MkdirAll(epochDir, os.ModePerm); err != nil {
	// 	return fmt.Errorf("error creating epoch directory: %v", err)
	// }

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for roundID := range cfg.GAMES_PER_EPOCH {
		// roundDir := fmt.Sprintf("%s/epoch%d/round%d", e.outputDir, epoch.Id, roundID)
		// if err := os.MkdirAll(roundDir, os.ModePerm); err != nil {
		// 	return fmt.Errorf("error creating round directory: %v", err)
		// }
		neat.InfoLog(fmt.Sprintf("\nStarting round %d; epoch %d\n", roundID, epoch.Id))
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

		resultsCh := make(chan struct {
			groupID int
			err     error
		}, len(groups))
		var groupID int = 0
		for groupID < len(groups) {
			threads := 1
			for groupID < len(groups) && threads%cfg.MAX_THREADS != 0 {
				group := groups[groupID]
				fmt.Printf("Starting group %d\n", groupID)
				go startGroup(ctx, roundID, groupID, epoch.Id, group, e, resultsCh)

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
					fmt.Printf("[%d] Group %d finished successfully\n", i, result.groupID)
				}
			}
			if err != nil {
				return err
			}
		}

		neat.InfoLog(fmt.Sprintf("\nRound %d finished successfully; epoch %d\n", roundID, epoch.Id))
	}
	for _, org := range pop.Organisms {
		neat.InfoLog(fmt.Sprintf("Organism %d finished with fitness %f\n", org.Genotype.Id, org.Fitness))
	}
	epoch.FillPopulationStatistics(pop)
	best := epoch.Champion
	neat.InfoLog(fmt.Sprintf("Champion of epoch %d is organism %d with fitness %f\n", epoch.Id, best.Genotype.Id, best.Fitness))
	return nil
}

func startGroup(ctx context.Context, roundID int, groupID int, epochId int, g []*NEATMonopolyPlayer, e *MonopolyEvaluator, resultsCh chan struct {
	groupID int
	err     error
}) {
	defer func() {
		if r := recover(); r != nil {
			resultsCh <- struct {
				groupID int
				err     error
			}{groupID: groupID, err: fmt.Errorf("panic in group %d: %v", groupID, r)}
		}
	}()
	options, ok := neat.FromContext(ctx)
	if !ok {
		resultsCh <- struct {
			groupID int
			err     error
		}{groupID: groupID, err: fmt.Errorf("failed to get options from context")}
	}
	playerGroup, err := NewNEATPlayerGroup(groupID, g)
	if err != nil {
		resultsCh <- struct {
			groupID int
			err     error
		}{groupID: groupID, err: fmt.Errorf("error creating player group for group %d: %v", groupID, err)}
		return
	}

	logger, err := NewTrainerLogger(fmt.Sprintf("%s/games/temp/round%d/group%d.log", e.outputDir, roundID, groupID), epochId != options.NumGenerations-1)
	if err != nil {
		resultsCh <- struct {
			groupID int
			err     error
		}{groupID: groupID, err: fmt.Errorf("error creating logger for group %d: %v", groupID, err)}
		return
	}
	game := monopoly.NewGame(ctx, playerGroup, logger, 0)
	game.Start()
	resultsCh <- struct {
		groupID int
		err     error
	}{groupID: groupID, err: nil}
}
