package neatnetwork

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	cfg "monopoly/pkg/config"
	"os"
	"os/signal"
	"syscall"

	"github.com/yaricom/goNEAT/v4/experiment"
	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

func TrainNetwork(seed int64, neatOptionsFile string, genomeFile string, outputDir string) {
	rng := rand.New(rand.NewSource(seed))
	neatOptions, err := neat.ReadNeatOptionsFromFile(neatOptionsFile)
	if err != nil {
		log.Fatal("Failed to load NEAT options:", err)
	}
	genomeReader, err := genetics.NewGenomeReaderFromFile(genomeFile)
	if err != nil {
		log.Fatal("Failed to create genome reader:", err)
	}
	startGenome, err := genomeReader.Read()
	if err != nil {
		log.Fatal("Failed to read start genome:", err)
	}

	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatal("Failed to create output directory:", err)
	}
	exp := experiment.Experiment{
		Id:       0,
		Trials:   make(experiment.Trials, neatOptions.NumRuns),
		RandSeed: seed,
	}
	evaluator := NewMonopolyEvaluator(outputDir, cfg.GROUP_SIZE, rng)
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err = exp.Execute(neat.NewContext(ctx, neatOptions), startGenome, evaluator, nil)
		if err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
	}()

	go func(cancel context.CancelFunc) {
		fmt.Println("Press ctrl+C to stop the experiment...")
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
		select {
		case <-signalChan:
			cancel()
		case <-errChan:
			return
		}
	}(cancel)
	err = <-errChan
	if err != nil {
		log.Fatal("Experiment failed:", err)
	}
	fmt.Println("Experiment completed successfully.")
	exp.PrintStatistics()
	best, epoch, ok := exp.BestOrganism(false)
	if ok {
		fmt.Printf("Best organism found in epoch %d: ID %d with fitness %f\n", epoch, best.Genotype.Id, best.Fitness)
	}
}
