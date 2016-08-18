package samepic

import (
	"image"

	"github.com/unixpickle/num-analysis/linalg"
)

const (
	DefaultColorProfBinCount  = 8
	DefaultColorProfThreshold = 0.97
)

// ColorProf compares two images by generating color
// histograms for each image and comparing the resulting
// histograms using correlation.
type ColorProf struct {
	// BinCount is the number of bins in each of the
	// three histograms (R, G, and B).
	//
	// If this is 0, DefaultColorProfBinCount is used.
	BinCount int

	// Threshold is the minimum correlation between two
	// color histograms for them to be considered the
	// same.
	//
	// If this is 0, DefaultColorProfThreshold is used.
	Threshold float64
}

// Same decides if two images are the same by comparing
// their color histograms.
func (c *ColorProf) Same(img1, img2 image.Image) bool {
	hist1 := c.Histograms(img1)
	hist2 := c.Histograms(img2)

	joinedHist1 := joinVecs(hist1[:])
	joinedHist2 := joinVecs(hist2[:])

	correlation := joinedHist1.Dot(joinedHist2) / (joinedHist1.Mag() * joinedHist2.Mag())
	if c.Threshold == 0 {
		return correlation >= DefaultColorProfThreshold
	} else {
		return correlation >= c.Threshold
	}
}

// Histograms generates the R, G, and B histograms for
// the given image.
func (c *ColorProf) Histograms(img image.Image) [3]linalg.Vector {
	var res [3]linalg.Vector
	for i := 0; i < 3; i++ {
		binCount := c.BinCount
		if binCount == 0 {
			binCount = DefaultColorProfBinCount
		}
		res[i] = make(linalg.Vector, binCount)
	}
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			res[0][c.binIdx(r)]++
			res[1][c.binIdx(g)]++
			res[2][c.binIdx(b)]++
		}
	}
	return res
}

func (c *ColorProf) binIdx(component uint32) int {
	binCount := c.BinCount
	if binCount == 0 {
		binCount = DefaultColorProfBinCount
	}
	idx := int(float64(binCount) * float64(component) / 0xffff)
	if idx >= binCount {
		idx = binCount - 1
	}
	return idx
}

func joinVecs(vecs []linalg.Vector) linalg.Vector {
	var res linalg.Vector
	for _, v := range vecs {
		res = append(res, v...)
	}
	return res
}
