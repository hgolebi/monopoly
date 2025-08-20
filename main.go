package main

import (
	"context"
	"flag"
	"monopoly/pkg/consoleCLI"
	"monopoly/pkg/monopoly"
	neatnetwork "monopoly/pkg/neat"
)

func runConsoleMonopoly() {
	cliMode := flag.Bool("cli", false, "run in CLI client mode")
	flag.Parse()
	io := monopoly.ConsoleServer{}
	logger := monopoly.ConsoleLogger{}
	logger.Init()
	ctx := context.Background()
	if *cliMode {
		consoleCLI.StartClient()
	} else {
		game := monopoly.NewGame(ctx, &io, &logger, 0)
		game.Start()
	}
}

func trainNEATNetwork() {
	neatOptionsFile := "neat_options.yaml"
	neatGenomeFile := "genome.yaml"
	outputDir := "output"
	neatnetwork.TrainNetwork(0, neatOptionsFile, neatGenomeFile, outputDir)
}

func main() {
	trainNEATNetwork()
	// runConsoleMonopoly()
}
