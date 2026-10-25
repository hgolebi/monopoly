package neatnetwork

import (
	"errors"
	"fmt"
	"math"

	"github.com/yaricom/goNEAT/v4/neat/genetics"
	neatmath "github.com/yaricom/goNEAT/v4/neat/math"
	"github.com/yaricom/goNEAT/v4/neat/network"
)

type fastNetwork struct {
	*fastStructure
	values []float64 // per node: sensors first (genome order), then the evaluated neurons in evaluation order
}

type fastStructure struct {
	inputCount int   // input neurons, loaded from the sensors in genome order
	biasIndex  []int // value indexes of bias neurons (always 1)
	neurons    []fastNeuron
	links      []fastLink // incoming links of all neurons, grouped by neuron
	outputs    []int      // value index of each output neuron, in genome order (as Network.Outputs)
	valueCount int
}

type fastNeuron struct {
	value      int // index in values
	firstLink  int // links[firstLink:endLink] are the incoming links, in genome order
	endLink    int
	activation neatmath.NodeActivationType
	params     []float64
}

type fastLink struct {
	source int // index in values
	weight float64
}

var errRecurrentNetwork = errors.New("network has recurrent links")

// compileFastNetwork compiles the phenotype of genome. It returns errRecurrentNetwork when the network has cycles
// or recurrent links, and an error for unsupported genomes (control genes, unknown activation functions).
func compileFastNetwork(genome *genetics.Genome) (*fastNetwork, error) {
	if len(genome.ControlGenes) > 0 {
		return nil, errors.New("networks with control genes are not supported")
	}

	// incoming links per node, in gene order (like Genesis)
	incoming := make(map[int][]*network.Link)
	for _, gene := range genome.Genes {
		if !gene.IsEnabled {
			continue
		}
		if gene.Link.IsRecurrent || gene.Link.IsTimeDelayed {
			return nil, errRecurrentNetwork
		}
		incoming[gene.Link.OutNode.Id] = append(incoming[gene.Link.OutNode.Id], gene.Link)
	}

	s := &fastStructure{}
	index := make(map[int]int) // node id → value index
	var neurons []*network.NNode
	for _, node := range genome.Nodes {
		switch {
		case node.NeuronType == network.InputNeuron:
			index[node.Id] = len(index)
			s.inputCount++
		case node.NeuronType == network.BiasNeuron:
			s.biasIndex = append(s.biasIndex, len(index))
			index[node.Id] = len(index)
		default:
			neurons = append(neurons, node)
		}
	}
	// the inputs are loaded by position, so they must come before the bias neurons, as in our genomes
	for _, b := range s.biasIndex {
		if b < s.inputCount {
			return nil, errors.New("bias neuron placed before input neurons")
		}
	}

	order, err := topologicalOrder(neurons, incoming)
	if err != nil {
		return nil, err
	}
	needed := neuronsLeadingToOutputs(neurons, incoming)

	active := make(map[int]bool) // goNEAT activates a neuron only when some input is a sensor or an active neuron
	for _, node := range order {
		links := incoming[node.Id]
		for _, link := range links {
			if link.InNode.IsSensor() || active[link.InNode.Id] {
				active[node.Id] = true
				break
			}
		}
		if !active[node.Id] || !needed[node.Id] {
			continue // stays at 0, like a neuron goNEAT never activates
		}
		if _, err := neatmath.NodeActivators.ActivateByType(0, node.Params, node.ActivationType); err != nil {
			return nil, fmt.Errorf("neuron %d: %w", node.Id, err)
		}
		index[node.Id] = len(index)
		n := fastNeuron{value: index[node.Id], firstLink: len(s.links), activation: node.ActivationType, params: node.Params}
		for _, link := range links {
			source, ok := index[link.InNode.Id]
			if !ok {
				continue // inactive source: contributes weight * 0
			}
			s.links = append(s.links, fastLink{source: source, weight: link.ConnectionWeight})
		}
		n.endLink = len(s.links)
		s.neurons = append(s.neurons, n)
	}
	s.valueCount = len(index) + 1 // the last value is the constant 0 of inactive outputs

	for _, node := range genome.Nodes {
		if node.NeuronType == network.OutputNeuron {
			if i, ok := index[node.Id]; ok {
				s.outputs = append(s.outputs, i)
			} else {
				s.outputs = append(s.outputs, s.valueCount-1)
			}
		}
	}
	if len(s.outputs) == 0 {
		return nil, errors.New("network without outputs")
	}
	return s.newInstance(), nil
}

// topologicalOrder sorts the neurons so that every neuron comes after the neurons it gets input from.
// Neurons that are ready at the same time keep the genome order.
func topologicalOrder(neurons []*network.NNode, incoming map[int][]*network.Link) ([]*network.NNode, error) {
	waiting := make(map[int]int, len(neurons)) // number of neuron inputs not placed yet
	outgoing := make(map[int][]*network.NNode)
	for _, node := range neurons {
		for _, link := range incoming[node.Id] {
			if !link.InNode.IsSensor() {
				waiting[node.Id]++
				outgoing[link.InNode.Id] = append(outgoing[link.InNode.Id], node)
			}
		}
	}
	order := make([]*network.NNode, 0, len(neurons))
	for _, node := range neurons {
		if waiting[node.Id] == 0 {
			order = append(order, node)
		}
	}
	for i := 0; i < len(order); i++ {
		for _, next := range outgoing[order[i].Id] {
			waiting[next.Id]--
			if waiting[next.Id] == 0 {
				order = append(order, next)
			}
		}
	}
	if len(order) != len(neurons) {
		return nil, errRecurrentNetwork
	}
	return order, nil
}

// neuronsLeadingToOutputs returns the neurons whose value can reach an output; the others need not be computed.
func neuronsLeadingToOutputs(neurons []*network.NNode, incoming map[int][]*network.Link) map[int]bool {
	needed := make(map[int]bool)
	var stack []int
	for _, node := range neurons {
		if node.NeuronType == network.OutputNeuron {
			needed[node.Id] = true
			stack = append(stack, node.Id)
		}
	}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, link := range incoming[id] {
			if in := link.InNode; !in.IsSensor() && !needed[in.Id] {
				needed[in.Id] = true
				stack = append(stack, in.Id)
			}
		}
	}
	return needed
}

// newInstance returns a network with its own values, sharing the compiled structure.
func (s *fastStructure) newInstance() *fastNetwork {
	values := make([]float64, s.valueCount)
	for _, b := range s.biasIndex {
		values[b] = 1.0
	}
	return &fastNetwork{fastStructure: s, values: values}
}

// activate computes the outputs for inputs (one value per input neuron) and returns them in a new slice.
func (f *fastNetwork) activate(inputs []float64) ([]float64, error) {
	if len(inputs) != f.inputCount {
		return nil, fmt.Errorf("expected %d inputs, got %d", f.inputCount, len(inputs))
	}
	copy(f.values, inputs)
	for i := range f.neurons {
		n := &f.neurons[i]
		sum := 0.0
		for _, link := range f.links[n.firstLink:n.endLink] {
			sum += link.weight * f.values[link.source]
		}
		if n.activation == neatmath.SigmoidSteepenedActivation {
			// the same formula as goNEAT's steepenedSigmoid, without the lookup of the function by type
			f.values[n.value] = 1.0 / (1.0 + math.Exp(-4.924273*sum))
		} else {
			out, err := neatmath.NodeActivators.ActivateByType(sum, n.params, n.activation)
			if err != nil {
				return nil, err
			}
			f.values[n.value] = out
		}
	}
	outputs := make([]float64, len(f.outputs))
	for i, idx := range f.outputs {
		outputs[i] = f.values[idx]
	}
	return outputs, nil
}
