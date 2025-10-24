package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yaricom/goNEAT/v4/neat/genetics"
	"github.com/yaricom/goNEAT/v4/neat/network/formats"
)

func main() {
	// Retrieve the genome file path from command line arguments
	filePath := "C:\\Users\\Hubert\\Desktop\\genomes\\gen_2999"
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run graph.go <genome_file_path>")
	} else {
		filePath = os.Args[1]
	}
	useDotFormat := false
	if len(os.Args) > 2 && os.Args[2] == "--dot" {
		useDotFormat = true
	}
	genomeReader, err := genetics.NewGenomeReaderFromFile(filePath)
	if err != nil {
		log.Fatal("Failed to create genome reader:", err)
	}
	startGenome, err := genomeReader.Read()
	if err != nil {
		log.Fatal("Failed to read genome:", err)
	}
	net, err := startGenome.Genesis(1)
	if err != nil {
		log.Fatal("Failed to create network from genome:", err)
	}

	if useDotFormat {
		dotFile, err := os.OpenFile("graph.dot", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal("Failed to open dot file:", err)
		}
		defer dotFile.Close()
		err = formats.WriteDOT(dotFile, net)
		if err != nil {
			log.Fatal("Failed to write file:", err)
		}
	} else {
		jsonFile, err := os.OpenFile("graph.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal("Failed to open json file:", err)
		}
		defer jsonFile.Close()
		err = formats.WriteCytoscapeJSON(jsonFile, net)
		if err != nil {
			log.Fatal("Failed to write file:", err)
		}
	}
	fmt.Printf("Number of nodes: %d\n", len(startGenome.Nodes))
	fmt.Printf("Number of genes: %d\n", len(startGenome.Genes))
	fmt.Printf("Number of control genes: %d\n", len(startGenome.ControlGenes))
}
