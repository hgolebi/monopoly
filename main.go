package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"math/rand"
	randv2 "math/rand/v2"
	"monopoly/pkg/config"
	"monopoly/pkg/consoleCLI"
	"monopoly/pkg/monopoly"
	neatnetwork "monopoly/pkg/neat"
	"monopoly/pkg/server"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

// command is a mode of the program, selected by the first argument: go run . <command> [flags]
type command struct {
	name        string
	description string
	run         func(args []string) error
}

var commands = []command{
	{"play", "game server: human players (console clients) against bots", runPlay},
	{"client", "console client of a human player (connects to the game server)", runClient},
	{"train", "NEAT training", runTrain},
	{"match", "bots play GAMES_PER_EPOCH games against each other, prints scores", runMatch},
	{"debug", "one game of bots with game and decision logs", runDebug},
}

var defaultGenome = filepath.Join("genomes", "bestForNow")

var defaultGenomes = strings.Join([]string{
	filepath.Join("genomes", "bestForNow"),
	filepath.Join("genomes", "bestForNow copy"),
	filepath.Join("genomes", "bestForNow copy 2"),
	filepath.Join("genomes", "bestForNow copy 3"),
}, ",")

func main() {
	if len(os.Args) < 2 || os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printUsage()
		if len(os.Args) < 2 {
			os.Exit(2)
		}
		return
	}
	for _, cmd := range commands {
		if cmd.name == os.Args[1] {
			if err := cmd.run(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", cmd.name, err)
				os.Exit(1)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
	printUsage()
	os.Exit(2)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: go run . <command> [flags]")
	fmt.Fprintln(os.Stderr, "\nCommands:")
	for _, cmd := range commands {
		fmt.Fprintf(os.Stderr, "  %-8s %s\n", cmd.name, cmd.description)
	}
	fmt.Fprintln(os.Stderr, "\nFlags of a command: go run . <command> -h. Details: README.md")
}

func newFlagSet(name string, usage string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: go run . %s [flags]\n%s\n\nFlags:\n", name, usage)
		fs.PrintDefaults()
	}
	return fs
}

func runPlay(args []string) error {
	fs := newFlagSet("play", "Starts the game server on :12345. Every human player connects with: go run . client")
	humans := fs.Int("humans", -1, "number of human players (0-4); asked interactively when not given")
	genome := fs.String("genome", defaultGenome, "genome file of the bots (all bots use it)")
	_ = fs.Parse(args)

	if *humans < 0 {
		*humans = askHumanPlayers()
	}
	if *humans > config.MAX_PLAYERS {
		return fmt.Errorf("number of human players must be between 0 and %d, got %d", config.MAX_PLAYERS, *humans)
	}
	neat.InitLogger("debug")
	var bots []server.PlayerIO
	for range config.MAX_PLAYERS - *humans {
		bot, err := loadNEATPlayer(*genome)
		if err != nil {
			return err
		}
		bots = append(bots, bot)
	}

	io := server.NewConsoleServer(*humans, bots)
	logger := monopoly.ConsoleLogger{}
	logger.Init()
	game := monopoly.NewGame(context.Background(), io, &logger, 0)
	game.Start()
	return nil
}

func askHumanPlayers() int {
	for {
		var n int
		fmt.Printf("Enter number of human players (0-%d): ", config.MAX_PLAYERS)
		if _, err := fmt.Scan(&n); err == nil && n >= 0 && n <= config.MAX_PLAYERS {
			return n
		}
		fmt.Printf("Number must be between 0 and %d\n", config.MAX_PLAYERS)
	}
}

func runClient(args []string) error {
	fs := newFlagSet("client", "Connects to the game server on localhost:12345 (started with: go run . play)")
	_ = fs.Parse(args)
	consoleCLI.StartClient()
	return nil
}

func runTrain(args []string) error {
	fs := newFlagSet("train", "Runs a NEAT experiment from scratch (or from a starting genome).\n"+
		"The seed determines the whole run (group draws, dice, mutations, crossovers): the same seed and options\n"+
		"reproduce the same result. Without -seed a random seed is drawn. The seed is printed and saved\n"+
		"to run_info.txt in the output directory.")
	seed := fs.Int64("seed", 0, "seed of the training run (0: random)")
	options := fs.String("options", "neat_options.yaml", "NEAT options file")
	out := fs.String("out", "", "output directory (default: output/seed<seed>)")
	genome := fs.String("genome", "", "starting genome file (default: built from INPUT_COUNT/OUTPUT_COUNT)")
	allToAll := fs.Bool("all-to-all", true, "built starting genome: connect every input and the bias to every output")
	_ = fs.Parse(args)

	if *seed == 0 {
		*seed = randomSeed()
	}
	if *out == "" {
		*out = filepath.Join("output", fmt.Sprintf("seed%d", *seed))
	}
	fmt.Printf("Training: seed=%d options=%s output=%s\n", *seed, *options, *out)
	neatnetwork.TrainNetwork(*seed, *options, *out, *allToAll, *genome)
	return nil
}

func runMatch(args []string) error {
	fs := newFlagSet("match", "Bots play GAMES_PER_EPOCH games in random groups of 4, like one training epoch.")
	genomes := fs.String("genomes", defaultGenomes, "comma-separated genome files (a multiple of 4)")
	seed := fs.Int64("seed", 0, "seed of the group draws and games (0: random)")
	options := fs.String("options", "neat_options.yaml", "NEAT options file")
	out := fs.String("out", "experiment", "output directory for game logs")
	_ = fs.Parse(args)

	players, err := loadNEATPlayers(*genomes)
	if err != nil {
		return err
	}
	if *seed == 0 {
		*seed = randomSeed()
	}
	neatOptions, err := neat.ReadNeatOptionsFromFile(*options)
	if err != nil {
		return fmt.Errorf("failed to load NEAT options: %w", err)
	}
	e := neatnetwork.NewMonopolyEvaluator(*out, config.GROUP_SIZE, rand.New(rand.NewSource(*seed)))
	if err := e.PlayRound(neat.NewContext(context.Background(), neatOptions), players, -1); err != nil {
		return err
	}

	fmt.Printf("Seed: %d, games per bot: %d\n", *seed, config.GAMES_PER_EPOCH)
	for i, path := range splitList(*genomes) {
		p := players[i]
		fmt.Printf("%-30s AvgScore=%6.1f Wins=%d Draws=%d\n", path,
			float64(p.GetScore())/config.GAMES_PER_EPOCH, p.GetWins(), p.GetDraws())
	}
	return nil
}

func runDebug(args []string) error {
	fs := newFlagSet("debug", "One game of 4 bots. The game log and the decision logs of selected bots go to -log-dir.")
	genomes := fs.String("genomes", defaultGenomes, "comma-separated genome files of the 4 bots")
	logDir := fs.String("log-dir", "debug_logs", "directory for the game log and the decision logs")
	logBots := fs.String("log-bots", "0", "comma-separated indexes (0-3) of the bots whose decisions are logged; empty: none")
	_ = fs.Parse(args)

	players, err := loadNEATPlayers(*genomes)
	if err != nil {
		return err
	}
	if len(players) != config.GROUP_SIZE {
		return fmt.Errorf("debug game needs %d genomes, got %d", config.GROUP_SIZE, len(players))
	}
	for _, s := range splitList(*logBots) {
		i, err := strconv.Atoi(s)
		if err != nil || i < 0 || i >= len(players) {
			return fmt.Errorf("invalid bot index %q in -log-bots", s)
		}
		bot := players[i].(*neatnetwork.NEATMonopolyPlayer)
		decisionLog, err := neatnetwork.NewBotDecisionLogger(*logDir, bot.GetName())
		if err != nil {
			return fmt.Errorf("failed to create decision logger: %w", err)
		}
		defer decisionLog.Close()
		bot.SetDecisionLogger(decisionLog)
	}
	return neatnetwork.RunDebugGame(context.Background(), players, *logDir)
}

// randomSeed draws a seed for a run started without -seed. It is never 0, which means "draw a seed".
func randomSeed() int64 {
	return randv2.Int64N(math.MaxInt64) + 1
}

func splitList(s string) []string {
	var items []string
	for _, item := range strings.Split(s, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func loadNEATPlayers(genomes string) ([]neatnetwork.MonopolyPlayer, error) {
	var players []neatnetwork.MonopolyPlayer
	for _, path := range splitList(genomes) {
		player, err := loadNEATPlayer(path)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}
	return players, nil
}

func loadNEATPlayer(filePath string) (*neatnetwork.NEATMonopolyPlayer, error) {
	genomeReader, err := genetics.NewGenomeReaderFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open genome %q: %w", filePath, err)
	}
	genome, err := genomeReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read genome %q: %w", filePath, err)
	}
	organism, err := genetics.NewOrganism(0.0, genome, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create organism from genome %q: %w", filePath, err)
	}
	bot, err := neatnetwork.NewNEATMonopolyPlayer(organism)
	if err != nil {
		return nil, fmt.Errorf("failed to create NEAT player from genome %q: %w", filePath, err)
	}
	return bot, nil
}
