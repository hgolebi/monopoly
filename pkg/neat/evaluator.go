package neatnetwork

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	cfg "monopoly/pkg/config"
	"monopoly/pkg/monopoly"

	"github.com/yaricom/goNEAT/v4/experiment"
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
	var players []*NEATMonopolyPlayer
	for i, org := range pop.Organisms {
		org.Fitness = 0
		org, err := NewNEATMonopolyPlayer(org)
		if err != nil {
			return fmt.Errorf("error creating NEATMonopolyPlayer for organism %d: %v", i, err)
		}
		players = append(players, org)
	}
	epochDir := fmt.Sprintf("%s/epoch%d", e.outputDir, epoch.Id)
	if err := os.MkdirAll(epochDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating epoch directory: %v", err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for roundID := range cfg.GAMES_PER_EPOCH {
		roundDir := fmt.Sprintf("%s/epoch%d/round%d", e.outputDir, epoch.Id, roundID)
		if err := os.MkdirAll(roundDir, os.ModePerm); err != nil {
			return fmt.Errorf("error creating round directory: %v", err)
		}

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

		errCh := make(chan error, len(groups))
		for groupID, group := range groups {
			go func(g []*NEATMonopolyPlayer) {
				// defer func() {
				// 	if r := recover(); r != nil {
				// 		errCh <- fmt.Errorf("panic in group %d: %v", groupID, r)
				// 	}
				// }()
				playerGroup, err := NewNEATPlayerGroup(groupID, g)
				if err != nil {
					errCh <- fmt.Errorf("error creating player group for group %d: %v", groupID, err)
					return
				}
				logger, err := NewTrainerLogger(fmt.Sprintf("%s/epoch%d/round%d/group%d.log", e.outputDir, epoch.Id, roundID, groupID))
				if err != nil {
					errCh <- fmt.Errorf("error creating logger for group %d: %v", groupID, err)
					return
				}
				game := monopoly.NewGame(playerGroup, logger, 0)
				game.Start()
				errCh <- nil
			}(group)
		}

		for i := 0; i < len(groups); i++ {
			if err := <-errCh; err != nil {
				return fmt.Errorf("error in group %d: %v", i, err)
			} else {
				fmt.Printf("Group %d finished successfully\n", i)
			}
		}
	}
	return nil
}
