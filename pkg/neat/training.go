package neatnetwork

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	cfg "monopoly/pkg/config"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/yaricom/goNEAT/v4/experiment"
	"github.com/yaricom/goNEAT/v4/neat"
)

// TrainNetwork runs a NEAT experiment. If genomeFile is non-empty the starting genome is loaded
// from that file; otherwise NewStartGenome(1, allToAll) is used to build one automatically.
func TrainNetwork(seed int64, neatOptionsFile string, outputDir string, allToAll bool, genomeFile string) {
	rng := rand.New(rand.NewSource(seed))
	// goNEAT draws mutations and crossovers from the global math/rand source and does not seed it with RandSeed
	rand.Seed(seed) //nolint:staticcheck // still effective, go.mod targets go 1.23
	neatOptions, err := neat.ReadNeatOptionsFromFile(neatOptionsFile)
	if err != nil {
		log.Fatal("Failed to load NEAT options:", err)
	}
	if err := checkGroupSize(neatOptions.PopSize, cfg.GROUP_SIZE); err != nil {
		log.Fatal("Invalid NEAT options: ", err)
	}
	startGenome := loadOrBuildGenome(genomeFile, allToAll)

	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatal("Failed to create output directory:", err)
	}
	if err := writeRunInfo(outputDir, seed, neatOptionsFile, genomeFile, allToAll); err != nil {
		log.Fatal("Failed to save run info:", err)
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

// writeRunInfo saves what is needed to reproduce the run to the output directory: run_info.txt with the seed and
// the command, and options_used.yaml, a copy of the NEAT options (the options file may change after the run).
func writeRunInfo(outputDir string, seed int64, neatOptionsFile string, genomeFile string, allToAll bool) error {
	options, err := os.ReadFile(neatOptionsFile)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "options_used.yaml"), options, 0644); err != nil {
		return err
	}

	startGenome := fmt.Sprintf("built from INPUT_COUNT/OUTPUT_COUNT, all-to-all: %v", allToAll)
	quote := func(path string) string { return `"` + filepath.ToSlash(path) + `"` }
	command := fmt.Sprintf("go run . train -seed %d -options %s -out %s -all-to-all=%v", seed, quote(neatOptionsFile), quote(outputDir), allToAll)
	if genomeFile != "" {
		startGenome = genomeFile
		command = fmt.Sprintf("go run . train -seed %d -options %s -out %s -genome %s", seed, quote(neatOptionsFile), quote(outputDir), quote(genomeFile))
	}
	info := fmt.Sprintf("seed: %d\nstarted: %s\noptions: %s (copy: options_used.yaml)\nstart genome: %s\nreproduce: %s\n",
		seed, time.Now().Format("2006-01-02 15:04:05"), neatOptionsFile, startGenome, command)
	return os.WriteFile(filepath.Join(outputDir, "run_info.txt"), []byte(info), 0644)
}
