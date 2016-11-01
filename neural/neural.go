package neural

import (
	"github.com/fxsjy/gonn/gonn"
)


// These are just guesses at this point
const OUTPUT_NODES int = 1
const DOWNSAMPLING int = 5
const RATE1 float64 = 0.25
const RATE2 float64 = 0.1

// Will need to corolate to RAM at some point
const MAX_DATASET_SIZE int = 100000
const TRAIN_CYCLES int = 200

type Neural struct {
	nn  *gonn.NeuralNetwork
	input [][]float64
	expected [][]float64
}

func New(sampleSize int) *Neural {
	return &Neural{nn: gonn.NewNetwork(sampleSize, sampleSize / DOWNSAMPLING, OUTPUT_NODES, false, RATE1, RATE2)}
}

func (n *Neural) AddDataset(input, expected [][]float64) {
	// Should randomly discard old data to make room for new
	// Currently just discards all data
	if (len(n.input) + len(input) > MAX_DATASET_SIZE) {
		n.input = input
		n.expected = expected
	} else {
		n.input = append(n.input, input...)
		n.expected = append(n.expected, expected...)
	}
	n.train()
}

func (n *Neural) train() {
	n.nn.Train(n.input, n.expected, TRAIN_CYCLES)
}

