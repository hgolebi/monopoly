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
	// trainNEATNetwork()
	// runConsoleMonopoly()
	runBotMatch()
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
	e := neatnetwork.NewMonopolyEvaluator("experiment", config.GROUP_SIZE, rng)

	bot1_name := "first_good"
	bot2_name := "bracket_with_bot767"
	bot3_name := "bracket133"
	bot4_name := "bracket796"

	bot1 := loadNEATPlayer(".\\genomes\\" + bot1_name)
	bot2 := loadNEATPlayer(".\\genomes\\" + bot2_name)
	bot3 := loadNEATPlayer(".\\genomes\\" + bot3_name)
	bot4 := loadNEATPlayer(".\\genomes\\" + bot4_name)

	simpleBot := neatnetwork.SimplePlayerBot{}

	players := []neatnetwork.MonopolyPlayer{bot2, bot3, &simpleBot}

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
	fmt.Printf("%s: AvgScore=%d Wins=%d Draws=%d\n", bot1_name, bot1.GetScore()/config.GAMES_PER_EPOCH, bot1.GetWins(), bot1.GetDraws())
	fmt.Printf("%s: AvgScore=%d Wins=%d Draws=%d\n", bot2_name, bot2.GetScore()/config.GAMES_PER_EPOCH, bot2.GetWins(), bot2.GetDraws())
	fmt.Printf("%s: AvgScore=%d Wins=%d Draws=%d\n", bot3_name, bot3.GetScore()/config.GAMES_PER_EPOCH, bot3.GetWins(), bot3.GetDraws())
	fmt.Printf("%s: AvgScore=%d Wins=%d Draws=%d\n", bot4_name, bot4.GetScore()/config.GAMES_PER_EPOCH, bot4.GetWins(), bot4.GetDraws())

	fmt.Printf("simpleBot: AvgScore=%d Wins=%d Draws=%d\n", simpleBot.GetScore()/config.GAMES_PER_EPOCH, simpleBot.GetWins(), simpleBot.GetDraws())
}
