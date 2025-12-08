package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"monopoly/pkg/consoleCLI"
	"monopoly/pkg/monopoly"
	neatnetwork "monopoly/pkg/neat"
	"monopoly/pkg/server"

	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

func main() {
	runConsoleMonopoly()
}

func runConsoleMonopoly() {
	cliMode := flag.Bool("cli", false, "run in CLI client mode")
	flag.Parse()
	if *cliMode {
		consoleCLI.StartClient()
		return
	}
	neat.InitLogger("error")
	bots := []server.PlayerIO{
		loadNEATPlayer("./genomes/trained"),
		loadNEATPlayer("./genomes/trained"),
		loadNEATPlayer("./genomes/trained"),
		loadNEATPlayer("./genomes/trained"),
	}

	// Get number of human players from user
	var numHumanPlayers int
	for {
		fmt.Print("Enter number of human players (0-4): ")
		_, err := fmt.Scan(&numHumanPlayers)
		if err != nil || numHumanPlayers < 0 || numHumanPlayers > 4 {
			fmt.Println("Invalid input for number of human players")
			fmt.Println("Number must be between 0 and 4")
			continue
		}
		break
	}

	io := server.NewConsoleServer(numHumanPlayers, bots[:4-numHumanPlayers])
	logger := monopoly.ConsoleLogger{}
	logger.Init()
	ctx := context.Background()
	game := monopoly.NewGame(ctx, io, &logger, 0)
	game.Start()

}

func trainNEATNetwork() {
	neatOptionsFile := "neat_options.yaml"
	neatGenomeFile := "./genomes/base_genome.yaml"
	outputDir := "output"
	neatnetwork.TrainNetwork(0, neatOptionsFile, neatGenomeFile, outputDir)
}

func loadNEATPlayer(filePath string) *neatnetwork.NEATMonopolyPlayer {
	genomeReader, err := genetics.NewGenomeReaderFromFile(filePath)
	if err != nil {
		log.Fatal("Failed to create genome reader:", err)
	}
	genome, err := genomeReader.Read()
	if err != nil {
		log.Fatal("Failed to read genome:", err)
	}
	organism, err := genetics.NewOrganism(0.0, genome, 0)
	if err != nil {
		log.Fatal("Failed to create organism from genome:", err)
	}
	bot, err := neatnetwork.NewNEATMonopolyPlayer(organism)
	if err != nil {
		log.Fatal("Failed to create NEAT player from organism:", err)
	}
	return bot
}
