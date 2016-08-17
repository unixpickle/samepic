package samepic

import (
	"errors"
	"image"
	"log"
	"math/rand"
	"sync"

	"github.com/nfnt/resize"
	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/neuralnet"
)

const (
	neuralSamerDefaultInSize = 300
)

func init() {
	var n NeuralSamer
	serializer.RegisterTypedDeserializer(n.SerializerType(), DeserializeNeuralSamer)
}

// NeuralSamer uses a convolutional neural network to
// determine if two images are of the same thing.
type NeuralSamer struct {
	inSize  int
	network neuralnet.Network
}

// DeserializeNeuralSamer deserializes a NeuralSamer.
func DeserializeNeuralSamer(d []byte) (*NeuralSamer, error) {
	slice, err := serializer.DeserializeSlice(d)
	if err != nil {
		return nil, err
	}
	if len(slice) != 2 {
		return nil, errors.New("invalid NeuralSamer slice")
	}
	inSize, ok1 := slice[0].(serializer.Int)
	network, ok2 := slice[1].(neuralnet.Network)
	if !ok1 || !ok2 {
		return nil, errors.New("invalid NeuralSamer slice")
	}
	return &NeuralSamer{
		inSize:  int(inSize),
		network: network,
	}, nil
}

// NewNeuralSamer makes a randomly initialized NeuralSamer.
func NewNeuralSamer() *NeuralSamer {
	convLayer1 := &neuralnet.ConvLayer{
		FilterCount:  20,
		FilterWidth:  3,
		FilterHeight: 3,
		Stride:       1,

		InputWidth:  neuralSamerDefaultInSize,
		InputHeight: neuralSamerDefaultInSize,
		InputDepth:  6,
	}
	poolingLayer1 := &neuralnet.MaxPoolingLayer{
		XSpan:       3,
		YSpan:       3,
		InputWidth:  convLayer1.OutputWidth(),
		InputHeight: convLayer1.OutputHeight(),
		InputDepth:  convLayer1.FilterCount,
	}
	denseLayer1 := &neuralnet.DenseLayer{
		InputCount: poolingLayer1.OutputWidth() * poolingLayer1.OutputHeight() *
			convLayer1.FilterCount,
		OutputCount: 100,
	}
	denseLayer2 := &neuralnet.DenseLayer{
		InputCount:  denseLayer1.OutputCount,
		OutputCount: 1,
	}
	network := neuralnet.Network{
		convLayer1,
		&neuralnet.HyperbolicTangent{},
		poolingLayer1,
		denseLayer1,
		&neuralnet.HyperbolicTangent{},
		denseLayer2,
	}
	network.Randomize()
	return &NeuralSamer{
		inSize:  neuralSamerDefaultInSize,
		network: network,
	}
}

// Same uses the neural network to predict whether the
// two images are derived from the same source image.
func (n *NeuralSamer) Same(img1, img2 image.Image) bool {
	pair1 := n.pairToTensor(img1, img2)
	pair2 := n.pairToTensor(img2, img1)
	out1 := n.network.Apply(&autofunc.Variable{Vector: pair1.Data}).Output()
	out2 := n.network.Apply(&autofunc.Variable{Vector: pair2.Data}).Output()
	return (out1[0] + out2[0]) > 0
}

// Train trains the neural network using a source of
// samples and a means of manipulating images to create
// positive samples.
// It uses sgd.SGDInteractive, so it stops when the user
// sends a kill signal.
func (n *NeuralSamer) Train(samples Samples, manip Manipulator) {
	batchGrad := &neuralnet.BatchRGradienter{
		Learner:  n.network.BatchLearner(),
		CostFunc: &neuralnet.SigmoidCECost{},

		MaxGoroutines: 3,
		MaxBatchSize:  5,
	}
	batchSize := batchGrad.MaxBatchSize * batchGrad.MaxGoroutines
	gradienter := &sgd.Adam{
		Gradienter: batchGrad,
	}
	sampleSet := make(sgd.SliceSampleSet, batchSize)
	sgd.SGDInteractive(gradienter, sampleSet, 0.001, batchSize, func() bool {
		n.createSampleSet(sampleSet, samples, manip)
		cost := n.totalCost(sampleSet, batchGrad.MaxGoroutines)
		log.Println("minibatch cost", cost)
		return true
	})
}

// SerializerType returns the unique ID used to serialize
// a NeuralSamer with the serializer package.
func (n *NeuralSamer) SerializerType() string {
	return "github.com/unixpickle/samepic.NeuralSamer"
}

// Serialize serializes the NeuralSamer.
func (n *NeuralSamer) Serialize() ([]byte, error) {
	slice := []serializer.Serializer{
		serializer.Int(n.inSize),
		n.network,
	}
	return serializer.SerializeSlice(slice)
}

func (n *NeuralSamer) createSampleSet(set sgd.SliceSampleSet, s Samples, m Manipulator) error {
	for i := range set {
		if rand.Intn(2) == 0 {
			img1, img2, err := s.RandomPair()
			if err != nil {
				return err
			}
			set[i] = neuralnet.VectorSample{
				Input:  n.pairToTensor(img1, img2).Data,
				Output: []float64{0},
			}
		} else {
			img, err := s.Random()
			if err != nil {
				return err
			}
			img1 := m.Manipulate(img)
			img2 := m.Manipulate(img)
			set[i] = neuralnet.VectorSample{
				Input:  n.pairToTensor(img1, img2).Data,
				Output: []float64{1},
			}
		}
	}
	return nil
}

func (n *NeuralSamer) totalCost(set sgd.SliceSampleSet, maxGos int) float64 {
	var total float64
	var totalLock sync.Mutex

	sampleChan := make(chan sgd.SampleSet, set.Len())
	for i := 0; i < set.Len(); i++ {
		sampleChan <- set.Subset(i, i+1)
	}
	close(sampleChan)

	var wg sync.WaitGroup
	for i := 0; i < maxGos; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for set := range sampleChan {
				cost := neuralnet.TotalCost(&neuralnet.SigmoidCECost{}, n.network, set)
				totalLock.Lock()
				total += cost
				totalLock.Unlock()
			}
		}()
	}

	wg.Wait()
	return total
}

func (n *NeuralSamer) pairToTensor(img1, img2 image.Image) *neuralnet.Tensor3 {
	t1 := n.imageToTensor(img1)
	t2 := n.imageToTensor(img2)
	res := neuralnet.NewTensor3(n.inSize, n.inSize, 6)
	for y := 0; y < t1.Height; y++ {
		for x := 0; x < t1.Width; x++ {
			for z := 0; z < 3; z++ {
				res.Set(x, y, z, t1.Get(x, y, z))
				res.Set(x, y, z+3, t2.Get(x, y, z))
			}
		}
	}
	return res
}

func (n *NeuralSamer) imageToTensor(img image.Image) *neuralnet.Tensor3 {
	if img.Bounds().Dx() > img.Bounds().Dy() {
		img = resize.Resize(uint(n.inSize), 0, img, resize.Bilinear)
	} else {
		img = resize.Resize(0, uint(n.inSize), img, resize.Bilinear)
	}
	res := neuralnet.NewTensor3(n.inSize, n.inSize, 3)
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			res.Set(x, y, 0, float64(r)/0xffff)
			res.Set(x, y, 1, float64(g)/0xffff)
			res.Set(x, y, 2, float64(b)/0xffff)
		}
	}
	return res
}
