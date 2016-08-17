// Command same_gen generates a manipulation of an
// image to demonstrate samepic.DefaultManipulator.
package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"time"

	"github.com/unixpickle/samepic"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage", os.Args[0], "original_image manipulated.png")
		os.Exit(1)
	}
	inFile, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read input:", err)
		os.Exit(1)
	}
	defer inFile.Close()
	img, _, err := image.Decode(inFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to decode input:", err)
		os.Exit(1)
	}
	manip := samepic.DefaultManipulator.Manipulate(img)
	outFile, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create output:", err)
		os.Exit(1)
	}
	defer outFile.Close()
	if err := png.Encode(outFile, manip); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to encode output:", err)
		os.Exit(1)
	}
}
