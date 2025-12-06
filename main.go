package main

import (
	"context"
	"flag"
	"log"
	"monopoly/pkg/consoleCLI"
	"monopoly/pkg/monopoly"
	neatnetwork "monopoly/pkg/neat"
	"monopoly/pkg/server"

	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

func runConsoleMonopoly() {
	cliMode := flag.Bool("cli", false, "run in CLI client mode")
	flag.Parse()
	if *cliMode {
		consoleCLI.StartClient()
		return
	}
	neat.InitLogger("debug")
	bots := []server.PlayerIO{
		loadNEATPlayer(),
		loadNEATPlayer(),
		loadNEATPlayer(),
	}
	io := server.NewConsoleServer(1, bots)
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

func main() {
	// trainNEATNetwork()
	runConsoleMonopoly()
}

func loadNEATPlayer() *neatnetwork.NEATMonopolyPlayer {
	filePath := ".\\genomes\\trained"
	// if len(os.Args) < 2 {
	// 	fmt.Println("Usage: go run graph.go <genome_file_path>")
	// } else {
	// 	filePath = os.Args[1]
	// }
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
