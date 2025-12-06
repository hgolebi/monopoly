package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"monopoly/pkg/config"
	"monopoly/pkg/consoleCLI"
	"monopoly/pkg/monopoly"
	neatnetwork "monopoly/pkg/neat"
	"monopoly/pkg/server"
	"time"

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
	filePath := ".\\genomes\\trained"
	neat.InitLogger("debug")
	bots := []server.PlayerIO{
		loadNEATPlayer(filePath),
		loadNEATPlayer(filePath),
		loadNEATPlayer(filePath),
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
	trainNEATNetwork()
	// runConsoleMonopoly()
	// runBotMatch()
}

func loadNEATPlayer(filePath string) *neatnetwork.NEATMonopolyPlayer {

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

func runBotMatch() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	e := neatnetwork.NewMonopolyEvaluator("experiment", 4, rng)

	bot1 := loadNEATPlayer(".\\genomes\\trained")
	bot2 := loadNEATPlayer(".\\genomes\\100_wins")
	bot3 := loadNEATPlayer(".\\genomes\\draw_machine")
	bot4 := loadNEATPlayer(".\\genomes\\bracket133")
	// bot5 := neatnetwork.SimplePlayerBot{}

	players := []neatnetwork.MonopolyPlayer{bot1, bot2, bot3, bot4}

	neatOptionsFile := "neat_options.yaml"
	neatOptions, err := neat.ReadNeatOptionsFromFile(neatOptionsFile)
	if err != nil {
		log.Fatal("Failed to load NEAT options:", err)
	}
	ctx := neat.NewContext(context.Background(), neatOptions)
	err = e.PlayRound(ctx, players, -1)
	if err != nil {
		panic(fmt.Sprintf("Error during bot match: %v", err))
	}

	fmt.Printf("Games played: %d\n", config.GAMES_PER_EPOCH)
	fmt.Printf("Trained: AvgScore=%d Wins=%d Draws=%d\n", bot1.GetScore()/config.GAMES_PER_EPOCH, bot1.GetWins(), bot1.GetDraws())
	fmt.Printf("100_wins: AvgScore=%d Wins=%d Draws=%d\n", bot2.GetScore()/config.GAMES_PER_EPOCH, bot2.GetWins(), bot2.GetDraws())
	fmt.Printf("draw_machine: AvgScore=%d Wins=%d Draws=%d\n", bot3.GetScore()/config.GAMES_PER_EPOCH, bot3.GetWins(), bot3.GetDraws())
	fmt.Printf("bracket133: AvgScore=%d Wins=%d Draws=%d\n", bot4.GetScore()/config.GAMES_PER_EPOCH, bot4.GetWins(), bot4.GetDraws())
	// fmt.Printf("SimpleBot: AvgScore=%d Wins=%d Draws=%d\n", bot5.GetScore()/config.GAMES_PER_EPOCH, bot5.GetWins(), bot5.GetDraws())
}
