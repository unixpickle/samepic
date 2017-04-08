package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/samepic"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	samerFlags := &samepic.Flags{}
	samerFlags.AddToSet(fs)

	var count int
	var sampleDir string
	fs.IntVar(&count, "count", 100, "number of samples to try")
	fs.StringVar(&sampleDir, "dir", "", "directory of samples")

	fs.Parse(os.Args[1:])

	if sampleDir == "" {
		essentials.Die("Required flag: -dir. See -help.")
	}

	samer, err := samerFlags.Samer()
	if err != nil {
		essentials.Die(err)
	}

	samples, err := samepic.NewDirSamples(sampleDir)
	if err != nil {
		essentials.Die(err)
	}

	fmt.Println("Rating...")
	pos, neg, err := samepic.Rate(samer, samples, samepic.DefaultManipulator, count)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to rate:", err)
		os.Exit(1)
	}
	fmt.Println("Positive rating:", pos)
	fmt.Println("Negative rating:", neg)
}
