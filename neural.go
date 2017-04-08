package samepic

import (
	"image"

	"github.com/unixpickle/anydiff"
	"github.com/unixpickle/anynet"
	"github.com/unixpickle/anyvec"
	"github.com/unixpickle/anyvec/anyvec32"
	"github.com/unixpickle/imagenet"
	"github.com/unixpickle/serializer"
)

const (
	DefaultNeuralSamerCutoff = 0.4
)

// NeuralSamer is a Samer which uses a pre-trained neural
// network.
//
// In particular, a NeuralSamer uses classifiers from
// https://github.com/unixpickle/imagenet.
type NeuralSamer struct {
	// Net takes image tensors as input and produces feature
	// vectors.
	Net anynet.Net

	// Cutoff is the maximum MSE in feature vectors for which
	// the images are considered the same.
	//
	// If this is 0, DefaultNeuralSamerCutoff is used.
	Cutoff float64
}

// LoadNeuralSamer loads a NeuralSamer from an imagenet
// classifier path.
func LoadNeuralSamer(path string) (*NeuralSamer, error) {
	var cl *imagenet.Classifier
	if err := serializer.LoadAny(path, &cl); err != nil {
		return nil, err
	}
	return &NeuralSamer{
		Net: cl.Net[:len(cl.Net)-2],
	}, nil
}

// Same uses the neural network to predict whether the
// two images are derived from the same source image.
func (n *NeuralSamer) Same(img1, img2 image.Image) bool {
	in1 := imagenet.ImageToTensor(img1)
	in2 := imagenet.ImageToTensor(img2)
	joinedIn := anydiff.NewConst(in1.Creator().Concat(in1, in2))
	outs := n.Net.Apply(joinedIn, 2).Output()
	out1 := outs.Slice(0, outs.Len()/2)
	out2 := outs.Slice(outs.Len()/2, outs.Len())
	out1.Sub(out2)
	mse := out1.Dot(out1).(float32) / float32(out1.Len())
	return float64(mse) < n.cutoff()
}

// SameBatch finds pairs of near duplicates.
func (n *NeuralSamer) SameBatch(images <-chan *IDImage) <-chan *Pair {
	res := make(chan *Pair, 1)
	go func() {
		defer close(res)
		c := anyvec32.CurrentCreator()
		ids := []interface{}{}
		features := c.MakeVector(0)
		for image := range images {
			in := imagenet.ImageToTensor(image.Image)
			out := n.Net.Apply(anydiff.NewConst(in), 1).Output().Copy()
			diffs := features.Copy()
			diffs.Scale(c.MakeNumeric(-1))
			anyvec.AddRepeated(diffs, out)
			anyvec.Pow(diffs, c.MakeNumeric(2))
			sqErrors := anyvec.SumCols(diffs, len(ids)).Data().([]float32)
			for i, sqError := range sqErrors {
				mse := float64(sqError) / float64(out.Len())
				if mse < n.cutoff() {
					res <- &Pair{ids[i], image.ID}
				}
			}
			ids = append(ids, image.ID)
			features = c.Concat(features, out)
		}
	}()
	return res
}

func (n *NeuralSamer) cutoff() float64 {
	if n.Cutoff == 0 {
		return DefaultNeuralSamerCutoff
	}
	return n.Cutoff
}
