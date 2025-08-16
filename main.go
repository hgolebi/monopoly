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
	logger.Init()

	if *cliMode {
		consoleCLI.StartClient()
	} else {
		game := monopoly.NewGame(&io, &logger, 0)
		game.Start()
	}
}
