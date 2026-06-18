package neatnetwork

import (
	"log"

	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
	neatmath "github.com/yaricom/goNEAT/v4/neat/math"
	"github.com/yaricom/goNEAT/v4/neat/network"
)

func loadOrBuildGenome(genomeFile string, allToAll bool) *genetics.Genome {
	if genomeFile == "" {
		return NewStartGenome(1, allToAll)
	}
	reader, err := genetics.NewGenomeReaderFromFile(genomeFile)
	if err != nil {
		log.Fatalf("Failed to open genome file %q: %v", genomeFile, err)
	}
	g, err := reader.Read()
	if err != nil {
		log.Fatalf("Failed to read genome from %q: %v", genomeFile, err)
	}
	return g
}

// NewStartGenome builds a minimal starting genome derived from INPUT_COUNT and OUTPUT_COUNT.
// If allToAll is true every input and the bias node are connected to every output (weight 0).
// If allToAll is false the genome has no genes — NEAT will grow connections through mutation.
func NewStartGenome(id int, allToAll bool) *genetics.Genome {
	trait := neat.NewTrait()
	trait.Id = 1

	nodes := buildNodes(trait)
	var genes []*genetics.Gene
	if allToAll {
		genes = buildAllToAllGenes(trait, nodes)
	}

	return genetics.NewGenome(id, []*neat.Trait{trait}, nodes, genes)
}

func buildNodes(trait *neat.Trait) []*network.NNode {
	inputCount := int(INPUT_COUNT)
	outputCount := int(OUTPUT_COUNT)
	nodes := make([]*network.NNode, 0, inputCount+1+outputCount)

	for i := 0; i < inputCount; i++ {
		n := network.NewNNode(i+1, network.InputNeuron)
		n.ActivationType = neatmath.NullActivation
		n.Trait = trait
		nodes = append(nodes, n)
	}

	biasNode := network.NewNNode(inputCount+1, network.BiasNeuron)
	biasNode.ActivationType = neatmath.NullActivation
	biasNode.Trait = trait
	nodes = append(nodes, biasNode)

	for i := 0; i < outputCount; i++ {
		n := network.NewNNode(inputCount+2+i, network.OutputNeuron)
		n.ActivationType = neatmath.SigmoidSteepenedActivation
		n.Trait = trait
		nodes = append(nodes, n)
	}

	return nodes
}

func buildAllToAllGenes(trait *neat.Trait, nodes []*network.NNode) []*genetics.Gene {
	inputCount := int(INPUT_COUNT)
	outputCount := int(OUTPUT_COUNT)
	sourcesCount := inputCount + 1 // inputs + bias
	genes := make([]*genetics.Gene, 0, sourcesCount*outputCount)

	innovNum := int64(1)
	for si := 0; si < sourcesCount; si++ {
		for oi := 0; oi < outputCount; oi++ {
			outNode := nodes[sourcesCount+oi]
			gene := genetics.NewGeneWithTrait(trait, 0.0, nodes[si], outNode, false, innovNum, 0.0)
			genes = append(genes, gene)
			innovNum++
		}
	}
	return genes
}
