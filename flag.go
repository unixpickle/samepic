package samepic

import (
	"errors"
	"flag"
)

// Flags is used to create a Samer from command-line
// arguments.
type Flags struct {
	Name       string
	NeuralPath string
	Threshold  float64
}

// AddToSet adds the struct fields of f as arguments to
// the flag set.
func (f *Flags) AddToSet(set *flag.FlagSet) {
	set.StringVar(&f.Name, "samer", "", "name of samer "+
		"(avghash, colorprof, squashcomp, or neuralnet)")
	set.StringVar(&f.NeuralPath, "netpath", "", "path to neural network "+
		"(for neuralnet samer)")
	set.Float64Var(&f.Threshold, "threshold", 0, "threshold "+
		"(applicable for most samers)")
}

// Samer creates a samer from the parsed flags.
func (f *Flags) Samer() (Samer, error) {
	if f.Name == "" {
		return nil, errors.New("missing -samer flag")
	}

	switch f.Name {
	case "avghash":
		return &AverageHash{Threshold: f.Threshold}, nil
	case "colorprof":
		return &ColorProf{Threshold: f.Threshold}, nil
	case "squashcomp":
		return &SquashComp{Threshold: f.Threshold}, nil
	case "neuralnet":
		if f.NeuralPath == "" {
			return nil, errors.New("missing -netpath flag")
		}
		res, err := LoadNeuralSamer(f.NeuralPath)
		if err != nil {
			return nil, err
		}
		res.Cutoff = f.Threshold
		return res, nil
	default:
		return nil, errors.New("unknown samer: " + f.Name)
	}
}

// BatchSamer is like Samer but it produces a BatchSamer.
func (f *Flags) BatchSamer() (BatchSamer, error) {
	samer, err := f.Samer()
	if err != nil {
		return nil, err
	}
	if bs, ok := samer.(BatchSamer); ok {
		return bs, nil
	} else {
		return nil, errors.New("samer cannot be used for batches")
	}
}
