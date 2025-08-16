package neatnetwork

import (
	"context"
	"fmt"
	"math/rand"
	"monopoly/pkg/monopoly"
	"time"

	"github.com/yaricom/goNEAT/v4/experiment"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

type MonopolyEvaluator struct {
	outputDir string
	groupSize int
}

func (e *MonopolyEvaluator) GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {
	organisms := make([]*genetics.Organism, len(pop.Organisms))
	copy(organisms, pop.Organisms)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(organisms), func(i, j int) {
		organisms[i], organisms[j] = organisms[j], organisms[i]
	})

	var groups [][]*genetics.Organism

	for i := 0; i < len(organisms); i += e.groupSize {
		end := i + e.groupSize
		if end > len(organisms) {
			end = len(organisms)
		}
		groups = append(groups, organisms[i:end])
	}

	errCh := make(chan error, len(groups))
	for idx, group := range groups {
		go func(g []*genetics.Organism) {
			defer func() {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("panic in group %d: %v", idx, r)
				}
			}()
			playerGroup, err := NewNEATPlayerGroup(idx, g)
			if err != nil {
				errCh <- fmt.Errorf("error creating player group for group %d: %v", idx, err)
				return
			}
			logger, err := NewTrainerLogger(fmt.Sprintf("%s/epoch%d/group%d.log", e.outputDir, epoch.Id, idx))
			if err != nil {
				errCh <- fmt.Errorf("error creating logger for group %d: %v", idx, err)
				return
			}
			game := monopoly.NewGame(playerGroup, logger, 0)
			game.Start()
		}(group)
	}
	atLeastOneSuccess := false
	for i := 0; i < len(groups); i++ {
		if err := <-errCh; err != nil {
			fmt.Println(err)
		} else {
			atLeastOneSuccess = true
		}
	}
	if !atLeastOneSuccess {
		return fmt.Errorf("all groups failed")
	}
	return nil
}
