package main

import (
	"flag"
	"monopoly/pkg/consoleCLI"
	"monopoly/pkg/monopoly"
)

func startGame() {
	game := monopoly.Game{}
	game.Start()
}

func main() {
	cliMode := flag.Bool("cli", false, "run in CLI client mode")
	flag.Parse()

	if *cliMode {
		consoleCLI.StartClient()
	} else {
		startGame()
	}
}
