package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/unixpickle/samepic"
	"github.com/unixpickle/serializer"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "sample_dir network_file")
		os.Exit(1)
	}

	sampleDir, err := samepic.NewDirSamples(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read samples:", err)
		os.Exit(1)
	}

	network := samepic.NewNeuralSamer()
	networkData, err := ioutil.ReadFile(os.Args[2])
	if err == nil {
		net, err := serializer.DeserializeWithType(networkData)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to deserialize existing network:", err)
			os.Exit(1)
		}
		var ok bool
		network, ok = net.(*samepic.NeuralSamer)
		if !ok {
			fmt.Fprintf(os.Stderr, "Unexpected type: %T\n", net)
			os.Exit(1)
		}
		log.Println("Loaded network from file.")
	} else {
		log.Println("Created new network.")
	}

	log.Println("Training...")
	network.Train(sampleDir, samepic.DefaultManipulator)
}
