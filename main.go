package main

import (
	"flag"
	"monopoly/pkg/consoleCLI"
	"monopoly/pkg/monopoly"
)

func main() {
	cliMode := flag.Bool("cli", false, "run in CLI client mode")
	flag.Parse()
	io := monopoly.ConsoleServer{}
	logger := monopoly.ConsoleLogger{}
	numberOfPlayers := 4

	if *cliMode {
		consoleCLI.StartClient()
	} else {
		game := monopoly.NewGame(numberOfPlayers, &io, &logger)
		game.Start()
	}
}
